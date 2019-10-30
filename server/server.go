package server

import (
	"TraefikAccessControl/manager"
	"TraefikAccessControl/models"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/weppos/publicsuffix-go/publicsuffix"

	"github.com/gin-gonic/gin"
)

const (
	errorURLParam    = "error"
	successURLParam  = "success"
	redirectURLParam = "redirect_url"

	userContextKey = "user"
)

var (
	errorServer               = "Internal Server Error"
	errorSite                 = "Site could not be found"
	errorIncorrectCredentials = "Username or password incorrect"

	successLogout = "Successfully logged out"
)

type Server struct {
	Router         *gin.Engine
	cookieName     string
	userHeaderName string
}

func NewServer(cookieName, userHeaderName string) *Server {
	s := &Server{
		Router:         gin.New(),
		cookieName:     cookieName,
		userHeaderName: userHeaderName,
	}
	s.parseTemplates()
	s.buildRoutes()
	return s
}

func (s *Server) buildRoutes() {
	s.Router.Use(gin.Recovery())

	s.buildCoreRoutes()
	s.buildUIRoutes()
	s.buildAPIRoutes()
}

func (s *Server) buildCoreRoutes() {
	s.Router.GET("/access", s.accessHandler())
	s.Router.Static("/static", "./static")
	s.Router.POST("/login", s.loginHandler())
	s.Router.GET("/logout", s.logoutHandler())
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

		if header := c.Request.Header.Get(manager.CheckConfigHeader); header != "" {
			c.Header(manager.CheckConfigHeader, manager.CheckConfigHeaderContent)
			c.Status(http.StatusUnauthorized)
			return
		}

		var accessGranted = false
		var promptBasicAuth = false
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

			accessGranted, promptBasicAuth, err = manager.GetAccessManager().CheckAnonymousAccess(host, path, requestLogger)
		}

		isBrowser := strings.Contains(c.Request.UserAgent(), "Mozilla")
		if accessGranted {
			if user != nil {
				c.Header("X-TAC-User", user.Username)
			}
			c.Status(http.StatusOK)
		} else if promptBasicAuth {
			c.Header("WWW-Authenticate", "Basic realm="+strconv.Quote("Authorization required"))
			c.Status(http.StatusUnauthorized)
		} else if isBrowser && user == nil {
			c.Redirect(http.StatusFound, s.getRedirectURL(*c.Request.URL, "/login", &completeURLString, nil, nil))
		} else if isBrowser && user != nil {
			c.Redirect(http.StatusFound, s.getRedirectURL(*c.Request.URL, "/forbidden", &completeURLString, nil, nil))
		} else {
			c.String(http.StatusUnauthorized, "No access granted")
		}
	}
}

func (s *Server) getRedirectURL(reqURL url.URL, path string, origURL, errVal, successVal *string) string {
	redirectURL := reqURL
	redirectURL.Path = path
	q := redirectURL.Query()
	if origURL != nil {
		q.Set(redirectURLParam, *origURL)
	} else {
		q.Del(redirectURLParam)
	}
	if errVal != nil {
		q.Set(errorURLParam, *errVal)
	} else {
		q.Del(errorURLParam)
	}
	if successVal != nil {
		q.Set(successURLParam, *successVal)
	} else {
		q.Del(successURLParam)
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

func (s *Server) userMustBeValid(forAPI bool, hasToBeAdmin bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		user, ok := c.Get(userContextKey)
		if ok == false || user == nil {
			if forAPI {
				c.String(http.StatusUnauthorized, "Not authenticated")
			} else {
				c.Redirect(http.StatusFound, s.getRedirectURL(*c.Request.URL, "/login", nil, nil, nil))
			}
			return
		}
		if hasToBeAdmin && !user.(*models.User).IsAdmin {
			if forAPI {
				c.String(http.StatusForbidden, "Not an admin")
			} else {
				c.Redirect(http.StatusFound, s.getRedirectURL(*c.Request.URL, "/forbidden", nil, nil, nil))
			}
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
			c.Redirect(http.StatusFound, s.getRedirectURL(*c.Request.URL, "/login", nil, &errorIncorrectCredentials, nil))
			return
		}
		token, err := manager.GetAuthManager().CreateUserToken(user.ID, nil)
		if err != nil {
			c.Redirect(http.StatusFound, s.getRedirectURL(*c.Request.URL, "/login", nil, &errorServer, nil))
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
			c.Redirect(http.StatusFound, s.getRedirectURL(*c.Request.URL, "/", nil, nil, nil))
		}
	}
}

func (s *Server) logoutHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		cookie, err := c.Request.Cookie(s.cookieName)
		if err == nil {
			manager.GetAuthManager().DeleteToken(cookie.Value)
		}
		c.SetCookie(s.cookieName, "", -1, "", "", false, true)
		c.Redirect(http.StatusFound, s.getRedirectURL(*c.Request.URL, "/login", nil, nil, &successLogout))
	}
}
