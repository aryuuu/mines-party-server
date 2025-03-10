package minesweeper

import (
	"sync"
	"time"

	"github.com/aryuuu/mines-party-server/utils"
	"github.com/google/uuid"
)

type Player struct {
	PlayerID   string       `json:"id_player,omitempty"`
	Name       string       `json:"name,omitempty"`
	Avatar     string       `json:"avatar,omitempty"`
	IsHost     bool         `json:"is_host,omitempty"`
	ScoreWLock sync.RWMutex `json:"_"`
	Score      int          `json:"score"`
	Color      string       `json:"color"`
}

func NewPlayer(name, avatar string) *Player {
	return &Player{
		PlayerID: uuid.NewString(),
		Name:     name,
		Avatar:   avatar,
		Score:    0,
		Color:    randColor(),
	}
}

func (p *Player) AddScore(val int) {
	p.ScoreWLock.Lock()
	p.Score += val
	p.ScoreWLock.Unlock()
}

func randColor() string {
	colors := []string{
		"#8fbcbb",
		"#88c0d0",
		"#81a1c1",
		"#5e81ac",
		"#bf616a",
		"#d08770",
		"#ebcb8b",
		"#a3be8c",
		"#b48ead",
	}

	return colors[utils.GenerateRandomInt(0, len(colors)-1)]
}

// GameRoom :nodoc:
type GameRoom struct {
	RoomID     string             `json:"id_room,omitempty"`
	IsStarted  bool               `json:"is_started,omitempty"`
	Players    map[string]*Player `json:"players"`
	VoteBallot map[string]int     `json:"-"`
	Settings   Settings           `json:"settings"`

	FieldWLoc sync.RWMutex `json:"-"`
	Field     *Field       `json:"-"`

	ScoreTicker *time.Ticker `json:"-"`
}

type Settings struct {
	Capacity      int    `json:"capacity"`
	HostID        string `json:"id_host"`
	Difficulty    string `json:"difficulty"`
	CellScore     int    `json:"cell_score"`
	MineScore     int    `json:"mine_score"`
	CountColdOpen bool   `json:"count_cold_open"`
}

func NewGameRoom(roomID string, hostID string, capacity int) *GameRoom {
	return &GameRoom{
		RoomID:     roomID,
		IsStarted:  false,
		Players:    map[string]*Player{},
		VoteBallot: map[string]int{},
		Field:      &Field{},
		Settings: Settings{
			Capacity:      capacity,
			HostID:        hostID,
			CellScore:     DEFAULT_CELL_POINT,
			MineScore:     DEFAULT_MINE_POINT,
			CountColdOpen: false,
		},
	}
}

func (gr *GameRoom) IsEmpty() bool {
	return len(gr.Players) == 0
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
		gr.Settings.HostID = id
		return id
	}
	return ""
}

func (gr *GameRoom) Start() error {
	gr.Field = NewFieldBuilder().
		WithDifficulty(gr.Settings.Difficulty).
		WithCellScore(gr.Settings.CellScore).
		WithMineScore(gr.Settings.MineScore).
		WithCountColdOpen(gr.Settings.CountColdOpen).
		Build()
	gr.IsStarted = true

	return nil
}

func (gr *GameRoom) End() error {
	gr.IsStarted = false
	if gr.ScoreTicker != nil {
		gr.ScoreTicker.Stop()
	}

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
