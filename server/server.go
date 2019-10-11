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

const (
	errorParam       = "error"
	redirectURLParam = "redirect_url"
	userContextKey   = "user"
)

type Server struct {
	Router     *gin.Engine
	cookieName string
}

func NewServer(cookieName string) *Server {
	s := &Server{
		Router:     gin.Default(),
		cookieName: cookieName,
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

	s.Router.Static("/static", "./static")
	s.Router.GET("/", s.fillUserFromCookie(), s.userMustBeValid(), s.dashboardUIHandler())
	s.Router.GET("/login", s.fillUserFromCookie(), s.loginUIHandler())
	s.Router.POST("/login", s.loginHandler())
	s.Router.GET("/forbidden", s.fillUserFromCookie(), s.forbiddenUIHandler())
}

func (s *Server) accessHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		host := c.Request.Header.Get("X-Forwarded-Host")
		path := c.Request.Header.Get("X-Forwarded-Uri")
		proto := c.Request.Header.Get("X-Forwarded-Proto")
		requestLogger := log.WithFields(log.Fields{"host": host, "path": path, "proto": proto})

		if host == "" || path == "" || proto == "" {
			c.String(http.StatusBadRequest, "Not all needed headers present")
			return
		}
		completeURL, err := url.Parse(host + path)
		if err != nil {
			c.String(http.StatusBadRequest, "Could not parse host as URL")
			return
		}
		completeURL.Scheme = proto

		var accessGranted = false
		var user *models.User
		var completeURLString = completeURL.String()

		if cookie, err := c.Request.Cookie(s.cookieName); err == nil {
			requestLogger.Info("Cookie request")

			accessGranted, user, err = manager.GetAccessManager().CheckAccessToken(host, path, cookie.Value, false, requestLogger)
		} else if username, password, hasAuth := c.Request.BasicAuth(); hasAuth {
			requestLogger.WithField("username", username).Info("BasicAuth request")

			accessGranted, user, err = manager.GetAccessManager().CheckAccessCredentials(host, path, username, password, true, requestLogger)
		} else if bearer := c.Request.Header.Get("Authorization"); bearer != "" {
			requestLogger.Info("Bearer request")

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
			c.Redirect(http.StatusFound, s.getRedirectURL(*c.Request.URL, "/login", &completeURLString, nil))
		} else if isBrowser && user != nil {
			c.Redirect(http.StatusFound, s.getRedirectURL(*c.Request.URL, "/forbidden", &completeURLString, nil))
		} else {
			c.String(http.StatusUnauthorized, "No access granted")
		}
	}
}

func (s *Server) getRedirectURL(reqURL url.URL, path string, origURL, errVal *string) string {
	redirectURL := reqURL
	redirectURL.Path = path
	q := redirectURL.Query()
	if origURL != nil {
		q.Set(redirectURLParam, *origURL)
	}
	if errVal != nil {
		q.Set(errorParam, *errVal)
	}
	redirectURL.RawQuery = q.Encode()

	return redirectURL.String()
}

func (s *Server) fillUserFromCookie() gin.HandlerFunc {
	return func(c *gin.Context) {
		cookie, err := c.Request.Cookie(s.cookieName)
		if err != nil {
			c.Set(userContextKey, nil)
			c.Next()
			return
		}
		user, err := manager.GetAuthManager().ValidateTokenString(cookie.Value, false)
		if err != nil {
			c.Set(userContextKey, nil)
			c.Next()
			return
		}

		c.Set(userContextKey, user)
		c.Next()
	}
}

func (s *Server) userMustBeValid() gin.HandlerFunc {
	return func(c *gin.Context) {
		if user, ok := c.Get(userContextKey); ok == false || user == nil {
			c.Redirect(http.StatusFound, s.getRedirectURL(*c.Request.URL, "/login", nil, nil))
			return
		}
		c.Next()
	}
}

func (s *Server) loginHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
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
		c.SetCookie(s.cookieName, token.Token, maxAge, "", host, false, true)

		redirect := c.Query(redirectURLParam)
		if redirect != "" {
			c.Redirect(http.StatusFound, redirect)
		} else {
			c.Redirect(http.StatusFound, s.getRedirectURL(*c.Request.URL, "/", nil, nil))
		}
	}
}

func (s *Server) dashboardUIHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		userInt, _ := c.Get(userContextKey)
		user, ok := userInt.(*models.User)
		if !ok {
			c.HTML(http.StatusInternalServerError, "dashboard", gin.H{"error": "server"})
			return
		}

		rawSiteMappings, err := manager.GetSiteManager().GetSiteMappingsByUser(user)
		if err != nil {
			c.HTML(http.StatusInternalServerError, "dashboard", gin.H{"error": "server"})
			return
		}

		siteMappings := make([]struct {
			SiteMapping *models.SiteMapping
			Site        *models.Site
		}, len(rawSiteMappings))

		for it, rawSiteMapping := range rawSiteMappings {
			site, err := manager.GetSiteManager().GetSiteByID(rawSiteMapping.SiteID)
			if err != nil {
				continue
			}
			siteMappings[it].SiteMapping = rawSiteMapping
			siteMappings[it].Site = site
		}

		c.HTML(http.StatusOK, "dashboard", gin.H{
			"siteMappings": siteMappings,
		})
	}
}

func (s *Server) loginUIHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		if user, ok := c.Get(userContextKey); ok == true && user != nil {
			c.Redirect(http.StatusFound, s.getRedirectURL(*c.Request.URL, "/", nil, nil))
			return
		}

		c.HTML(http.StatusOK, "login", gin.H{
			"error": c.Query(errorParam),
		})
	}
}

func (s *Server) forbiddenUIHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.HTML(http.StatusForbidden, "forbidden", nil)
	}
}
div class="col s2"