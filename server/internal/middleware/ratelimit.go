package middleware

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"

	"server/internal/logger"
	"server/internal/response"
)

const rateLimitKeyPrefix = "rl"

// KeyGen derives the rate-limit key for a request. The returned string
// identifies the bucket the request belongs to, such as a client IP, an API
// key, or a user ID.
type KeyGen func(c *gin.Context) string

type RateLimiterConfig struct {
	Client redis.UniversalClient
	Limit  int
	Window time.Duration
	KeyGen KeyGen
}

func rateLimitRedisKey(key string) string {
	return rateLimitKeyPrefix + ":" + key
}

// slidingWindowScript atomically trims expired entries, counts the current
// window, and records the request. It uses Redis server time so every
// instance evaluates the same window.
var slidingWindowScript = redis.NewScript(`
local key = KEYS[1]
local windowMillis = tonumber(ARGV[1])
local limit = tonumber(ARGV[2])
local member = ARGV[3]

local redisTime = redis.call("TIME")
local nowMillis = tonumber(redisTime[1]) * 1000 + math.floor(tonumber(redisTime[2]) / 1000)

redis.call("ZREMRANGEBYSCORE", key, "-inf", nowMillis - windowMillis)

local count = redis.call("ZCARD", key)
if count >= limit then
	return 0
end

redis.call("ZADD", key, nowMillis, member)
redis.call("PEXPIRE", key, windowMillis)

return 1
`)

// RateLimiter returns a Gin middleware that enforces a sliding-window log
// rate limit backed by Redis. Register it globally with router.Use, per group
// with group.Use, or per route by passing it before the route handler:
//
//	router.Use(middleware.RateLimiter(middleware.RateLimiterConfig{
//		Client: redisClient,
//		Limit:  100,
//		Window: time.Minute,
//		KeyGen: func(c *gin.Context) string {
//			return c.ClientIP()
//		},
//	}))
//
// Window is the length of the rate-limit window. The limit is enforced per
// distinct key returned by KeyGen.
//
// If Redis is unavailable, the request is allowed and the error is logged so
// the API stays available; enforcement resumes when Redis recovers.
func RateLimiter(config RateLimiterConfig) gin.HandlerFunc {
	if config.Client == nil {
		panic("middleware: RateLimiter requires a Redis client")
	}
	if config.Limit <= 0 {
		panic("middleware: RateLimiter limit must be greater than zero")
	}
	if config.Window < time.Millisecond {
		panic("middleware: RateLimiter window must be at least one millisecond")
	}
	if config.KeyGen == nil {
		panic("middleware: RateLimiter requires a KeyGen")
	}

	windowMillis := config.Window.Milliseconds()

	return func(c *gin.Context) {
		key := config.KeyGen(c)
		member := GetRequestID(c)

		allowed, err := slidingWindowScript.Run(
			c.Request.Context(),
			config.Client,
			[]string{rateLimitRedisKey(key)},
			windowMillis,
			config.Limit,
			member,
		).Int64()
		if err != nil {
			logger.Get().Error("rate limiter: redis unavailable", "error", err)
			c.Next()
			return
		}

		if allowed == 0 {
			c.Abort()
			response.Error(c, http.StatusTooManyRequests, "Rate limit exceeded")
			return
		}

		c.Next()
	}
}
