package events

import (
	"github.com/aryuuu/mines-party-server/minesweeper"
	"github.com/gorilla/websocket"
)

type SocketEvent struct {
	EventType EventType       `json:"event_type"`
	RoomID    string          `json:"id_room"`
	Conn      *websocket.Conn `json:"conn"`
	Message   interface{}     `json:"message"`
}

// ClientEvent is events coming from client to the server
type ClientEvent struct {
	EventType   EventType `json:"event_type,omitempty"`
	ClientName  string    `json:"client_name"`
	AvatarURL   string    `json:"avatar_url"`
	Message     string    `json:"message,omitempty"`
	PlayerID    string    `json:"id_player,omitempty"`
	AgreeToKick bool      `json:"agree_to_kick,omitempty"`
	Row         int       `json:"row,omitempty"`
	Col         int       `json:"col,omitempty"`
}

type EventType string

const (
	CreateRoomEvent            EventType = "create_room"
	JoinRoomEvent              EventType = "join_room"
	JoinRoomBroadcastEvent     EventType = "join_room_broadcast"
	GameLeftEvent              EventType = "leave_room"
	StartGameEvent             EventType = "start_game"
	StartGameBroadcastEvent    EventType = "start_game_broadcast"
	PauseGameEvent             EventType = "pause_game"
	HostChangedEvent           EventType = "host_changed"
	ResumeGameEvent            EventType = "resume_game"
	OpenCellEvent              EventType = "open_cell"
	FlagCellEvent              EventType = "flag_cell"
	BoardUpdatedEvent          EventType = "board_updated"
	MineOpened                 EventType = "mine_opened"
	GameCleared                EventType = "game_cleared"
	KickPlayerEvent            EventType = "kick_player"
	VoteKickIssuedEvent        EventType = "vote_kick_player"
	ChatEvent                  EventType = "chat"
	NotificationBroadcastEvent EventType = "notification"
	UnicastSocketEvent         EventType = "unicast"
	BroadcastSocketEvent       EventType = "broadcast"
)

type RoomCreatedUnicast struct {
	EventType EventType            `json:"event_type"`
	GameRoom  minesweeper.GameRoom `json:"game_room"`
	Board     *[][]string          `json:"board"`
	Success   bool                 `json:"success"`
	Message   string               `json:"message"`
}

type GameStartedBroadcast struct {
	EventType EventType `json:"event_type"`
	Success   bool      `json:"success"`
	Detail    string    `json:"detail"`
	// TODO: maybe put it in game created event?
	Board *[][]string `json:"board"`
}

type GameStartedUnicast struct {
	EventType EventType `json:"event_type"`
	Success   bool      `json:"success"`
	Detail    string    `json:"detail"`
}

// probably won't use
type GamePausedBroadcast struct {
	EventType string `json:"event_type"`
}

// probably won't use
type GameResumedBroadcast struct {
	EventType string `json:"event_type"`
}

type GameEndedBroadcast struct {
	EventType string `json:"event_type"`
	Cause     string `json:"cause"`
}

type RoomJoinedUnicast struct {
	EventType EventType             `json:"event_type"`
	GameRoom  *minesweeper.GameRoom `json:"game_room"`
	Detail    string                `json:"detail"`
}

type RoomJoinedBroacast struct {
	EventType EventType           `json:"event_type"`
	Player    *minesweeper.Player `json:"player"`
}

type VoteKickPlayerUnicast struct {
	EventType EventType `json:"event_type"`
	Success   bool      `json:"success"`
}

type VoteKickPlayerBroadcast struct {
	EventType EventType `json:"event_type"`
	PlayerID  string    `json:"id_player"`
	IssuerID  string    `json:"id_issuer"`
}

type HostChangedBroadcast struct {
	EventType EventType `json:"event_type"`
	PlayerID  string    `json:"id_player"`
}

type GameLeftUnicast struct {
	EventType EventType `json:"event_type"`
}

type GameLeftBroadcast struct {
	EventType EventType `json:"event_type"`
	PlayerID  string    `json:"id_player"`
}

type BoardUpdatedBroadcast struct {
	EventType EventType   `json:"event_type"`
	Board     *[][]string `json:"board"`
}

type MineOpenedBroadcast struct {
	EventType EventType   `json:"event_type"`
	Board     *[][]string `json:"board"`
}

type GameClearedBroadcast struct {
	EventType EventType   `json:"event_type"`
	Board     *[][]string `json:"board"`
}

type ScoreUpdatedBroadcast struct {
	EventType string `json:"event_type"`
	Scores    []struct {
		PlayerID string `json:"id_player"`
		Score    int    `json:"score"`
	} `json:"scores"`
}

type NotificationBroadcast struct {
	EventType EventType `json:"event_type"`
	Message   string    `json:"message"`
}

type ChatBroadcast struct {
	EventType EventType `json:"event_type,omitempty"`
	Sender    string    `json:"sender,omitempty"`
	Message   string    `json:"message,omitempty"`
}

func NewRoomCreatedUnicast(roomID string, host *minesweeper.Player, message string, board *[][]string) *RoomCreatedUnicast {
	return &RoomCreatedUnicast{
		EventType: CreateRoomEvent,
		GameRoom: minesweeper.GameRoom{
			RoomID: roomID,
		},
		Board:   board,
		Success: true,
		Message: message,
	}
}

func NewFailCreateRoomUnicast(roomID string, host *minesweeper.Player, message string) *RoomCreatedUnicast {
	return &RoomCreatedUnicast{
		EventType: CreateRoomEvent,
		GameRoom: minesweeper.GameRoom{
			RoomID: roomID,
		},
		Success: false,
		Message: message,
	}
}

func NewFailJoinRoomUnicast(roomID string, message string) *RoomJoinedUnicast {
	return &RoomJoinedUnicast{
		EventType: JoinRoomEvent,
		Detail:    message,
	}
}

func NewRoomJoinedUnicast(roomID string, room *minesweeper.GameRoom) *RoomJoinedUnicast {
	return &RoomJoinedUnicast{
		EventType: JoinRoomEvent,
		GameRoom:  room,
		Detail:    "success",
	}
}

func NewRoomJoinedBroadcast(player *minesweeper.Player) *RoomJoinedBroacast {
	return &RoomJoinedBroacast{
		EventType: JoinRoomBroadcastEvent,
		Player:    player,
	}
}

func NewVoteKickPlayerUnicast(success bool) *VoteKickPlayerUnicast {
	return &VoteKickPlayerUnicast{
		EventType: VoteKickIssuedEvent,
		Success:   success,
	}
}

func NewVoteKickPlayerBroadcast(playerID string, issuerID string) *VoteKickPlayerBroadcast {
	return &VoteKickPlayerBroadcast{
		EventType: VoteKickIssuedEvent,
		PlayerID:  playerID,
		IssuerID:  issuerID,
	}
}

func NewGameLeftUnicast(success bool) *GameLeftUnicast {
	return &GameLeftUnicast{
		EventType: GameLeftEvent,
	}
}

func NewGameLeftBroadcast(playerID string) *GameLeftBroadcast {
	return &GameLeftBroadcast{
		EventType: GameLeftEvent,
		PlayerID:  playerID,
	}
}

func NewChangeHostBroadcast(playerID string) *HostChangedBroadcast {
	return &HostChangedBroadcast{
		EventType: HostChangedEvent,
		PlayerID:  playerID,
	}
}

func NewGameStartedUnicast(success bool, detail string) *GameStartedUnicast {
	return &GameStartedUnicast{
		EventType: StartGameEvent,
		Success:   success,
		Detail:    detail,
	}
}

func NewGameStartedBroadcast(success bool, detail string, board *[][]string) *GameStartedBroadcast {
	return &GameStartedBroadcast{
		EventType: StartGameEvent,
		Success:   success,
		Detail:    detail,
		Board:     board,
	}
}

func NewNotificationBroadcast(message string) *NotificationBroadcast {
	return &NotificationBroadcast{
		EventType: NotificationBroadcastEvent,
		Message:   message,
	}
}

func NewBroadcastEvent(roomID string, message any) *SocketEvent {
	return &SocketEvent{
		EventType: BroadcastSocketEvent,
		RoomID:    roomID,
		Message:   message,
	}
}

func NewUnicastEvent(roomID string, conn *websocket.Conn, message any) *SocketEvent {
	return &SocketEvent{
		EventType: UnicastSocketEvent,
		RoomID:    roomID,
		Conn:      conn,
		Message:   message,
	}
}

func NewMessageBroadcast(message, sender string) *ChatBroadcast {
	return &ChatBroadcast{
		EventType: ChatEvent,
		Message:   message,
		Sender:    sender,
	}
}

func NewBoardUpdatedBroadcast(board *[][]string) *BoardUpdatedBroadcast {
	return &BoardUpdatedBroadcast{
		EventType: BoardUpdatedEvent,
		Board:     board,
	}
}

func NewMinesOpenedBroadcast(board *[][]string) *MineOpenedBroadcast {
	return &MineOpenedBroadcast{
		EventType: MineOpened,
		Board:     board,
	}
}

func NewGameClearedBroadcast(board *[][]string) *GameClearedBroadcast {
	return &GameClearedBroadcast{
		EventType: GameCleared,
		Board:     board,
	}
}
