package routes

import (
	"fmt"
	"log"
	"net/http"

	"github.com/aryuuu/mines-party-server/usecases"
	"github.com/aryuuu/mines-party-server/utils"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
)

type GameRouter struct {
	upgrader websocket.Upgrader
	GameUsecase usecases.GameUsecase
}

func InitGameRouter(r *mux.Router, upgrader websocket.Upgrader, guc usecases.GameUsecase) {
	gameRouter := &GameRouter{
		upgrader: upgrader,
		GameUsecase: guc,
	}

	go gameRouter.GameUsecase.RunSwitch()

	r.HandleFunc("/create", gameRouter.HandleCreateRoom)
	r.HandleFunc("/{roomID}", gameRouter.HandleGameEvent)
}

func (m GameRouter) HandleCreateRoom(w http.ResponseWriter, r *http.Request) {
	ID := utils.GenRandomString(5)
	log.Printf("Create new room with ID: %s", ID)

	// TODO: check for duplicate room ID
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
