package server

import (
	"TraefikAccessControl/manager"
	"TraefikAccessControl/models"
	"net/http"

	"github.com/gin-gonic/gin"
)

func (s *Server) buildAPIRoutes() {
	api := s.Router.Group("/api", s.fillUserFromCookie(), s.userMustBeValid(true, false))
	api.POST("/user", s.createUserAPIHandler())
	api.PUT("/user", s.updateUserAPIHandler())
	api.DELETE("/user", s.deleteUserAPIHandler())
	api.POST("/site", s.createSiteAPIHandler())
	api.PUT("/site", s.updateSiteAPIHandler())
	api.DELETE("/site", s.deleteSiteAPIHandler())
	api.POST("/bearer", s.createBearerAPIHandler())
	api.DELETE("/bearer", s.deleteBearerAPIHandler())
	api.POST("/mapping", s.createMappingAPIHandler())
	api.DELETE("/mapping", s.deleteMappingAPIHandler())
}

func (s *Server) createUserAPIHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		userInt, _ := c.Get(userContextKey)
		user := userInt.(*models.User)

		if !user.IsAdmin {
			c.JSON(http.StatusForbidden, gin.H{"error": "Not authorized to create user"})
			return
		}

		var userIn models.User
		if err := c.ShouldBindJSON(&userIn); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		err := manager.GetAuthManager().CreateUser(&userIn)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{})
	}
}

func (s *Server) updateUserAPIHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		userInt, _ := c.Get(userContextKey)
		user := userInt.(*models.User)

		var userIn models.User
		if err := c.ShouldBindJSON(&userIn); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		err := manager.GetAuthManager().UpdateUserPassword(user, userIn.Password)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{})
	}
}

func (s *Server) deleteUserAPIHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		userInt, _ := c.Get(userContextKey)
		user := userInt.(*models.User)

		if !user.IsAdmin {
			c.JSON(http.StatusForbidden, gin.H{"error": "Not authorized to delete user"})
			return
		}

		var userIn models.User
		if err := c.ShouldBindJSON(&userIn); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		err := manager.GetAuthManager().DeleteUser(userIn.ID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		err = manager.GetSiteManager().DeleteUserMappings(userIn.ID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{})
	}
}

func (s *Server) createSiteAPIHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		userInt, _ := c.Get(userContextKey)
		user := userInt.(*models.User)

		if !user.IsAdmin {
			c.JSON(http.StatusForbidden, gin.H{"error": "Not authorized to create site"})
			return
		}

		var siteIn models.Site
		if err := c.ShouldBindJSON(&siteIn); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		err := manager.GetSiteManager().CreateSite(&siteIn)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{})
	}
}

func (s *Server) updateSiteAPIHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		userInt, _ := c.Get(userContextKey)
		user := userInt.(*models.User)

		if !user.IsAdmin {
			c.JSON(http.StatusForbidden, gin.H{"error": "Not authorized to delete site"})
			return
		}
		var siteIn models.Site
		if err := c.ShouldBindJSON(&siteIn); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		err := manager.GetSiteManager().UpdateSite(&siteIn)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{})
	}
}

func (s *Server) deleteSiteAPIHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		userInt, _ := c.Get(userContextKey)
		user := userInt.(*models.User)

		if !user.IsAdmin {
			c.JSON(http.StatusForbidden, gin.H{"error": "Not authorized to delete site"})
			return
		}
		var siteIn models.Site
		if err := c.ShouldBindJSON(&siteIn); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		err := manager.GetSiteManager().DeleteSite(siteIn.ID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{})
	}
}

func (s *Server) createBearerAPIHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		userInt, _ := c.Get(userContextKey)
		user := userInt.(*models.User)

		var tokenIn models.Token
		if err := c.ShouldBindJSON(&tokenIn); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		token, err := manager.GetAuthManager().CreateUserToken(user.ID, tokenIn.Name)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, token)
	}
}

func (s *Server) deleteBearerAPIHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		userInt, _ := c.Get(userContextKey)
		user := userInt.(*models.User)

		var tokenIn models.Token
		if err := c.ShouldBindJSON(&tokenIn); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		err := manager.GetAuthManager().DeleteTokenByName(user.ID, *tokenIn.Name)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{})
	}
}

func (s *Server) createMappingAPIHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		userInt, _ := c.Get(userContextKey)
		user := userInt.(*models.User)

		if !user.IsAdmin {
			c.JSON(http.StatusForbidden, gin.H{"error": "Not authorized to create mapping"})
			return
		}
		var mappingIn models.SiteMapping
		if err := c.ShouldBindJSON(&mappingIn); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		err := manager.GetSiteManager().CreateSiteMapping(&mappingIn)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{})
	}
}

func (s *Server) deleteMappingAPIHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		userInt, _ := c.Get(userContextKey)
		user := userInt.(*models.User)

		if !user.IsAdmin {
			c.JSON(http.StatusForbidden, gin.H{"error": "Not authorized to create mapping"})
			return
		}
		var mappingIn models.SiteMapping
		if err := c.ShouldBindJSON(&mappingIn); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		err := manager.GetSiteManager().DeleteSiteMapping(mappingIn.UserID, mappingIn.SiteID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{})
	}
}
