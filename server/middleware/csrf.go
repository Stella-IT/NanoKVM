package middleware

import (
	"net/http"
	"net/url"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"

	"NanoKVM-Server/config"
)

// CSRFProtection validates the Origin header on state-changing requests
// to prevent cross-site request forgery attacks.
func CSRFProtection() gin.HandlerFunc {
	return func(c *gin.Context) {
		conf := config.GetInstance()
		if conf.Authentication == "disable" {
			c.Next()
			return
		}

		// Safe methods don't need CSRF protection
		switch c.Request.Method {
		case "GET", "HEAD", "OPTIONS":
			c.Next()
			return
		}

		origin := c.GetHeader("Origin")

		// Allow requests without Origin header (non-browser clients)
		if origin == "" {
			c.Next()
			return
		}

		parsed, err := url.Parse(origin)
		if err != nil {
			log.Warnf("CSRF: invalid origin header: %s", origin)
			c.JSON(http.StatusForbidden, gin.H{"code": -1, "msg": "invalid origin"})
			c.Abort()
			return
		}

		requestHost := c.Request.Host
		originHost := parsed.Host

		if originHost != requestHost {
			log.Warnf("CSRF: origin mismatch: origin=%s, host=%s", originHost, requestHost)
			c.JSON(http.StatusForbidden, gin.H{"code": -1, "msg": "cross-origin request blocked"})
			c.Abort()
			return
		}

		c.Next()
	}
}
