package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

const (
	maxLoginFailures   = 5
	loginBlockDuration = 5 * time.Minute
	cleanupInterval    = 10 * time.Minute
)

type loginAttempt struct {
	failures  int
	blockedAt time.Time
}

var (
	loginAttempts   = make(map[string]*loginAttempt)
	loginAttemptsMu sync.Mutex
)

func init() {
	go cleanupLoginAttempts()
}

func LoginRateLimit() gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.ClientIP()

		loginAttemptsMu.Lock()
		attempt, exists := loginAttempts[ip]
		if exists && attempt.failures >= maxLoginFailures {
			if time.Since(attempt.blockedAt) < loginBlockDuration {
				loginAttemptsMu.Unlock()
				c.JSON(http.StatusTooManyRequests, gin.H{
					"code": -1,
					"msg":  "too many login attempts, try again later",
				})
				c.Abort()
				return
			}
			delete(loginAttempts, ip)
		}
		loginAttemptsMu.Unlock()

		c.Next()
	}
}

func RecordLoginFailure(ip string) {
	loginAttemptsMu.Lock()
	defer loginAttemptsMu.Unlock()

	attempt, exists := loginAttempts[ip]
	if !exists {
		loginAttempts[ip] = &loginAttempt{failures: 1, blockedAt: time.Now()}
		return
	}
	attempt.failures++
	attempt.blockedAt = time.Now()
}

func ResetLoginAttempts(ip string) {
	loginAttemptsMu.Lock()
	defer loginAttemptsMu.Unlock()

	delete(loginAttempts, ip)
}

func cleanupLoginAttempts() {
	for {
		time.Sleep(cleanupInterval)

		loginAttemptsMu.Lock()
		for ip, attempt := range loginAttempts {
			if time.Since(attempt.blockedAt) > loginBlockDuration {
				delete(loginAttempts, ip)
			}
		}
		loginAttemptsMu.Unlock()
	}
}
