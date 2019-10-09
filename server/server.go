package server

import (
	"TraefikAccessControl/manager"
	"net/http"
	"net/url"
	"strings"

	log "github.com/sirupsen/logrus"

	"github.com/julienschmidt/httprouter"
)

type Server struct {
	Router *httprouter.Router
}

func NewServer() *Server {
	s := &Server{}
	s.buildRoutes()
	return s
}

func (s *Server) buildRoutes() {
	s.Router = httprouter.New()

	s.Router.GET("/access", s.accessHandler())
}

func (s *Server) accessHandler() httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		host := r.Header.Get("X-Forwarded-Host")
		rawURL := r.Header.Get("X-Forwarded-Uri")
		requestLogger := log.WithFields(log.Fields{"host": host, "rawURL": rawURL})

		if host == "" || rawURL == "" {
			requestLogger.Info("Not all needed headers present")
			http.Error(w, "Not all needed headers present", http.StatusBadRequest)
			return
		}

		parsedURL, err := url.Parse(rawURL)
		if err != nil {
			requestLogger.Error("Failed to parse URL")
			http.Error(w, "Faild to parse URL", http.StatusBadRequest)
			return
		}
		path := parsedURL.Path
		if !strings.HasPrefix(path, "/") {
			path = "/" + path
		}
		requestLogger = requestLogger.WithField("path", path)

		var accessGranted = false
		var redirectToLogin = false

		if cookie, err := r.Cookie("tac_token"); err == nil {
			requestLogger.WithField("cookie_value", cookie.Value).Info("Cookie request")

			accessGranted, err = manager.GetAccessManager().CheckAccessToken(host, path, cookie.Value, false, requestLogger)
			redirectToLogin = true
		} else if username, password, hasAuth := r.BasicAuth(); hasAuth {
			requestLogger.WithField("username", username).Info("BasicAuth request")

			accessGranted, err = manager.GetAccessManager().CheckAccessCredentials(host, path, username, password, true, requestLogger)
		} else if bearer := r.Header.Get("Authorization"); bearer != "" {
			requestLogger.WithField("bearer", bearer).Info("Bearer request")

			bearerParts := strings.Split(bearer, "Bearer")
			if len(bearerParts) == 2 {
				bearer = strings.TrimSpace(bearerParts[1])
			}
			accessGranted, err = manager.GetAccessManager().CheckAccessToken(host, path, bearer, true, requestLogger)
		} else {
			requestLogger.Info("No auth information in request")
			redirectToLogin = true
		}

		if accessGranted {
			w.WriteHeader(http.StatusOK)
		} else if redirectToLogin {
			http.Redirect(w, r, "loginurl", http.StatusTemporaryRedirect)
		} else {
			http.Error(w, "No access granted", http.StatusUnauthorized)
		}
	}
}
