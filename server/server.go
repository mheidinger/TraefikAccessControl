package server

import (
	"TraefikAccessControl/manager"
	"TraefikAccessControl/models"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/gin-contrib/multitemplate"
	log "github.com/sirupsen/logrus"
	ginlogrus "github.com/toorop/gin-logrus"
	"github.com/weppos/publicsuffix-go/publicsuffix"

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
	r.AddFromFiles("dashboard", "templates/base.html", "templates/dashboard.html")
	r.AddFromFiles("forbidden", "templates/base.html", "templates/forbidden.html")
	s.Router.HTMLRender = r
}

func (s *Server) buildRoutes() {
	ginLogger := log.New()
	s.Router.Use(ginlogrus.Logger(ginLogger), gin.Recovery())

	s.Router.GET("/access", s.accessHandler())

	s.Router.Static("/static", "static")
	s.Router.GET("/", s.dashboardUIHandler())
	s.Router.GET("/login", s.loginUIHandler())
	s.Router.POST("/login", s.loginHandler())
	s.Router.GET("/forbidden", s.forbiddenUIHandler())
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
			c.Redirect(http.StatusFound, s.getRedirectURL(*c.Request.URL, "/login", &rawURL, nil))
		} else if isBrowser && user != nil {
			c.Redirect(http.StatusFound, s.getRedirectURL(*c.Request.URL, "/forbidden", &rawURL, nil))
		} else {
			c.String(http.StatusUnauthorized, "No access granted")
		}
	}
}

func (s *Server) getRedirectURL(reqURL url.URL, path string, origURL, errVal *string) string {
	redirectURL := reqURL
	redirectURL.Path = path
	q := url.Values{}
	if origURL != nil {
		q.Set("orig_url", *origURL)
	}
	if errVal != nil {
		q.Set("error", *errVal)
	}
	redirectURL.RawQuery = q.Encode()

	return redirectURL.String()
}

func (s *Server) loginHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		log.WithField("postform", c.Request.PostForm).Info("Post form")

		username := c.PostForm("username")
		password := c.PostForm("password")

		user, err := manager.GetAuthManager().ValidateCredentials(username, password)
		if err != nil {
			errVal := "incorrect"
			c.Redirect(http.StatusFound, s.getRedirectURL(*c.Request.URL, "/login", nil, &errVal))
			return
		}
		token, err := manager.GetAuthManager().CreateUserToken(user.ID, false)
		if err != nil {
			errVal := "server"
			c.Redirect(http.StatusFound, s.getRedirectURL(*c.Request.URL, "/login", nil, &errVal))
			return
		}

		maxAge := int(token.ExpiresAt.Sub(time.Now()).Seconds())
		host, err := publicsuffix.Domain(c.Request.Host)
		if err != nil {
			log.WithFields(log.Fields{"host": c.Request.Host, "err": err}).Error("Could not determine domain, use host instead")
			host = c.Request.Host
		}
		c.SetCookie("tac_token", token.Token, maxAge, "", host, false, true)

		c.Redirect(http.StatusFound, s.getRedirectURL(*c.Request.URL, "/", nil, nil))
	}
}

func (s *Server) dashboardUIHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.HTML(http.StatusOK, "dashboard", nil)
	}
}

func (s *Server) loginUIHandler() gin.HandlerFunc {
	return func(c *gin.Context) {

		c.HTML(http.StatusOK, "login", gin.H{
			"error": c.Query("error"),
		})
	}
}

func (s *Server) forbiddenUIHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.HTML(http.StatusForbidden, "forbidden", nil)
	}
}
