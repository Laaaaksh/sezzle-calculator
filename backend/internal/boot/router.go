package boot

import (
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"

	"github.com/Laaaaksh/sezzle-calculator/backend/internal/config"
	"github.com/Laaaaksh/sezzle-calculator/backend/internal/constants"
	"github.com/Laaaaksh/sezzle-calculator/backend/internal/httpx"
	"github.com/Laaaaksh/sezzle-calculator/backend/internal/interceptors"
	"github.com/Laaaaksh/sezzle-calculator/backend/internal/modules/calculator"
	"github.com/Laaaaksh/sezzle-calculator/backend/internal/modules/health"
)

// Modules is every feature module the router needs. Adding a feature means
// adding a field here and registering its routes in NewRouter.
type Modules struct {
	Health     health.IModule
	Calculator calculator.IModule
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
	r.MethodNotAllowed(methodNotAllowed(r))

	modules.Health.GetHandler().RegisterRoutes(r)

	// Feature modules register their routes on this sub-router.
	r.Route(constants.APIBasePath, func(api chi.Router) {
		modules.Calculator.GetHandler().RegisterRoutes(api)
		for _, register := range apiRoutes {
			register(api)
		}
	})
	return r
}

// routableMethods are the methods probed to build the Allow header of a 405.
var routableMethods = []string{
	http.MethodGet,
	http.MethodHead,
	http.MethodPost,
	http.MethodPut,
	http.MethodPatch,
	http.MethodDelete,
	http.MethodOptions,
	http.MethodConnect,
	http.MethodTrace,
}

// methodNotAllowed answers with the JSON 405 and the Allow header RFC 9110
// requires. chi only sets Allow in its own default responder and drops the
// allowed methods once a custom responder is installed, so the root router is
// probed for every method that does match the path.
func methodNotAllowed(root chi.Routes) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.RawPath
		if path == "" {
			path = r.URL.Path
		}
		allowed := make([]string, 0, len(routableMethods))
		for _, method := range routableMethods {
			if root.Match(chi.NewRouteContext(), method, path) {
				allowed = append(allowed, method)
			}
		}
		if len(allowed) > 0 {
			w.Header().Set(constants.HeaderAllow, strings.Join(allowed, ", "))
		}
		httpx.MethodNotAllowed(w, r)
	}
}
