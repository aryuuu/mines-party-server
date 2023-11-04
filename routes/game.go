package routes

import (
	"fmt"
	"log"
	"net/http"

	// gameModel "github.com/aryuuu/cepex-server/models/game"
	"github.com/aryuuu/mines-party-server/utils"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
)

type GameRouter struct {
	upgrader websocket.Upgrader
	// Rooms       map[string]map[*websocket.Conn]string
	// GameRooms   map[string]*gameModel.Room
	GameUsecase gameModel.GameUsecase
}

func InitGameRouter(r *mux.Router, upgrader websocket.Upgrader, guc gameModel.GameUsecase) {
	gameRouter := &GameRouter{
		upgrader: upgrader,
		// Rooms:       make(map[string]map[*websocket.Conn]string),
		// GameRooms:   make(map[string]*gameModel.Room),
		GameUsecase: guc,
	}

	go gameRouter.GameUsecase.RunSwitch()

	r.HandleFunc("/create", gameRouter.HandleCreateRoom)
	r.HandleFunc("/{roomID}", gameRouter.HandleGameEvent)
}

func (m GameRouter) HandleCreateRoom(w http.ResponseWriter, r *http.Request) {
	ID := utils.GenRandomString(5)
	log.Printf("Create new room with ID: %s", ID)

	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "%s", ID)
}

func (m GameRouter) HandleGameEvent(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	roomID := vars["roomID"]

	conn, err := m.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Print(err)
		return
	}

	m.GameUsecase.Connect(conn, roomID)
}
