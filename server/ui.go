package server

import (
	"TraefikAccessControl/manager"
	"TraefikAccessControl/models"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-contrib/multitemplate"
	"github.com/gin-gonic/gin"
)

func (s *Server) parseTemplates() {
	r := multitemplate.NewRenderer()
	r.AddFromFiles("login", "templates/base.html", "templates/login.html")
	r.AddFromFiles("site", "templates/base.html", "templates/site.html")
	r.AddFromFiles("dashboard", "templates/base.html", "templates/dashboard.html")
	r.AddFromFiles("forbidden", "templates/base.html", "templates/forbidden.html")
	r.AddFromFiles("notfound", "templates/base.html", "templates/notfound.html")
	s.Router.HTMLRender = r
}

func (s *Server) buildUIRoutes() {
	s.Router.GET("/", s.fillUserFromCookie(), s.userMustBeValid(false, false), s.dashboardUIHandler())
	s.Router.GET("/site/:id", s.fillUserFromCookie(), s.userMustBeValid(false, true), s.siteUIHandler())
	s.Router.GET("/login", s.fillUserFromCookie(), s.loginUIHandler())
	s.Router.GET("/forbidden", s.fillUserFromCookie(), s.forbiddenUIHandler())
	s.Router.NoRoute(s.notfoundUIHandler())
}

func (s *Server) dashboardUIHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		userInt, _ := c.Get(userContextKey)
		user, ok := userInt.(*models.User)
		if !ok {
			c.HTML(http.StatusInternalServerError, "dashboard", gin.H{"error": errorServer})
			return
		}

		rawSiteMappings, err := manager.GetSiteManager().GetSiteMappingsByUser(user)
		if err != nil {
			c.HTML(http.StatusInternalServerError, "dashboard", gin.H{"error": errorServer})
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

		tokens, err := manager.GetAuthManager().GetBearerTokens(user)
		for _, token := range tokens {
			if time.Now().After(token.CreatedAt.Add(2 * time.Minute)) {
				token.Token = token.Token[0:4] + strings.Repeat("×", len(token.Token)-4)
			}
		}

		users, err := manager.GetAuthManager().GetAllUsers()
		if err != nil {
			c.HTML(http.StatusInternalServerError, "dashboard", gin.H{"error": errorServer})
			return
		}

		sites, err := manager.GetSiteManager().GetAllSites()
		if err != nil {
			c.HTML(http.StatusInternalServerError, "dashboard", gin.H{"error": errorServer})
			return
		}

		c.HTML(http.StatusOK, "dashboard", gin.H{
			"error":        c.Query(errorURLParam),
			"success":      c.Query(successURLParam),
			"user":         user,
			"siteMappings": siteMappings,
			"tokens":       tokens,
			"users":        users,
			"sites":        sites,
		})
	}
}

func (s *Server) siteUIHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		userInt, _ := c.Get(userContextKey)
		user, ok := userInt.(*models.User)
		if !ok {
			c.HTML(http.StatusInternalServerError, "dashboard", gin.H{"error": errorServer})
			return
		}

		siteID, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			c.Redirect(http.StatusFound, s.getRedirectURL(*c.Request.URL, "/", nil, &errorSite))
			return
		}

		site, err := manager.GetSiteManager().GetSiteByID(siteID)
		if err != nil {
			c.Redirect(http.StatusFound, s.getRedirectURL(*c.Request.URL, "/", nil, &errorSite))
			return
		}

		rawSiteMappings, err := manager.GetSiteManager().GetSiteMappingsBySite(site)
		if err != nil {
			c.Redirect(http.StatusFound, s.getRedirectURL(*c.Request.URL, "/", nil, &errorServer))
			return
		}

		siteMappings := make([]struct {
			SiteMapping *models.SiteMapping
			User        *models.User
		}, len(rawSiteMappings))

		for it, rawSiteMapping := range rawSiteMappings {
			user, err := manager.GetAuthManager().GetUserByID(rawSiteMapping.UserID)
			if err != nil {
				continue
			}
			siteMappings[it].SiteMapping = rawSiteMapping
			siteMappings[it].User = user
		}

		allUsers, err := manager.GetAuthManager().GetAllUsers()
		if err != nil {
			c.Redirect(http.StatusFound, s.getRedirectURL(*c.Request.URL, "/", nil, &errorServer))
			return
		}
		availUsers := make([]*models.User, 0)
		for _, user := range allUsers {
			found := false
			for _, siteMapping := range siteMappings {
				if user.ID == siteMapping.User.ID {
					found = true
					break
				}
			}
			if !found {
				availUsers = append(availUsers, user)
			}
		}

		c.HTML(http.StatusOK, "site", gin.H{
			"error":        c.Query(errorURLParam),
			"success":      c.Query(successURLParam),
			"user":         user,
			"site":         site,
			"siteMappings": siteMappings,
			"availUsers":   availUsers,
		})
	}
}

func (s *Server) notfoundUIHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		if strings.Contains(c.Request.UserAgent(), "Mozilla") {
			c.HTML(http.StatusNotFound, "notfound", gin.H{
				"error":   c.Query(errorURLParam),
				"success": c.Query(successURLParam),
			})
		} else {
			c.String(http.StatusNotFound, "Not Found", gin.H{})
		}
	}
}

func (s *Server) loginUIHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		if user, ok := c.Get(userContextKey); ok == true && user != nil {
			c.Redirect(http.StatusFound, s.getRedirectURL(*c.Request.URL, "/", nil, nil))
			return
		}

		c.HTML(http.StatusOK, "login", gin.H{
			"error":   c.Query(errorURLParam),
			"success": c.Query(successURLParam),
		})
	}
}

func (s *Server) forbiddenUIHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		userInt, _ := c.Get(userContextKey)
		user := userInt.(*models.User)

		c.HTML(http.StatusForbidden, "forbidden", gin.H{
			"error":   c.Query(errorURLParam),
			"success": c.Query(successURLParam),
			"user":    user,
		})
	}
}
