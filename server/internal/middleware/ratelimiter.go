package middleware

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"

	"server/internal/logger"
	"server/internal/response"
)

const rateLimitKeyPrefix = "rl"

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

var slidingWindowScript = redis.NewScript(`
local key = KEYS[1]
local windowMillis = tonumber(ARGV[1])
local limit = tonumber(ARGV[2])
local member = ARGV[3]

local redisTime = redis.call("TIME")
local nowMillis = tonumber(redisTime[1]) * 1000 + math.floor(tonumber(redisTime[2]) / 1000)

redis.call("ZREMRANGEBYSCORE", key, "-inf", nowMillis - windowMillis)

local count = redis.call("ZCARD", key)
local allowed = 0
if count < limit then
        redis.call("ZADD", key, nowMillis, member)
        redis.call("PEXPIRE", key, windowMillis)
        count = count + 1
        allowed = 1
end

local remaining = limit - count
if remaining < 0 then
        remaining = 0
end

local resetMillis = nowMillis + windowMillis
local oldest = redis.call("ZRANGE", key, 0, 0, "WITHSCORES")
if #oldest > 0 then
        resetMillis = tonumber(oldest[2]) + windowMillis
end

local retryAfter = 0
if allowed == 0 then
        retryAfter = math.ceil((resetMillis - nowMillis) / 1000)
        if retryAfter < 1 then
                retryAfter = 1
        end
end

return {allowed, remaining, math.floor(resetMillis / 1000), retryAfter}
`)

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

		res, err := slidingWindowScript.Run(
			c.Request.Context(),
			config.Client,
			[]string{rateLimitRedisKey(key)},
			windowMillis,
			config.Limit,
			member,
		).Slice()
		if err != nil {
			log := logger.Get()
			log.Error("rate limiter: redis unavailable", "error", err)
			c.Next()
			return
		}

		log := logger.Get()
		if len(res) < 4 {
			log.Error("rate limiter: unexpected response length from redis script")
			c.Next()
			return
		}

		allowed, ok0 := res[0].(int64)
		remaining, ok1 := res[1].(int64)
		reset, ok2 := res[2].(int64)
		retryAfter, ok3 := res[3].(int64)

		if !ok0 || !ok1 || !ok2 || !ok3 {
			log.Error("rate limiter: unexpected response types from redis script")
			c.Next()
			return
		}

		c.Header("X-RateLimit-Limit", strconv.Itoa(config.Limit))
		c.Header("X-RateLimit-Remaining", strconv.FormatInt(remaining, 10))
		c.Header("X-RateLimit-Reset", strconv.FormatInt(reset, 10))

		if allowed == 0 {
			c.Header("Retry-After", strconv.FormatInt(retryAfter, 10))
			c.Abort()
			response.Error(c, http.StatusTooManyRequests, "Rate limit exceeded")
			return
		}

		c.Next()
	}
}
