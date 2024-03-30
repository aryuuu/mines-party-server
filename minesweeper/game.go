package minesweeper

import (
	"sync"

	"github.com/google/uuid"
)

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
	Players    map[string]*Player `json:"players"`
	VoteBallot map[string]int     `json:"-"`

	FieldWLoc sync.RWMutex `json:"-"`
	Field     *Field       `json:"-"`
}

func NewGameRoom(roomID string, hostID string, capacity int) *GameRoom {
	return &GameRoom{
		RoomID:     roomID,
		Capacity:   capacity,
		HostID:     hostID,
		IsStarted:  false,
		Players:    map[string]*Player{},
		VoteBallot: map[string]int{},
		Field:      &Field{},
	}
}

func (gr *GameRoom) IsUsernameExist(username string) bool {
	for _, player := range gr.Players {
		if player.Name == username {
			return true
		}
	}

	return false
}

func (gr *GameRoom) PickRandomHost() string {
	for id := range gr.Players {
		gr.Players[id].IsHost = true
		return id
	}
	return ""
}

func (gr *GameRoom) Start() error {
	gr.Field = NewField(10, 20, 30)
	gr.IsStarted = true

	return nil
}

func (gr *GameRoom) End() error {
	gr.IsStarted = false

	return nil
}

func (r *GameRoom) AddPlayer(player *Player) {
	r.Players[player.PlayerID] = player
}

func (r *GameRoom) RemovePlayer(id string) {
	delete(r.Players, id)
}

func (r *GameRoom) OpenCell(row, col int, playerID string) (int, error) {
	r.FieldWLoc.Lock()
	points, err := r.Field.OpenCell(row, col, playerID)
	r.FieldWLoc.Unlock()
	return points, err
}

func (r *GameRoom) FlagCell(row, col int, playerID string) error {
	_, err := r.Field.ToggleFlagCell(row, col, playerID)
	return err
}
