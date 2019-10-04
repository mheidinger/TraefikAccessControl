package server

import (
	"io"
	"net/http"
	"net/url"

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
		method := r.Header.Get("X-Forwarded-Method")
		host := r.Header.Get("X-Forwarded-Host")
		url, err := url.Parse(r.Header.Get("X-Forwarded-Uri"))
		log.WithFields(log.Fields{"method": method, "host": host, "url": url}).Info("New request")

		if method == "" || host == "" || err != nil {
			w.WriteHeader(http.StatusBadRequest)
			io.WriteString(w, "Not all needed headers present")
			return
		}

		if cookie, err := r.Cookie("tac_token"); err == nil {
			log.WithField("cookie_value", cookie.Value).Info("Cookie request")
		} else if bearer := r.Header.Get("Authorization"); bearer != "" {
			log.WithField("bearer", bearer).Info("Bearer request")
		} else if user, _ /*password*/, hasAuth := r.BasicAuth(); hasAuth {
			log.WithField("username", user).Info("BasicAuth request")
		} else {
			log.Info("No auth information in request")
		}

		w.WriteHeader(http.StatusOK)
	}
}
