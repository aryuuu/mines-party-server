package usecases

import (
	"log"

	"github.com/aryuuu/mines-party-server/configs"
	"github.com/aryuuu/mines-party-server/events"
	"github.com/aryuuu/mines-party-server/minesweeper"
	"github.com/gorilla/websocket"
)

type connection struct {
	ID    string
	Queue chan interface{}
}

type LeaderboardItem struct {
	PlayerID string `json:"id_player,omitempty"`
	// TODO: more fields?
}

type gameUsecase struct {
	ConnectionRooms map[string]map[*websocket.Conn]*connection
	GameRooms       map[string]*minesweeper.GameRoom
	SwitchQueue     chan *events.SocketEvent
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
		ConnectionRooms: make(map[string]map[*websocket.Conn]*connection),
		GameRooms:       make(map[string]*minesweeper.GameRoom),
		SwitchQueue:     make(chan *events.SocketEvent, 256),
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
		log.Printf("clientEvent: %v", clientEvent)

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
			u.voteKickPlayer(conn, roomID, clientEvent)
		case events.StartGameEvent:
			u.startGame(conn, roomID)
		case events.FlagCellEvent:
			u.flagCell(conn, roomID, clientEvent)
		case events.OpenCellEvent:
			u.openCell(conn, roomID, clientEvent)
		case events.ChatEvent:
			u.broadcastChat(conn, roomID, clientEvent)
		default:
		// TODO: send some kind of error to the client
		}
	}
}

func (u *gameUsecase) createRoom(conn *websocket.Conn, roomID string, clientEvent events.ClientEvent) {
	log.Printf("Client trying to create a new room with ID %v", roomID)

	if len(u.ConnectionRooms) >= int(configs.Constant.Capacity) {
		message := events.NewFailCreateRoomUnicast(roomID, nil, "Server is full")
		u.pushMessage(false, roomID, conn, message)
		return
	}

	_, ok := u.ConnectionRooms[roomID]

	if ok {
		message := events.NewFailCreateRoomUnicast(roomID, nil, "Room already exists")
		u.pushMessage(false, roomID, conn, message)
		return
	}

	player := minesweeper.NewPlayer(clientEvent.ClientName, clientEvent.AvatarURL)

	u.createConnectionRoom(roomID, conn)
	u.createGameRoom(roomID, player.PlayerID)
	u.registerPlayer(roomID, conn, player)

	res := events.NewRoomCreatedUnicast(roomID, player, "Room created successfully", u.GameRooms[roomID].Field.GetCellString())
	u.pushMessage(false, roomID, conn, res)
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

	log.Printf("found room %v", roomID)
	gameRoom := u.GameRooms[roomID]
	if gameRoom.IsUsernameExist(clientEvent.ClientName) {
		log.Printf("username %s already exist", clientEvent.ClientName)
		res := events.NewFailJoinRoomUnicast(roomID, "username already exist")
		conn.WriteJSON(res)
		return
	}

	player := minesweeper.NewPlayer(clientEvent.ClientName, clientEvent.AvatarURL)
	u.registerPlayer(roomID, conn, player)

	res := events.NewRoomJoinedUnicast(roomID, gameRoom)
	u.pushMessage(false, roomID, conn, res)

	broadcast := events.NewRoomJoinedBroadcast(player)
	u.pushMessage(true, roomID, nil, broadcast)
}

func (u *gameUsecase) kickPlayer(conn *websocket.Conn, roomID string, clientEvent events.ClientEvent) {
	log.Printf("Client trying to leave room %v", roomID)

	var playerID string

	if clientEvent.PlayerID == "" {
		player := u.ConnectionRooms[roomID][conn]
		if player != nil {
			playerID = u.ConnectionRooms[roomID][conn].ID
		}
	} else {
		playerID = clientEvent.PlayerID
		room := u.GameRooms[roomID]
		if room == nil {
			res := events.NewVoteKickPlayerUnicast(false)
			u.pushMessage(false, roomID, conn, res)
			return
		}

		_, ok := room.PlayerMap[playerID]
		if !ok {
			res := events.NewVoteKickPlayerUnicast(false)
			u.pushMessage(false, roomID, conn, res)
			return
		}

		res := events.NewVoteKickPlayerUnicast(true)
		u.pushMessage(false, roomID, conn, res)

		u.GameRooms[roomID].VoteBallot[playerID] = 0
		issuerID := u.ConnectionRooms[roomID][conn].ID
		voteKickBroadcast := events.NewVoteKickPlayerBroadcast(playerID, issuerID)
		u.pushMessage(true, roomID, conn, voteKickBroadcast)
		return
	}

	_, ok := u.ConnectionRooms[roomID]
	res := events.NewGameLeftUnicast(true)
	u.pushMessage(false, roomID, conn, res)

	if ok {
		broadcast := events.NewGameLeftBroadcast(playerID)
		u.pushMessage(true, roomID, conn, broadcast)
	}

	gameRoom := u.GameRooms[roomID]

	if gameRoom == nil {
		return
	}

	// appoint new host if necessary
	if gameRoom.HostID == playerID {
		newHostID := gameRoom.PickRandomHost()
		changeHostBroadcast := events.NewChangeHostBroadcast(newHostID)
		u.pushMessage(true, roomID, conn, changeHostBroadcast)
	}
}

func (u *gameUsecase) voteKickPlayer(conn *websocket.Conn, roomID string, clientEvent events.ClientEvent) {
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

	if clientEvent.AgreeToKick && gameRoom.VoteBallot[clientEvent.PlayerID] > len(gameRoom.PlayerMap)/2 {
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
		u.pushMessage(false, roomID, targetConn, evictionNotice)

		broadcast := events.NewGameLeftBroadcast(clientEvent.PlayerID)
		u.pushMessage(true, roomID, conn, broadcast)

		// appoint new host if necessary
		if gameRoom.HostID == clientEvent.PlayerID {
			newHostID := gameRoom.PickRandomHost()
			changeHostBroadcast := events.NewChangeHostBroadcast(newHostID)
			u.pushMessage(true, roomID, conn, changeHostBroadcast)
		}
	}
}

func (u *gameUsecase) startGame(conn *websocket.Conn, roomID string) {
	log.Printf("Client trying to start game on room %v", roomID)
	gameRoom := u.GameRooms[roomID]
	playerID := u.ConnectionRooms[roomID][conn].ID

	if playerID != gameRoom.HostID {
		res := events.NewGameStartedUnicast(false, "Only host can start the game")
		u.pushMessage(false, roomID, conn, res)
		return
	}

	if gameRoom.IsStarted {
		res := events.NewGameStartedUnicast(false, "Game already started")
		u.pushMessage(false, roomID, conn, res)
		return
	}

	if len(gameRoom.PlayerMap) < 1 {
		res := events.NewGameStartedUnicast(false, "Not enough players to start the game")
		u.pushMessage(false, roomID, conn, res)
		return
	}

	err := gameRoom.Start()
	if err != nil {
		res := events.NewGameStartedUnicast(false, err.Error())
		u.pushMessage(false, roomID, conn, res)
		return
	}

	// TODO: broadcast game started, with the fields and everything
	// u.dealCard(roomID)

	notifContent := "game started"
	notification := events.NewNotificationBroadcast(notifContent)
	// TODO: broadcast game started, with the fields and everything
	res := events.NewGameStartedBroadcast(true, "Game started", gameRoom.Field.GetCellString())

	u.pushMessage(true, roomID, conn, res)
	u.pushMessage(true, roomID, conn, notification)
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

	// TODO: maybe add flag log with the player id in it
	// player := gameRoom.PlayerMap[playerID]

	var boardUpdatedBroadcast events.BoardUpdatedBroadcast

	err := gameRoom.FlagCell(gameRequest.Row, gameRequest.Col)
	if err != nil {
		log.Printf("error flagging cell: %v", err)
		// TODO: send error response
		return
	}

	boardUpdatedBroadcast = *events.NewBoardUpdatedBroadcast(gameRoom.Field.GetCellString())

	// TODO: update the score
	// TODO: broadcast updated board/field

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

	// player := gameRoom.PlayerMap[playerID]

	var boardUpdatedBroadcast events.BoardUpdatedBroadcast

	err := gameRoom.OpenCell(gameRequest.Row, gameRequest.Col)
	if err != nil {
		log.Printf("error opening cell: %v", err)
		// TODO: send error response
		gameRoom.End()
		minesExplodeEvent := events.NewMinesOpenedBroadcast(gameRoom.Field.GetCellString())
		u.pushBroadcastMessage(roomID, minesExplodeEvent)
		// gameOverEvent := events.
		return
	}

	boardUpdatedBroadcast = *events.NewBoardUpdatedBroadcast(gameRoom.Field.GetCellString())
	// TODO: update the score
	// TODO: broadcast updated board/field

	u.pushBroadcastMessage(roomID, boardUpdatedBroadcast)
}

func (u *gameUsecase) broadcastChat(conn *websocket.Conn, roomID string, gameRequest events.ClientEvent) {
	log.Printf("Client is sending chat on room %v", roomID)

	room, ok := u.ConnectionRooms[roomID]
	if ok {
		playerID := room[conn].ID
		playerName := u.GameRooms[roomID].PlayerMap[playerID].Name

		log.Printf("player %s send chat", playerName)
		broadcast := events.NewMessageBroadcast(gameRequest.Message, playerName)
		u.pushMessage(true, roomID, conn, broadcast)
	}
}

func (u *gameUsecase) createConnectionRoom(roomID string, conn *websocket.Conn) {
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
	if len(u.GameRooms[roomID].PlayerMap) == 0 {
		log.Printf("delete room %v", roomID)
		delete(u.GameRooms, roomID)
		delete(u.ConnectionRooms, roomID)
	}
}

func (u *gameUsecase) writePump(conn *websocket.Conn, roomID string) {
	c := u.ConnectionRooms[roomID][conn]

	defer func() {
		conn.Close()
	}()

	for {
		message := <-c.Queue
		conn.WriteJSON(message)

		if _, ok := message.(events.GameLeftBroadcast); ok {
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

func (u *gameUsecase) pushMessage(broadcast bool, roomID string, conn *websocket.Conn, message interface{}) {
	if broadcast {
		event := events.NewBroadcastEvent(roomID, message)
		u.SwitchQueue <- event
	} else {
		event := events.NewUnicastEvent(roomID, conn, message)
		u.SwitchQueue <- event
	}
}

func (u *gameUsecase) pushUnicastMessage(roomID string, conn *websocket.Conn, message interface{}) {
	u.SwitchQueue <- events.NewUnicastEvent(roomID, conn, message)
}

func (u *gameUsecase) pushBroadcastMessage(roomID string, message interface{}) {
	u.SwitchQueue <- events.NewBroadcastEvent(roomID, message)
}
