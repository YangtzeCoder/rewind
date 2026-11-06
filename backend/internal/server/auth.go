package server

import (
	"errors"
	"strings"
	"time"

	"rewind/internal/auth"

	"github.com/gin-gonic/gin"
)

func registerAuthRoutes(router *gin.RouterGroup, service *auth.Service) {
	h := &authHandler{service: service}
	group := router.Group("/auth")
	group.POST("/login", h.login)
	secured := group.Group("")
	secured.Use(h.authenticate)
	secured.POST("/logout", h.logout)
	secured.POST("/password", h.password)
	secured.GET("/me", h.me)
}

type authHandler struct{ service *auth.Service }

func (h *authHandler) login(c *gin.Context) {
	var input struct {
		Username string `json:"username" binding:"required"`
		Password string `json:"password" binding:"required"`
	}
	if c.ShouldBindJSON(&input) != nil {
		failure(c, errInvalidArgument, "username and password are required")
		return
	}
	token, expires, err := h.service.Login(c.Request.Context(), input.Username, input.Password)
	if err != nil {
		authError(c, err, errInvalidCredentials, "invalid credentials")
		return
	}
	success(c, gin.H{"token": token, "expires_at": expires.Format(time.RFC3339Nano)})
}

func (h *authHandler) authenticate(c *gin.Context) {
	parts := strings.Fields(c.GetHeader("Authorization"))
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		failure(c, errUnauthorized, "unauthorized")
		c.Abort()
		return
	}
	if err := h.service.Authenticate(c.Request.Context(), parts[1]); err != nil {
		authError(c, err, errUnauthorized, "unauthorized")
		c.Abort()
		return
	}
	c.Set("login_token", parts[1])
	c.Next()
}

func (h *authHandler) logout(c *gin.Context) {
	if err := h.service.Logout(c.Request.Context(), c.GetString("login_token")); err != nil {
		internalError(c)
		return
	}
	success(c, nil)
}

func (h *authHandler) password(c *gin.Context) {
	var input struct {
		Current string `json:"current_password" binding:"required"`
		New     string `json:"new_password" binding:"required"`
	}
	if c.ShouldBindJSON(&input) != nil || len(input.New) > 72 {
		failure(c, errInvalidArgument, "current_password and new_password are required; new_password must be at most 72 bytes")
		return
	}
	if err := h.service.ChangePassword(c.Request.Context(), input.Current, input.New); err != nil {
		authError(c, err, errInvalidCredentials, "invalid credentials")
		return
	}
	success(c, nil)
}

func (h *authHandler) me(c *gin.Context) {
	username, err := h.service.Username(c.Request.Context())
	if err != nil {
		internalError(c)
		return
	}
	success(c, gin.H{"username": username})
}

func authError(c *gin.Context, err error, code, message string) {
	if errors.Is(err, auth.ErrUnauthorized) {
		failure(c, code, message)
		return
	}
	internalError(c)
}

func internalError(c *gin.Context) {
	failure(c, errInternal, "internal server error")
}
