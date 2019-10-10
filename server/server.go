package server

import (
	"TraefikAccessControl/manager"
	"TraefikAccessControl/models"
	"net/http"
	"net/url"
	"strings"

	"github.com/gin-contrib/multitemplate"
	log "github.com/sirupsen/logrus"
	ginlogrus "github.com/toorop/gin-logrus"

	"github.com/gin-gonic/gin"
)

type Server struct {
	Router *gin.Engine
}

func NewServer() *Server {
	s := &Server{
		Router: gin.Default(),
	}
	s.parseTemplates()
	s.buildRoutes()
	return s
}

func (s *Server) parseTemplates() {
	r := multitemplate.NewRenderer()
	r.AddFromFiles("login", "templates/base.html", "templates/login.html")
	s.Router.HTMLRender = r
}

func (s *Server) buildRoutes() {
	ginLogger := log.New()
	s.Router.Use(ginlogrus.Logger(ginLogger), gin.Recovery())

	s.Router.GET("/access", s.accessHandler())

	s.Router.Static("/static", "static")
	s.Router.GET("/", s.dashboardHandler())
	s.Router.GET("/login", s.loginHandler())
	s.Router.GET("/forbidden", s.forbiddenHandler())
}

func (s *Server) accessHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		host := c.Request.Header.Get("X-Forwarded-Host")
		rawURL := c.Request.Header.Get("X-Forwarded-Uri")
		requestLogger := log.WithFields(log.Fields{"host": host, "rawURL": rawURL})

		if host == "" || rawURL == "" {
			requestLogger.Info("Not all needed headers present")
			c.String(http.StatusBadRequest, "Not all needed headers present")
			return
		}

		parsedURL, err := url.Parse(rawURL)
		if err != nil {
			requestLogger.Error("Failed to parse URL")
			c.String(http.StatusBadRequest, "Faild to parse URL")
			return
		}
		path := parsedURL.Path
		if !strings.HasPrefix(path, "/") {
			path = "/" + path
		}
		requestLogger = requestLogger.WithField("path", path)

		var accessGranted = false
		var user *models.User

		if cookie, err := c.Request.Cookie("tac_token"); err == nil {
			requestLogger.WithField("cookie_value", cookie.Value).Info("Cookie request")

			accessGranted, user, err = manager.GetAccessManager().CheckAccessToken(host, path, cookie.Value, false, requestLogger)
		} else if username, password, hasAuth := c.Request.BasicAuth(); hasAuth {
			requestLogger.WithField("username", username).Info("BasicAuth request")

			accessGranted, user, err = manager.GetAccessManager().CheckAccessCredentials(host, path, username, password, true, requestLogger)
		} else if bearer := c.Request.Header.Get("Authorization"); bearer != "" {
			requestLogger.WithField("bearer", bearer).Info("Bearer request")

			bearerParts := strings.Split(bearer, "Bearer")
			if len(bearerParts) == 2 {
				bearer = strings.TrimSpace(bearerParts[1])
			}
			accessGranted, user, err = manager.GetAccessManager().CheckAccessToken(host, path, bearer, true, requestLogger)
		} else {
			requestLogger.Info("No auth information in request")
		}

		isBrowser := strings.Contains(c.Request.UserAgent(), "Mozilla")
		if accessGranted {
			c.Status(http.StatusOK)
		} else if isBrowser && user == nil {
			c.Redirect(http.StatusTemporaryRedirect, s.getRedirectURL(*c.Request.URL, "/login", rawURL))
		} else if isBrowser && user != nil {
			c.Redirect(http.StatusTemporaryRedirect, s.getRedirectURL(*c.Request.URL, "/forbidden", rawURL))
		} else {
			c.String(http.StatusUnauthorized, "No access granted")
		}
	}
}

func (s *Server) getRedirectURL(reqURL url.URL, path, origURL string) string {
	redirectURL := reqURL
	redirectURL.Path = path
	q := url.Values{}
	q.Set("orig_url", origURL)
	redirectURL.RawQuery = q.Encode()

	return redirectURL.String()
}

func (s *Server) dashboardHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.String(http.StatusOK, "Dashboard")
	}
}

func (s *Server) loginHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.HTML(http.StatusOK, "login", nil)
	}
}

func (s *Server) forbiddenHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.String(http.StatusForbidden, "Access forbidden")
	}
}
