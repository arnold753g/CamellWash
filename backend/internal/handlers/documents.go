package handlers

import (
	"errors"
	"net/http"
	"starter-backend/internal/respond"
	"starter-backend/internal/services"
)

type DocumentHandler struct{ documents *services.DocumentService }

func NewDocumentHandler(documents *services.DocumentService) *DocumentHandler {
	return &DocumentHandler{documents: documents}
}
func (h *DocumentHandler) IssueReceipt(w http.ResponseWriter, r *http.Request) {
	h.issue(w, r, "RECEIPT")
}
func (h *DocumentHandler) IssueReception(w http.ResponseWriter, r *http.Request) {
	h.issue(w, r, "RECEPTION")
}
func (h *DocumentHandler) issue(w http.ResponseWriter, r *http.Request, kind string) {
	actor, ok := orderActor(w, r)
	if !ok {
		return
	}
	id, ok := pathID(w, r)
	if !ok {
		return
	}
	result, err := h.documents.Issue(r.Context(), actor, kind, id)
	if err != nil {
		handleDocumentError(w, r, err)
		return
	}
	w.Header().Set("Cache-Control", "no-store")
	respond.JSON(w, http.StatusOK, result, "Document ready")
}
func (h *DocumentHandler) Get(w http.ResponseWriter, r *http.Request) {
	actor, ok := orderActor(w, r)
	if !ok {
		return
	}
	id, ok := pathID(w, r)
	if !ok {
		return
	}
	result, err := h.documents.Get(r.Context(), actor, id)
	if err != nil {
		handleDocumentError(w, r, err)
		return
	}
	w.Header().Set("Cache-Control", "no-store")
	respond.JSON(w, http.StatusOK, result, "Document loaded")
}
func (h *DocumentHandler) Print(w http.ResponseWriter, r *http.Request) {
	actor, ok := orderActor(w, r)
	if !ok {
		return
	}
	id, ok := pathID(w, r)
	if !ok {
		return
	}
	var input struct {
		IdempotencyKey string `json:"idempotency_key"`
	}
	if !decodeJSON(w, r, &input) {
		return
	}
	result, err := h.documents.RequestPrint(r.Context(), actor, id, input.IdempotencyKey)
	if err != nil {
		handleDocumentError(w, r, err)
		return
	}
	w.Header().Set("Cache-Control", "no-store")
	respond.JSON(w, http.StatusOK, result, "Print dialog requested; physical printing is not confirmed")
}
func handleDocumentError(w http.ResponseWriter, r *http.Request, err error) {
	if errors.Is(err, services.ErrDocumentNotFound) {
		respond.Error(w, r, http.StatusNotFound, "NOT_FOUND", "Document was not found", nil)
		return
	}
	handlePaymentError(w, r, err)
}
