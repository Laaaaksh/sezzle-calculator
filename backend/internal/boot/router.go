package boot

import (
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/Laaaaksh/sezzle-calculator/backend/internal/config"
	"github.com/Laaaaksh/sezzle-calculator/backend/internal/constants"
	"github.com/Laaaaksh/sezzle-calculator/backend/internal/httpx"
	"github.com/Laaaaksh/sezzle-calculator/backend/internal/interceptors"
	"github.com/Laaaaksh/sezzle-calculator/backend/internal/modules/calculator"
	"github.com/Laaaaksh/sezzle-calculator/backend/internal/modules/health"
	"github.com/Laaaaksh/sezzle-calculator/backend/internal/ratelimit"
)

// Modules is every feature module the router needs.
type Modules struct {
	Health     health.IModule
	Calculator calculator.IModule
}

// NewRouter assembles the middleware chain and routes. Order matters: request ID first, so every log line and error
// carries it; recovery outside the access log, so a panic still gets a 500 and a log line; CORS before the body limit
// and timeout, so a rejected preflight costs nothing; timeout innermost, so the deadline covers only handler work.
func NewRouter(cfg *config.Config, modules Modules) chi.Router {
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

	r.Route(constants.APIBasePath, func(api chi.Router) {
		modules.Calculator.GetHandler().RegisterRoutes(api, rateLimitGuards(cfg.RateLimit)...)
	})
	return r
}

// rateLimitGuards returns nothing when the limit is off. The limiter is built once, so all requests share one set of buckets.
func rateLimitGuards(cfg config.RateLimitConfig) []func(http.Handler) http.Handler {
	if !cfg.Enabled {
		return nil
	}
	limiter := ratelimit.New(cfg.RequestsPerSecond, cfg.Burst, time.Now)
	return []func(http.Handler) http.Handler{interceptors.RateLimit(limiter, cfg.ClientIPHeader)}
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

// methodNotAllowed sends the JSON 405 with the Allow header RFC 9110 requires. chi drops the allowed methods once a
// custom responder is installed, so the root router is probed for every method that matches the path.
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
