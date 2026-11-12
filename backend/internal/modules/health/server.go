package health

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/Laaaaksh/sezzle-calculator/backend/internal/httpx"
	"github.com/Laaaaksh/sezzle-calculator/backend/internal/modules/health/entities"
	"github.com/Laaaaksh/sezzle-calculator/backend/pkg/apperror"
)

// HTTPHandler is the transport layer of the health module.
type HTTPHandler struct {
	core ICore
}

// NewHTTPHandler returns a handler backed by core.
func NewHTTPHandler(core ICore) *HTTPHandler {
	return &HTTPHandler{core: core}
}

// RegisterRoutes mounts the health endpoints on r. Routes are registered flat
// (not under a sub-router) so GET is the only method each path accepts and
// anything else is answered by the router's JSON 405.
func (h *HTTPHandler) RegisterRoutes(r chi.Router) {
	r.Get(entities.PathLive, h.Live)
	r.Get(entities.PathReady, h.Ready)
}

// Live answers 200 while the process can serve a request at all. It checks no
// dependency on purpose: a failing dependency must not get the process killed
// and restarted.
func (h *HTTPHandler) Live(w http.ResponseWriter, r *http.Request) {
	httpx.WriteJSON(r.Context(), w, http.StatusOK, entities.LiveResponse{Status: entities.StatusOK})
}

// Ready answers 200 when the core reports ready and 503 otherwise.
func (h *HTTPHandler) Ready(w http.ResponseWriter, r *http.Request) {
	if !h.core.IsReady(r.Context()) {
		httpx.WriteError(r.Context(), w, apperror.New(apperror.CodeServiceUnavailable, ErrNotReady))
		return
	}
	httpx.WriteJSON(r.Context(), w, http.StatusOK, entities.ReadyResponse{Status: entities.StatusReady})
}
