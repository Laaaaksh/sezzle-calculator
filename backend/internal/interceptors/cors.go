package interceptors

import (
	"net/http"

	"github.com/go-chi/cors"

	"github.com/Laaaaksh/sezzle-calculator/backend/internal/constants"
)

// CORS lets the browser frontend call the API from the configured origins
// only. Credentials are off: the API is stateless and uses no cookies.
func CORS(allowedOrigins []string) func(http.Handler) http.Handler {
	return cors.Handler(cors.Options{
		AllowedOrigins:   allowedOrigins,
		AllowedMethods:   corsAllowedMethods,
		AllowedHeaders:   corsAllowedHeaders,
		ExposedHeaders:   corsExposedHeaders,
		AllowCredentials: false,
		MaxAge:           constants.CORSMaxAgeSeconds,
	})
}
