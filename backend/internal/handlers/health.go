package handlers

import (
	"net/http"

	"starter-backend/internal/respond"
)

func Health(w http.ResponseWriter, r *http.Request) {
	respond.JSON(w, http.StatusOK, map[string]string{"status": "ok"}, "")
}
