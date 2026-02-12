package middleware

import (
	"net/http"
	"net/url"

	log "github.com/sirupsen/logrus"

	"NanoKVM-Server/config"
)

// CheckWebSocketOrigin validates the Origin header on WebSocket upgrade requests.
// Allows non-browser clients (no Origin header) and same-host browser requests.
// When authentication is disabled, all origins are allowed.
func CheckWebSocketOrigin(r *http.Request) bool {
	conf := config.GetInstance()
	if conf.Authentication == "disable" {
		return true
	}

	origin := r.Header.Get("Origin")
	if origin == "" {
		return true
	}

	parsed, err := url.Parse(origin)
	if err != nil {
		log.Warnf("websocket: invalid origin header: %s", origin)
		return false
	}

	if parsed.Host != r.Host {
		log.Warnf("websocket: origin mismatch: origin=%s, host=%s", parsed.Host, r.Host)
		return false
	}

	return true
}
