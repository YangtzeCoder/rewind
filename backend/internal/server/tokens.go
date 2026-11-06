package server

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"math"
	"strconv"
	"strings"
	"time"

	"rewind/internal/auth"
	"rewind/internal/database"
	"rewind/internal/id"

	"github.com/gin-gonic/gin"
)

type tokenHandler struct{ store *database.TokenStore }

type tokenResponse struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	CreatedAt string `json:"created_at"`
}

func registerTokenRoutes(router *gin.RouterGroup, authService *auth.Service, store *database.TokenStore) {
	h := &tokenHandler{store: store}
	group := router.Group("/tokens")
	group.Use((&authHandler{service: authService}).authenticate)
	group.POST("", h.create)
	group.GET("", h.list)
	group.DELETE("/:token_id", h.delete)
}

func (h *tokenHandler) create(c *gin.Context) {
	var input struct {
		Name string `json:"name" binding:"required"`
	}
	if c.ShouldBindJSON(&input) != nil || strings.TrimSpace(input.Name) == "" {
		failure(c, errInvalidArgument, "name is required")
		return
	}
	tokenID, err := id.APIToken()
	if err != nil {
		internalError(c)
		return
	}
	var secret [32]byte
	if _, err := rand.Read(secret[:]); err != nil {
		internalError(c)
		return
	}
	credential := "api_" + base64.RawURLEncoding.EncodeToString(secret[:])
	hash := sha256.Sum256([]byte(credential))
	now := time.Now().UTC().Truncate(time.Millisecond)
	token := database.APIToken{ID: tokenID, Name: input.Name, CreatedAt: now.UnixMilli()}
	if err := h.store.Create(c.Request.Context(), token, hex.EncodeToString(hash[:])); err != nil {
		internalError(c)
		return
	}
	success(c, gin.H{"id": token.ID, "name": token.Name, "created_at": now.Format(time.RFC3339Nano), "token": credential})
}

func (h *tokenHandler) list(c *gin.Context) {
	page, err := paginationValue(c, "page", 1, math.MaxInt64)
	if err != nil {
		failure(c, errInvalidArgument, "invalid pagination parameters")
		return
	}
	size, err := paginationValue(c, "page_size", 20, 100)
	if err != nil || page-1 > math.MaxInt64/size {
		failure(c, errInvalidArgument, "invalid pagination parameters")
		return
	}
	tokens, err := h.store.List(c.Request.Context(), size, (page-1)*size)
	if err != nil {
		internalError(c)
		return
	}
	result := make([]tokenResponse, 0, len(tokens))
	for _, token := range tokens {
		result = append(result, tokenResponse{ID: token.ID, Name: token.Name, CreatedAt: time.UnixMilli(token.CreatedAt).UTC().Format(time.RFC3339Nano)})
	}
	success(c, result)
}

func paginationValue(c *gin.Context, name string, fallback, max int64) (int64, error) {
	values, exists := c.Request.URL.Query()[name]
	if !exists {
		return fallback, nil
	}
	if len(values) != 1 {
		return 0, strconv.ErrSyntax
	}
	value, err := strconv.ParseInt(values[0], 10, 64)
	if err != nil || value < 1 || value > max {
		return 0, strconv.ErrSyntax
	}
	return value, nil
}

func (h *tokenHandler) delete(c *gin.Context) {
	deleted, err := h.store.Delete(c.Request.Context(), c.Param("token_id"))
	if err != nil {
		internalError(c)
		return
	}
	if !deleted {
		failure(c, errNotFound, "token not found")
		return
	}
	success(c, nil)
}
