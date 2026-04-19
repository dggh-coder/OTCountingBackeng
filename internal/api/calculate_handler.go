package api

import (
	"encoding/json"
	"net/http"
	"otcountingbackend/internal/engine"
	"otcountingbackend/internal/service"
	"strconv"
)

func (h *Handler) calculate(w http.ResponseWriter, r *http.Request, sessionID int64) {
	out, err := h.CalculateSvc.CalculateAndPersist(r.Context(), sessionID)
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, out)
}

func (h *Handler) getResult(w http.ResponseWriter, r *http.Request, sessionID int64) {
	rows, err := h.CalculateSvc.GetResults(r.Context(), sessionID)
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"results": rows})
}

func (h *Handler) getRendered(w http.ResponseWriter, r *http.Request, sessionID int64) {
	employeeID := r.URL.Query().Get("employeeId")
	fragmentType := r.URL.Query().Get("fragmentType")
	fv, _ := strconv.Atoi(r.URL.Query().Get("formatVersion"))
	if employeeID == "" || fragmentType == "" || fv <= 0 {
		writeErr(w, service.ValidationError("employeeId, fragmentType, and formatVersion are required"))
		return
	}
	row, err := h.RenderedRepo.Get(r.Context(), sessionID, employeeID, fragmentType, fv)
	if err != nil {
		writeErr(w, service.InternalError("failed to read rendered fragment"))
		return
	}
	if row == nil {
		writeErr(w, service.NotFoundError("rendered fragment not found"))
		return
	}
	writeJSON(w, http.StatusOK, row)
}

func (h *Handler) backwardCompatibleCalculate(w http.ResponseWriter, r *http.Request) {
	var in engine.Input
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		writeErr(w, service.ValidationError("invalid JSON body"))
		return
	}
	out, err := h.CalculateSvc.Engine.Calculate(in)
	if err != nil {
		writeErr(w, service.ValidationError("calculation failed", err.Error()))
		return
	}
	writeJSON(w, http.StatusOK, out)
}
