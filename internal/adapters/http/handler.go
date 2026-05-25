package http

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"topq/internal/domain"
	"topq/internal/usecase"
)

const maxTopN = 1000

type Handler struct {
	top          *usecase.Top
	stopListUC   *usecase.StopList
	defaultTopN  int
	windowSecond int
}

func NewHandler(top *usecase.Top, stopList *usecase.StopList, defaultTopN int, windowSecond int) *Handler {
	return &Handler{top: top, stopListUC: stopList, defaultTopN: defaultTopN, windowSecond: windowSecond}
}

func (h *Handler) Register(mux *http.ServeMux) {
	mux.HandleFunc("/health", h.health)
	mux.HandleFunc("/top", h.getTop)
	mux.HandleFunc("/stoplist", h.handleStopList)
}

func (h *Handler) health(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("ok"))
}

func (h *Handler) getTop(w http.ResponseWriter, r *http.Request) {
	n := h.defaultTopN
	if v := r.URL.Query().Get("n"); v != "" {
		if parsed, err := strconv.Atoi(v); err == nil {
			n = parsed
		}
	}
	if n < 0 {
		n = 0
	}
	if n > maxTopN {
		n = maxTopN
	}

	items, err := h.top.Get(r.Context(), n)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to get top")
		return
	}

	resp := topResponse{
		WindowSeconds: h.windowSecond,
		GeneratedAt:   time.Now().UTC().Format(time.RFC3339Nano),
		Items:         items,
	}

	writeJSON(w, http.StatusOK, resp)
}

type topResponse struct {
	WindowSeconds int              `json:"window_seconds"`
	GeneratedAt   string           `json:"generated_at"`
	Items         []domain.TopItem `json:"items"`
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}
