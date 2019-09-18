package middleware

import (
	"github.com/gin-gonic/gin"
	rscors "github.com/rs/cors/wrapper/gin"
)

var (
	// DefaultAllowedOrigins are the default allowed origins for the api, allowing for access
	// via both of our internet uplinks when using the web interface.
	DefaultAllowedOrigins = []string{"https://temporal.cloud", "https://backup.temporal.cloud"}
)

// CORSMiddleware is used to load our CORS handling logic
func CORSMiddleware(devMode bool, debug bool, allowedOrigins []string) gin.HandlerFunc {
	opts := rscors.Options{}
	if devMode {
		opts.AllowedOrigins = []string{"*"}
		opts.AllowCredentials = true
	} else if len(allowedOrigins) == 1 && allowedOrigins[0] == "*" {
		// in order to white list all origins, this appears to be the only
		// way to do so, as even when specifying "*" it doesn't appear
		// to match all domains
		opts.AllowOriginFunc = func(origin string) bool {
			return true
		}
	} else if len(allowedOrigins) > 1 {
		opts.AllowedOrigins = allowedOrigins
	}
	opts.AllowedMethods = []string{"GET", "POST", "OPTIONS", "DELETE", "PUT", "HEAD"}
	opts.AllowedHeaders = []string{
		"cache-control",
		"Authorization",
		"X-Request-ID",
		"Origin",
		"Accept",
		"Content-Type",
		"X-Requested-With",
		"user-agent",
	}
	if debug {
		opts.Debug = true
	}
	return rscors.New(opts)
}
