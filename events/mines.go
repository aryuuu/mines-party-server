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
	EventType   EventType             `json:"event_type,omitempty"`
	ClientName  string                `json:"client_name"`
	AvatarURL   string                `json:"avatar_url"`
	Message     string                `json:"message,omitempty"`
	PlayerID    string                `json:"id_player,omitempty"`
	AgreeToKick bool                  `json:"agree_to_kick"`
	Row         int                   `json:"row"`
	Col         int                   `json:"col"`
	Settings    *minesweeper.Settings `json:"settings"`
}

type EventType string

const (
	CreateRoomEvent            EventType = "create_room"
	JoinRoomEvent              EventType = "join_room"
	JoinRoomBroadcastEvent     EventType = "join_room_broadcast"
	GameLeftEvent              EventType = "leave_room"
	LeaveRoomBroadcastEvent    EventType = "leave_room_broadcast"
	StartGameEvent             EventType = "start_game"
	StartGameBroadcastEvent    EventType = "start_game_broadcast"
	PauseGameEvent             EventType = "pause_game"
	ChangeSettingsEvent        EventType = "change_settings"
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
	PositionUpdatedEvent       EventType = "position_updated"
	ScoreUpdated               EventType = "score_updated"
	SettingsUpdatedEvent       EventType = "settings_updated"
	NotificationBroadcastEvent EventType = "notification"
	UnicastSocketEvent         EventType = "unicast"
	BroadcastSocketEvent       EventType = "broadcast"
)

type RoomCreatedUnicast struct {
	EventType EventType             `json:"event_type"`
	GameRoom  *minesweeper.GameRoom `json:"game_room"`
	Success   bool                  `json:"success"`
	Message   string                `json:"message"`
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

type SettingsUpdatedUnicast struct {
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
	PlayerID  string                `json:"id_player"`
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
	EventType EventType                      `json:"event_type"`
	Board     *[][]string                    `json:"board"`
	Players   map[string]*minesweeper.Player `json:"players"`
}

type GameClearedBroadcast struct {
	EventType EventType                      `json:"event_type"`
	Board     *[][]string                    `json:"board"`
	Players   map[string]*minesweeper.Player `json:"players"`
}

type ScoreUpdatedBroadcast struct {
	EventType  EventType      `json:"event_type"`
	Scoreboard map[string]int `json:"scoreboard"`
	Timestamp  int64          `json:"tick"`
}

type SettingsUpdatedBroadcast struct {
	EventType EventType            `json:"event_type"`
	Settings  minesweeper.Settings `json:"settings"`
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

type PositionBroadcast struct {
	EventType EventType `json:"event_type,omitempty"`
	SenderID  string    `json:"sender_id,omitempty"`
	Row       int       `json:"row"`
	Col       int       `json:"col"`
}

func NewRoomCreatedUnicast(room *minesweeper.GameRoom, message string) *RoomCreatedUnicast {
	return &RoomCreatedUnicast{
		EventType: CreateRoomEvent,
		GameRoom:  room,
		Success:   true,
		Message:   message,
	}
}

func NewFailCreateRoomUnicast(message string) *RoomCreatedUnicast {
	return &RoomCreatedUnicast{
		EventType: CreateRoomEvent,
		Success:   false,
		Message:   message,
	}
}

func NewFailJoinRoomUnicast(roomID string, message string) *RoomJoinedUnicast {
	return &RoomJoinedUnicast{
		EventType: JoinRoomEvent,
		Detail:    message,
	}
}

func NewRoomJoinedUnicast(playerID string, room *minesweeper.GameRoom) *RoomJoinedUnicast {
	return &RoomJoinedUnicast{
		EventType: JoinRoomEvent,
		GameRoom:  room,
		PlayerID:  playerID,
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
		EventType: LeaveRoomBroadcastEvent,
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

func NewChangeSettingsUnicast(success bool, detail string) *SettingsUpdatedUnicast {
	return &SettingsUpdatedUnicast{
		EventType: SettingsUpdatedEvent,
		Success:   success,
		Detail:    detail,
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

func NewPositionUpdateBroadcast(senderID string, row, col int) *PositionBroadcast {
	return &PositionBroadcast{
		EventType: PositionUpdatedEvent,
		Row:       row,
		Col:       col,
		SenderID:  senderID,
	}
}

func NewScoreUpdatedBroadcast(scoreboard map[string]int, timestamp int64) *ScoreUpdatedBroadcast {
	return &ScoreUpdatedBroadcast{
		EventType:  ScoreUpdated,
		Scoreboard: scoreboard,
		Timestamp:  timestamp,
	}
}

func NewBoardUpdatedBroadcast(board *[][]string) *BoardUpdatedBroadcast {
	return &BoardUpdatedBroadcast{
		EventType: BoardUpdatedEvent,
		Board:     board,
	}
}

func NewMinesOpenedBroadcast(board *[][]string, players map[string]*minesweeper.Player) *MineOpenedBroadcast {
	return &MineOpenedBroadcast{
		EventType: MineOpened,
		Board:     board,
		Players:   players,
	}
}

func NewGameClearedBroadcast(board *[][]string, players map[string]*minesweeper.Player) *GameClearedBroadcast {
	return &GameClearedBroadcast{
		EventType: GameCleared,
		Board:     board,
		Players:   players,
	}
}
