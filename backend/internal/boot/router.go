package boot

import (
	"github.com/go-chi/chi/v5"

	"github.com/Laaaaksh/sezzle-calculator/backend/internal/config"
	"github.com/Laaaaksh/sezzle-calculator/backend/internal/constants"
	"github.com/Laaaaksh/sezzle-calculator/backend/internal/httpx"
	"github.com/Laaaaksh/sezzle-calculator/backend/internal/interceptors"
	"github.com/Laaaaksh/sezzle-calculator/backend/internal/modules/health"
)

// Modules is every feature module the router needs. Adding a feature means
// adding a field here and registering its routes in NewRouter.
type Modules struct {
	Health health.IModule
}

// RouteRegistrar adds a feature's routes to the versioned API router.
type RouteRegistrar func(api chi.Router)

// NewRouter assembles the middleware chain and routes. Middleware order
// matters:
//
//	request ID  first, so every later log line and error body carries it
//	recovery    before the log, so a panic still gets a 500 and an access line
//	access log  wraps everything below, to time the real work
//	CORS        before the limits, so a rejected preflight costs nothing
//	body limit  before handlers read
//	timeout     innermost, so the deadline covers only handler work
func NewRouter(cfg *config.Config, modules Modules, apiRoutes ...RouteRegistrar) chi.Router {
	r := chi.NewRouter()
	r.Use(
		interceptors.RequestID,
		interceptors.Recovery,
		interceptors.AccessLog,
		interceptors.CORS(cfg.CORS.AllowedOrigins),
		interceptors.BodyLimit(cfg.Server.MaxBodyBytes),
		interceptors.Timeout(cfg.Server.RequestTimeout),
	)
	// Set before any Route/Mount: chi copies these handlers onto sub-routers
	// only if they exist when the sub-router is mounted.
	r.NotFound(httpx.NotFound)
	r.MethodNotAllowed(httpx.MethodNotAllowed)

	modules.Health.GetHandler().RegisterRoutes(r)

	// Feature modules register their routes on this sub-router.
	r.Route(constants.APIBasePath, func(api chi.Router) {
		for _, register := range apiRoutes {
			register(api)
		}
	})
	return r
}
