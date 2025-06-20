package usecases

import (
	"log"
	"time"

	"github.com/aryuuu/mines-party-server/configs"
	"github.com/aryuuu/mines-party-server/events"
	"github.com/aryuuu/mines-party-server/minesweeper"
	"github.com/gorilla/websocket"
)

const (
	scoreUpdateInterval = 3 * time.Second
)

type connection struct {
	ID    string
	Queue chan interface{}
}

type gameUsecase struct {
	ConnectionRooms   map[string]map[*websocket.Conn]*connection
	GameRooms         map[string]*minesweeper.GameRoom
	StopScoreCronChan map[string]chan bool
	SwitchQueue       chan *events.SocketEvent
}

type GameUsecase interface {
	Connect(conn *websocket.Conn, roomID string)
	RunSwitch()
}

func NewConnection(ID string) *connection {
	return &connection{
		ID:    ID,
		Queue: make(chan interface{}, 256),
	}
}

func NewGameUsecase() GameUsecase {
	return &gameUsecase{
		ConnectionRooms:   make(map[string]map[*websocket.Conn]*connection),
		GameRooms:         make(map[string]*minesweeper.GameRoom),
		SwitchQueue:       make(chan *events.SocketEvent, 256),
		StopScoreCronChan: make(map[string]chan bool),
	}
}

func (u *gameUsecase) Connect(conn *websocket.Conn, roomID string) {
	for {
		var clientEvent events.ClientEvent
		err := conn.ReadJSON(&clientEvent)

		if err != nil {
			log.Print(err)
			if websocket.IsUnexpectedCloseError(err, websocket.CloseNormalClosure) {
				log.Print("IsUnexpectedCloseError()", err)
				u.kickPlayer(conn, roomID, clientEvent)
			} else {
				log.Printf("expected close error: %v", err)
			}
			return
		}

		switch clientEvent.EventType {
		case events.CreateRoomEvent:
			u.createRoom(conn, roomID, clientEvent)
		case events.JoinRoomEvent:
			u.joinRoom(conn, roomID, clientEvent)
		case events.GameLeftEvent:
			u.kickPlayer(conn, roomID, clientEvent)
		case events.KickPlayerEvent:
			u.kickPlayer(conn, roomID, clientEvent)
		case events.VoteKickIssuedEvent:
			u.voteKickPlayer(roomID, clientEvent)
		case events.StartGameEvent:
			u.startGame(conn, roomID)
		case events.FlagCellEvent:
			u.flagCell(conn, roomID, clientEvent)
		case events.OpenCellEvent:
			u.openCell(conn, roomID, clientEvent)
		case events.ChatEvent:
			u.broadcastChat(conn, roomID, clientEvent)
		case events.PositionUpdatedEvent:
			u.broadcastPosition(conn, roomID, clientEvent)
		case events.ChangeSettingsEvent:
			u.changeSettings(conn, roomID, clientEvent)
		default:
			// TODO: send some kind of error to the client
		}
	}
}

func (u *gameUsecase) createRoom(conn *websocket.Conn, roomID string, clientEvent events.ClientEvent) {
	log.Printf("Client trying to create a new room with ID %v", roomID)

	if len(u.ConnectionRooms) >= int(configs.Constant.Capacity) {
		message := events.NewFailCreateRoomUnicast("Server is full")
		u.pushUnicastMessage(roomID, conn, message)
		return
	}

	_, ok := u.ConnectionRooms[roomID]

	if ok {
		message := events.NewFailCreateRoomUnicast("Room already exists")
		u.pushUnicastMessage(roomID, conn, message)
		return
	}

	player := minesweeper.NewPlayer(clientEvent.ClientName, clientEvent.AvatarURL)
	u.createConnectionRoom(roomID)
	u.createGameRoom(roomID, player.PlayerID)
	u.registerPlayer(roomID, conn, player)

	res := events.NewRoomCreatedUnicast(u.GameRooms[roomID], "Room created successfully")
	u.pushUnicastMessage(roomID, conn, res)
}

func (u *gameUsecase) joinRoom(conn *websocket.Conn, roomID string, clientEvent events.ClientEvent) {
	log.Printf("Client trying to join room %v", roomID)

	_, ok := u.ConnectionRooms[roomID]
	if !ok {
		log.Printf("room %v does not exist", roomID)
		res := events.NewFailJoinRoomUnicast(roomID, "room does not exist")
		conn.WriteJSON(res)
		return
	}

	gameRoom := u.GameRooms[roomID]
	if gameRoom.IsUsernameExist(clientEvent.ClientName) {
		log.Printf("username %s already exist", clientEvent.ClientName)
		res := events.NewFailJoinRoomUnicast(roomID, "username already exist")
		conn.WriteJSON(res)
		return
	}

	player := minesweeper.NewPlayer(clientEvent.ClientName, clientEvent.AvatarURL)
	u.registerPlayer(roomID, conn, player)

	res := events.NewRoomJoinedUnicast(player.PlayerID, gameRoom)
	u.pushUnicastMessage(roomID, conn, res)

	broadcast := events.NewRoomJoinedBroadcast(player)
	u.pushBroadcastMessage(roomID, broadcast)
}

func (u *gameUsecase) kickPlayer(conn *websocket.Conn, roomID string, clientEvent events.ClientEvent) {
	log.Printf("Client trying to leave room %v", roomID)

	var playerID string

	if clientEvent.PlayerID == "" {
		player := u.ConnectionRooms[roomID][conn]
		if player != nil {
			playerID = u.ConnectionRooms[roomID][conn].ID
		}
		delete(u.ConnectionRooms[roomID], conn)
		if u.GameRooms[roomID] != nil {
			delete(u.GameRooms[roomID].Players, playerID)
		}
		log.Printf("delete player %s from room %s", playerID, roomID)
	} else {
		playerID = clientEvent.PlayerID
		room := u.GameRooms[roomID]
		if room == nil {
			res := events.NewVoteKickPlayerUnicast(false)
			u.pushUnicastMessage(roomID, conn, res)
			return
		}

		_, ok := room.Players[playerID]
		if !ok {
			res := events.NewVoteKickPlayerUnicast(false)
			u.pushUnicastMessage(roomID, conn, res)
			return
		}

		res := events.NewVoteKickPlayerUnicast(true)
		u.pushUnicastMessage(roomID, conn, res)

		u.GameRooms[roomID].VoteBallot[playerID] = 0
		issuerID := u.ConnectionRooms[roomID][conn].ID
		voteKickBroadcast := events.NewVoteKickPlayerBroadcast(playerID, issuerID)
		u.pushBroadcastMessage(roomID, voteKickBroadcast)
		return
	}

	_, ok := u.ConnectionRooms[roomID]
	res := events.NewGameLeftUnicast(true)
	u.pushUnicastMessage(roomID, conn, res)

	if ok {
		broadcast := events.NewGameLeftBroadcast(playerID)
		u.pushBroadcastMessage(roomID, broadcast)
	}

	gameRoom := u.GameRooms[roomID]
	if gameRoom == nil {
		return
	}

	// appoint new host if necessary
	if gameRoom.Settings.HostID == playerID {
		newHostID := gameRoom.PickRandomHost()
		changeHostBroadcast := events.NewChangeHostBroadcast(newHostID)
		u.pushBroadcastMessage(roomID, changeHostBroadcast)
	}

	if gameRoom.IsEmpty() {
		if ch, ok := u.StopScoreCronChan[roomID]; ok {
			ch <- true
		}
		delete(u.GameRooms, roomID)
		delete(u.ConnectionRooms, roomID)
		log.Printf("delete room %v", roomID)
	}
}

func (u *gameUsecase) voteKickPlayer(roomID string, clientEvent events.ClientEvent) {
	log.Printf("Client is voting on room %v", roomID)
	gameRoom := u.GameRooms[roomID]

	_, ok := gameRoom.VoteBallot[clientEvent.PlayerID]
	if !ok {
		return
	}

	if clientEvent.AgreeToKick {
		gameRoom.VoteBallot[clientEvent.PlayerID]++
	}
	log.Printf("current tally %v", gameRoom.VoteBallot[clientEvent.PlayerID])

	if clientEvent.AgreeToKick && gameRoom.VoteBallot[clientEvent.PlayerID] > len(gameRoom.Players)/2 {
		log.Printf("vote kick success, removing player")
		delete(gameRoom.VoteBallot, clientEvent.PlayerID)

		var targetConn *websocket.Conn
		connRoom := u.ConnectionRooms[roomID]

		for key, val := range connRoom {
			if val.ID == clientEvent.PlayerID {
				targetConn = key
				break
			}
		}

		if targetConn == nil {
			return
		}

		evictionNotice := events.NewGameLeftUnicast(true)
		u.pushUnicastMessage(roomID, targetConn, evictionNotice)

		broadcast := events.NewGameLeftBroadcast(clientEvent.PlayerID)
		u.pushBroadcastMessage(roomID, broadcast)

		// appoint new host if necessary
		if gameRoom.Settings.HostID == clientEvent.PlayerID {
			newHostID := gameRoom.PickRandomHost()
			changeHostBroadcast := events.NewChangeHostBroadcast(newHostID)
			u.pushBroadcastMessage(roomID, changeHostBroadcast)
		}
	}
}

func (u *gameUsecase) getPlayerID(roomID string, conn *websocket.Conn) (string, bool) {
	cRoom, ok := u.ConnectionRooms[roomID]
	if !ok {
		return "", ok
	}
	pConn, ok := cRoom[conn]
	if !ok {
		return "", ok
	}

	return pConn.ID, ok
}

func (u *gameUsecase) startGame(conn *websocket.Conn, roomID string) {
	log.Printf("Client trying to start game on room %v", roomID)
	gameRoom := u.GameRooms[roomID]
	playerID, _ := u.getPlayerID(roomID, conn)

	if playerID != gameRoom.Settings.HostID {
		res := events.NewGameStartedUnicast(false, "Only host can start the game")
		u.pushUnicastMessage(roomID, conn, res)
		return
	}

	if gameRoom.IsStarted {
		res := events.NewGameStartedUnicast(false, "Game already started")
		u.pushUnicastMessage(roomID, conn, res)
		return
	}

	if len(gameRoom.Players) < 1 {
		res := events.NewGameStartedUnicast(false, "Not enough players to start the game")
		u.pushUnicastMessage(roomID, conn, res)
		return
	}

	err := gameRoom.Start()
	if err != nil {
		res := events.NewGameStartedUnicast(false, err.Error())
		u.pushUnicastMessage(roomID, conn, res)
		return
	}
	u.setupScoreCron(roomID)

	notifContent := "game started"
	notification := events.NewNotificationBroadcast(notifContent)
	// TODO: broadcast game started, with the fields and everything
	res := events.NewGameStartedBroadcast(true, "Game started", gameRoom.Field.GetCellString())

	u.pushBroadcastMessage(roomID, res)
	u.pushBroadcastMessage(roomID, notification)
}

func (u *gameUsecase) flagCell(conn *websocket.Conn, roomID string, gameRequest events.ClientEvent) {
	gameRoom := u.GameRooms[roomID]
	// TODO: maybe add flag log with the player id in it
	// playerID := u.ConnectionRooms[roomID][conn].ID
	if !gameRoom.IsStarted {
		log.Printf("game is not started")
		// i guess we don't need to send any response here, just like a real minesweeper game
		return
	}

	var boardUpdatedBroadcast events.BoardUpdatedBroadcast

	playerID, _ := u.getPlayerID(roomID, conn)
	err := gameRoom.FlagCell(gameRequest.Row, gameRequest.Col, playerID)
	if err != nil {
		log.Printf("error flagging cell: %v", err)
		// TODO: send error response
		return
	}

	boardUpdatedBroadcast = *events.NewBoardUpdatedBroadcast(gameRoom.Field.GetCellString())

	// TODO: update the score

	u.pushBroadcastMessage(roomID, boardUpdatedBroadcast)
}

func (u *gameUsecase) openCell(conn *websocket.Conn, roomID string, gameRequest events.ClientEvent) {
	gameRoom := u.GameRooms[roomID]
	// TODO: maybe add open log with the player id in it
	// playerID := u.ConnectionRooms[roomID][conn].ID
	if !gameRoom.IsStarted {
		log.Printf("game is not started")
		// i guess we don't need to send any response here, just like a real minesweeper game
		return
	}

	playerID, _ := u.getPlayerID(roomID, conn)

	var boardUpdatedBroadcast events.BoardUpdatedBroadcast

	player := gameRoom.Players[playerID]
	points, err := gameRoom.OpenCell(gameRequest.Row, gameRequest.Col, playerID)
	if err != nil && err == minesweeper.ErrOpenMine {
		log.Printf("error opening cell: %v", err)
		player.AddScore(points)
		u.updateScore(roomID, time.Now().Unix())
		gameRoom.End()
		mineOpened := events.NewMinesOpenedBroadcast(gameRoom.Field.GetCellStringBare(), gameRoom.Players)
		u.pushBroadcastMessage(roomID, mineOpened)

		notifContent := player.Name + " opened a mine, boo!"
		notification := events.NewNotificationBroadcast(notifContent)
		u.pushBroadcastMessage(roomID, notification)

		return
	}
	player.AddScore(points)

	boardUpdatedBroadcast = *events.NewBoardUpdatedBroadcast(gameRoom.Field.GetCellString())
	u.pushBroadcastMessage(roomID, boardUpdatedBroadcast)

	if gameRoom.Field.IsCleared() {
		log.Printf("game is cleared")
		u.updateScore(roomID, time.Now().Unix())
		gameRoom.End()

		notifContent := "mines are cleared, " + player.Name + " with the last sweep!"
		notification := events.NewNotificationBroadcast(notifContent)
		u.pushBroadcastMessage(roomID, notification)

		res := events.NewGameClearedBroadcast(gameRoom.Field.GetCellStringBare(), gameRoom.Players)
		u.pushBroadcastMessage(roomID, res)
	}
}

func (u *gameUsecase) broadcastChat(conn *websocket.Conn, roomID string, gameRequest events.ClientEvent) {
	log.Printf("Client is sending chat on room %v", roomID)

	room, ok := u.ConnectionRooms[roomID]
	if ok {
		playerID := room[conn].ID
		playerName := u.GameRooms[roomID].Players[playerID].Name

		log.Printf("player %s send chat", playerName)
		broadcast := events.NewMessageBroadcast(gameRequest.Message, playerName)
		u.pushBroadcastMessage(roomID, broadcast)
	}
}

func (u *gameUsecase) broadcastPosition(conn *websocket.Conn, roomID string, gameRequest events.ClientEvent) {
	room, ok := u.ConnectionRooms[roomID]
	if ok {
		playerID := room[conn].ID

		broadcast := events.NewPositionUpdateBroadcast(playerID, gameRequest.Row, gameRequest.Col)
		u.pushBroadcastMessage(roomID, broadcast)
	}
}

func (u *gameUsecase) changeSettings(conn *websocket.Conn, roomID string, gameRequest events.ClientEvent) {
	gRoom := u.GameRooms[roomID]
	// TODO: update all the settings
	playerID, _ := u.getPlayerID(roomID, conn)

	if gRoom.Settings.HostID != playerID {
		res := events.NewChangeSettingsUnicast(false, "Only host can change the settings")
		u.pushUnicastMessage(roomID, conn, res)
		return
	}

	if gRoom.IsStarted {
		res := events.NewChangeSettingsUnicast(false, "Cannot change settings while the game is running")
		u.pushUnicastMessage(roomID, conn, res)
		return
	}

	// TODO: skip if settings is nil
	if gameRequest.Settings == nil {
		res := events.NewChangeSettingsUnicast(false, "Please specify a valid settings")
		u.pushUnicastMessage(roomID, conn, res)
		return
	}

	gRoom.Settings.Capacity = gameRequest.Settings.Capacity
	gRoom.Settings.Difficulty = gameRequest.Settings.Difficulty
	gRoom.Settings.CellScore = gameRequest.Settings.CellScore
	gRoom.Settings.MineScore = gameRequest.Settings.MineScore
	gRoom.Settings.CountColdOpen = gameRequest.Settings.CountColdOpen

	res := events.NewChangeSettingsUnicast(true, "Settings has been updated successfully")
	u.pushUnicastMessage(roomID, conn, res)
}

func (u *gameUsecase) createConnectionRoom(roomID string) {
	u.ConnectionRooms[roomID] = make(map[*websocket.Conn]*connection)
}

func (u *gameUsecase) createGameRoom(roomID string, hostID string) {
	gameRoom := minesweeper.NewGameRoom(roomID, hostID, 4)
	u.GameRooms[roomID] = gameRoom
}

func (u *gameUsecase) registerPlayer(roomID string, conn *websocket.Conn, player *minesweeper.Player) {
	u.ConnectionRooms[roomID][conn] = NewConnection(player.PlayerID)
	gameRoom := u.GameRooms[roomID]
	gameRoom.AddPlayer(player)

	go u.writePump(conn, roomID)
}

func (u *gameUsecase) unregisterPlayer(roomID string, conn *websocket.Conn, playerID string) {
	gameRoom := u.GameRooms[roomID]
	gameRoom.RemovePlayer(playerID)
	delete(u.ConnectionRooms[roomID], conn)

	// delete empty room
	if len(u.GameRooms[roomID].Players) == 0 {
		// TODO: cleanup cron
		u.StopScoreCronChan[roomID] <- true
		delete(u.GameRooms, roomID)
		delete(u.ConnectionRooms, roomID)
	}
}

func (u *gameUsecase) updateScore(roomID string, timestamp int64) {
	gameRoom := u.GameRooms[roomID]

	scoreboard := map[string]int{}
	for pID, val := range gameRoom.Players {
		scoreboard[pID] = val.Score
	}
	u.pushBroadcastMessage(roomID, events.NewScoreUpdatedBroadcast(scoreboard, timestamp))
	// here's how the new score broadcast is going to look like
	// build a map of player id -> score, maybe include a timestamp or order id as well
	// broadcast the message to the room
}

func (u *gameUsecase) setupScoreCron(roomID string) {
	gameRoom := u.GameRooms[roomID]
	gameRoom.ScoreTicker = time.NewTicker(scoreUpdateInterval)
	stopChan := make(chan bool, 1)

	u.StopScoreCronChan[roomID] = stopChan
	go func(roomID string) {
		for {
			select {
			case t := <-gameRoom.ScoreTicker.C:
				u.updateScore(roomID, t.UnixNano())
			case <-stopChan:
				return
			}
		}
	}(roomID)
}

func (u *gameUsecase) writePump(conn *websocket.Conn, roomID string) {
	c := u.ConnectionRooms[roomID][conn]
	defer func() {
		conn.Close()
	}()

	for {
		message := <-c.Queue
		err := conn.WriteJSON(message)
		if err != nil {
			log.Println("failed to write json:", err.Error())
		}

		if _, ok := message.(*events.GameLeftUnicast); ok {
			u.unregisterPlayer(roomID, conn, c.ID)
			return
		}
	}
}

func (u *gameUsecase) RunSwitch() {
	for {
		event := <-u.SwitchQueue
		conRoom := u.ConnectionRooms[event.RoomID]
		if conRoom == nil {
			continue
		}

		if event.EventType == events.UnicastSocketEvent {
			pConn := conRoom[event.Conn]
			if pConn == nil {
				log.Println("nil conn for unicast")
				continue
			}
			pConn.Queue <- event.Message
		} else {
			for _, con := range conRoom {
				con.Queue <- event.Message
			}
		}
	}
}

func (u *gameUsecase) pushUnicastMessage(roomID string, conn *websocket.Conn, message interface{}) {
	u.SwitchQueue <- events.NewUnicastEvent(roomID, conn, message)
}

func (u *gameUsecase) pushBroadcastMessage(roomID string, message interface{}) {
	u.SwitchQueue <- events.NewBroadcastEvent(roomID, message)
}
