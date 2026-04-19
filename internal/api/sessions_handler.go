package api

import (
	"encoding/json"
	"net/http"
	"otcountingbackend/internal/service"
	"strconv"
	"strings"
)

type Handler struct {
	SessionSvc   *service.SessionService
	CalculateSvc *service.CalculateService
	RenderedRepo service.RenderedRepo
}

type createSessionReq struct {
	Date   string `json:"date"`
	Period string `json:"period"`
}

type replaceEntriesReq struct {
	Entries []service.EntryPayload `json:"entries"`
}

func (h *Handler) Register(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/sessions", h.createSession)
	mux.HandleFunc("PUT /api/sessions/", h.routeSessionActions)
	mux.HandleFunc("POST /api/sessions/", h.routeSessionActions)
	mux.HandleFunc("GET /api/sessions/", h.routeSessionActions)
	mux.HandleFunc("POST /api/calculate", h.backwardCompatibleCalculate)
}

func (h *Handler) createSession(w http.ResponseWriter, r *http.Request) {
	var req createSessionReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErr(w, service.ValidationError("invalid JSON body"))
		return
	}
	resp, err := h.SessionSvc.CreateOrGetSession(r.Context(), req.Date, req.Period)
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, resp)
}

func parseSessionPath(path string) (sessionID int64, action string, ok bool) {
	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) < 3 || parts[0] != "api" || parts[1] != "sessions" {
		return 0, "", false
	}
	sid, err := strconv.ParseInt(parts[2], 10, 64)
	if err != nil {
		return 0, "", false
	}
	if len(parts) == 3 {
		return sid, "", true
	}
	return sid, parts[3], true
}

func (h *Handler) routeSessionActions(w http.ResponseWriter, r *http.Request) {
	sessionID, action, ok := parseSessionPath(r.URL.Path)
	if !ok {
		http.NotFound(w, r)
		return
	}

	switch {
	case r.Method == http.MethodPut && action == "entries":
		h.replaceEntries(w, r, sessionID)
	case r.Method == http.MethodPost && action == "calculate":
		h.calculate(w, r, sessionID)
	case r.Method == http.MethodGet && action == "result":
		h.getResult(w, r, sessionID)
	case r.Method == http.MethodGet && action == "rendered":
		h.getRendered(w, r, sessionID)
	default:
		http.NotFound(w, r)
	}
}

func (h *Handler) replaceEntries(w http.ResponseWriter, r *http.Request, sessionID int64) {
	var req replaceEntriesReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErr(w, service.ValidationError("invalid JSON body"))
		return
	}
	if err := h.SessionSvc.ReplaceEntries(r.Context(), sessionID, req.Entries); err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"status": "ok"})
}
