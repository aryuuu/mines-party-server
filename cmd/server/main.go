package main

import (
	"log"
	"net/http"

	"github.com/aryuuu/mines-party-server/configs"
	"github.com/aryuuu/mines-party-server/routes"
	"github.com/aryuuu/mines-party-server/usecases"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
)

func main() {
	gameUsecase := usecases.NewGameUsecase()

	r := mux.NewRouter()
	cors := handlers.CORS(
		handlers.AllowedOrigins([]string{"*"}),
	)
	r.Use(cors)

	upgrader := websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin:     func(r *http.Request) bool { return true },
	}

	healthRouter := r.PathPrefix("/healthcheck").Subrouter()
	gameRouter := r.PathPrefix("/game").Subrouter()
	routes.InitHealthRouter(healthRouter)
	routes.InitGameRouter(gameRouter, upgrader, gameUsecase)

	server := &http.Server{
		Addr:    ":" + configs.Service.Port,
		Handler: r,
	}

	log.Printf("Listening on port %s...", configs.Service.Port)
	log.Fatal(server.ListenAndServe())
}
