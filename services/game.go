package services

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
	SwitchQueue     chan events.SocketEvent
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
		SwitchQueue:     make(chan events.SocketEvent, 256),
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
			u.updateCell(conn, roomID, clientEvent)
		case events.OpenCellEvent:
			u.updateCell(conn, roomID, clientEvent)
		case events.ChatEvent:
			u.broadcastChat(conn, roomID, clientEvent)
		default:
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

	res := events.NewRoomCreatedUnicast(roomID, player, "Room created successfully")
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

	if len(gameRoom.PlayerMap) < 2 {
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
	res := events.NewGameStartedBroadcast(true, "Game started")

	u.pushMessage(true, roomID, conn, res)
	u.pushMessage(true, roomID, conn, notification)
}

func (u *gameUsecase) updateCell(conn *websocket.Conn, roomID string, gameRequest events.ClientEvent) {
	gameRoom := u.GameRooms[roomID]
	playerID := u.ConnectionRooms[roomID][conn].ID
	if !gameRoom.IsStarted {
		log.Printf("game is not started")
		res := events.NewPlayCardResponse(false, nil, 3, "Game is not started")
		u.pushMessage(false, roomID, conn, res)
		return
	}

	if gameRoom.TurnID != playerID {
		log.Printf("its not your turn yet")
		res := events.NewPlayCardResponse(false, nil, 3, "Please wait for your turn")
		u.pushMessage(false, roomID, conn, res)
		return
	}

	playerIndex := gameRoom.GetPlayerIndex(playerID)

	player := gameRoom.PlayerMap[playerID]

	if !player.IsAlive {
		log.Printf("this player is dead")
		res := events.NewPlayCardResponse(false, nil, 3, "You are already dead")
		u.pushMessage(false, roomID, conn, res)
		return
	}

	playedCard := player.Hand[gameRequest.HandIndex]
	log.Printf("%v is playing: %v", player.Name, playedCard)

	var res events.PlayCardResponse

	success := true
	if err := gameRoom.PlayCard(playerID, gameRequest.HandIndex, gameRequest.IsAdd, gameRequest.PlayerID); err != nil {
		success = false

		if !gameRequest.IsDiscard {
			player.InsertHand(playedCard, gameRequest.HandIndex)
		}
	}

	if len(player.Hand) == 0 {
		player.IsAlive = false
		deadBroadcast := events.NewDeadPlayerBroadcast(player.PlayerID)
		u.pushMessage(true, roomID, conn, deadBroadcast)
	}

	if winner := gameRoom.GetWinner(); winner != nil && winner.PlayerID != "" {
		gameRoom.EndGame(winner.PlayerID)
		endBroadcast := events.NewEndGameBroadcast(winner)
		u.pushMessage(true, roomID, conn, endBroadcast)
	}

	message := ""
	status := 0
	if !success && !gameRequest.IsDiscard {
		status = 1
		res = events.NewPlayCardResponse(false, player.Hand, status, "Try discarding hand")
		res.HandIndex = gameRequest.HandIndex
		u.pushMessage(false, roomID, conn, res)
		return
	}

	if !success && gameRequest.IsDiscard {
		message = "Hand discarded"
	}
	res = events.NewPlayCardResponse(success, player.Hand, status, message)
	u.pushMessage(false, roomID, conn, res)

	var nextPlayerId string
	if gameRoom.IsStarted {
		if gameRoom.TurnID == playerID {
			nextPlayerId = gameRoom.NextPlayer(playerIndex)
		} else {
			nextPlayerId = gameRoom.TurnID
		}
	}

	if !success {
		playedCard = minesweeper.Card{}
	}
	broadcast := events.NewPlayCardBroadcast(playedCard, gameRoom.Count, gameRoom.IsClockwise, nextPlayerId)
	u.pushMessage(true, roomID, conn, broadcast)
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

// func (u *gameUsecase) dealCard(roomID string) {
// 	room := u.ConnectionRooms[roomID]

// 	for connection, playerID := range room {
// 		player := u.GameRooms[roomID].PlayerMap[playerID.ID]
// 		message := events.NewInitialHandResponse(player.Hand)
// 		u.pushMessage(false, roomID, connection, message)
// 	}
// }

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
