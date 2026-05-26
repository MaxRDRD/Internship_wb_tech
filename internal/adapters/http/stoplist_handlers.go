package http

import (
	"encoding/json"
	"net/http"
)

// Ручка для работы со стоп-листом.
// Поддерживает GET (получить весь стоп-лист),
// POST (добавить слово).
// DELETE (удалить слово)
func (h *Handler) handleStopList(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.stopListGet(w, r)
	case http.MethodPost:
		h.stopListAdd(w, r)
	case http.MethodDelete:
		h.stopListRemove(w, r)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

// Получение всего стоп-листа
func (h *Handler) stopListGet(w http.ResponseWriter, r *http.Request) {
	items, err := h.stopListUC.List(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to load stop-list")
		return
	}

	writeJSON(w, http.StatusOK, stopListResponse{Items: items})
}

// Добавление слова в стоп-лист
func (h *Handler) stopListAdd(w http.ResponseWriter, r *http.Request) {
	var req stopListRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := h.stopListUC.Add(r.Context(), req.Query); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// Удаление слова из стоп-листа
func (h *Handler) stopListRemove(w http.ResponseWriter, r *http.Request) {
	var req stopListRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := h.stopListUC.Remove(r.Context(), req.Query); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
