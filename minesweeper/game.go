package minesweeper

import "github.com/google/uuid"

type Player struct {
	PlayerID string `json:"id_player,omitempty"`
	Name     string `json:"name,omitempty"`
	Avatar   string `json:"avatar,omitempty"`
	IsHost   bool   `json:"is_host,omitempty"`
	Score    int    `json:"score,omitempty"`
}

func NewPlayer(name, avatar string) *Player {
	return &Player{
		PlayerID: uuid.NewString(),
		Name:     name,
		Avatar:   avatar,
		Score:    0,
	}
}

// GameRoom :nodoc:
type GameRoom struct {
	RoomID     string             `json:"id_room,omitempty"`
	Capacity   int                `json:"capacity,omitempty"`
	HostID     string             `json:"id_host,omitempty"`
	IsStarted  bool               `json:"is_started,omitempty"`
	PlayerMap  map[string]*Player `json:"-"`
	Count      int                `json:"count"`
	VoteBallot map[string]int     `json:"-"`
	Field      *Field             `json:"-"`
}

func (gr *GameRoom) IsUsernameExist(username string) bool {
	for _, player := range gr.PlayerMap {
		if player.Name == username {
			return true
		}
	}

	return false
}

func (gr *GameRoom) PickRandomHost() string {
	for id := range gr.PlayerMap {
		gr.PlayerMap[id].IsHost = true
		return id
	}
	return ""
}

func (gr *GameRoom) Start() error {
	gr.Field = NewField(8, 8, 10)
	gr.IsStarted = true

	return nil
}
