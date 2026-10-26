package middleware

import (
	"github.com/gin-gonic/gin"
)

// KeyGen derives the rate-limit key for a request. The returned string
// identifies the bucket the request belongs to, such as a client IP, an API
// key, or a user ID.
type KeyGen func(c *gin.Context) string

// RateLimiter returns a Gin middleware that limits requests by the key
// returned from keyGen. Register it globally with router.Use, per group with
// group.Use, or per route by passing it before the route handler:
//
//	router.Use(middleware.RateLimiter(100, 60_000, func(c *gin.Context) string {
//		return c.ClientIP()
//	}))
//
// windowMillis is the length of the rate-limit window in milliseconds.
//
// The enforcement backend is not implemented yet. The middleware extracts the
// key on every request and lets the request through, so the factory and key
// derivation can stay stable while the backing limiter is added later.
func RateLimiter(limit int, windowMillis int64, keyGen KeyGen) gin.HandlerFunc {
	// Reserved for the enforcement backend; no limits are applied yet.
	_ = limit
	_ = windowMillis

	return func(c *gin.Context) {
		key := keyGen(c)

		// TODO: enforce limit requests per key within windowMillis.
		_ = key

		c.Next()
	}
}
