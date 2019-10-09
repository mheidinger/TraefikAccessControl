package server

import (
	"TraefikAccessControl/manager"
	"TraefikAccessControl/models"
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

	s.Router.GET("/", s.dashboardHandler())
	s.Router.GET("/login", s.loginHandler())
	s.Router.GET("/forbidden", s.forbiddenHandler())
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
		var user *models.User

		if cookie, err := r.Cookie("tac_token"); err == nil {
			requestLogger.WithField("cookie_value", cookie.Value).Info("Cookie request")

			accessGranted, user, err = manager.GetAccessManager().CheckAccessToken(host, path, cookie.Value, false, requestLogger)
		} else if username, password, hasAuth := r.BasicAuth(); hasAuth {
			requestLogger.WithField("username", username).Info("BasicAuth request")

			accessGranted, user, err = manager.GetAccessManager().CheckAccessCredentials(host, path, username, password, true, requestLogger)
		} else if bearer := r.Header.Get("Authorization"); bearer != "" {
			requestLogger.WithField("bearer", bearer).Info("Bearer request")

			bearerParts := strings.Split(bearer, "Bearer")
			if len(bearerParts) == 2 {
				bearer = strings.TrimSpace(bearerParts[1])
			}
			accessGranted, user, err = manager.GetAccessManager().CheckAccessToken(host, path, bearer, true, requestLogger)
		} else {
			requestLogger.Info("No auth information in request")
		}

		isBrowser := strings.Contains(r.UserAgent(), "Mozilla")
		if accessGranted {
			w.WriteHeader(http.StatusOK)
		} else if isBrowser && user == nil {
			redirectURL := r.URL
			redirectURL.Path = "/login"
			q := url.Values{}
			q.Set("orig_url", rawURL)
			redirectURL.RawQuery = q.Encode()
			http.Redirect(w, r, redirectURL.String(), http.StatusTemporaryRedirect)
		} else if isBrowser && user != nil {
			http.Error(w, "HTML No access", http.StatusUnauthorized)
		} else {
			http.Error(w, "No access granted", http.StatusUnauthorized)
		}
	}
}

func (s *Server) dashboardHandler() httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		http.Error(w, "Dashboard!", 200)
	}
}

func (s *Server) loginHandler() httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		http.Error(w, "Login!", 200)
	}
}

func (s *Server) forbiddenHandler() httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		http.Error(w, "Access forbidden...", 200)
	}
}
