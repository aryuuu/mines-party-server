package routes

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
)

type HealthRouter struct { }

func InitHealthRouter(r *mux.Router) {
	healthRouter := &HealthRouter{ }

	r.HandleFunc("/liveness", healthRouter.LivenessHandler)
}

func (h HealthRouter) LivenessHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "OK")
}

