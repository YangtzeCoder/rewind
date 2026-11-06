package server

import (
	"encoding/json"
	"errors"
	"math"
	"strings"
	"time"

	"rewind/internal/auth"
	"rewind/internal/database"
	"rewind/internal/id"

	"github.com/gin-gonic/gin"
)

type projectHandler struct{ store *database.ProjectStore }

type projectResponse struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}

func projectJSON(p database.Project) projectResponse {
	return projectResponse{p.ID, p.Name, p.Description, time.UnixMilli(p.CreatedAt).UTC().Format(time.RFC3339Nano), time.UnixMilli(p.UpdatedAt).UTC().Format(time.RFC3339Nano)}
}

func registerProjectRoutes(router *gin.RouterGroup, authService *auth.Service, store *database.ProjectStore) {
	h := &projectHandler{store: store}
	group := router.Group("/projects")
	group.Use((&authHandler{service: authService}).authenticate)
	group.POST("", h.create)
	group.GET("", h.list)
	group.GET("/:project_id", h.get)
	group.PATCH("/:project_id", h.update)
	group.DELETE("/:project_id", h.delete)
}

// projectInput distinguishes absent fields from explicit null values.
func projectInput(c *gin.Context) (name, description *string, ok bool) {
	var input map[string]json.RawMessage
	if c.ShouldBindJSON(&input) != nil || input == nil {
		return nil, nil, false
	}
	for key, target := range map[string]**string{"name": &name, "description": &description} {
		if raw, exists := input[key]; exists {
			var value string
			if string(raw) == "null" || json.Unmarshal(raw, &value) != nil {
				return nil, nil, false
			}
			*target = &value
		}
	}
	if name != nil && strings.TrimSpace(*name) == "" {
		return nil, nil, false
	}
	return name, description, true
}

func (h *projectHandler) create(c *gin.Context) {
	name, description, ok := projectInput(c)
	if !ok || name == nil {
		failure(c, errInvalidArgument, "name is required and fields must be valid strings")
		return
	}
	projectID, err := id.Project()
	if err != nil {
		internalError(c)
		return
	}
	now := time.Now().UTC().UnixMilli()
	p := database.Project{ID: projectID, Name: *name, CreatedAt: now, UpdatedAt: now}
	if description != nil {
		p.Description = *description
	}
	if err := h.store.Create(c.Request.Context(), p); err != nil {
		internalError(c)
		return
	}
	success(c, projectJSON(p))
}

func (h *projectHandler) list(c *gin.Context) {
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
	projects, err := h.store.List(c.Request.Context(), size, (page-1)*size)
	if err != nil {
		internalError(c)
		return
	}
	result := make([]projectResponse, 0, len(projects))
	for _, p := range projects {
		result = append(result, projectJSON(p))
	}
	success(c, result)
}

func (h *projectHandler) get(c *gin.Context) {
	p, err := h.store.Get(c.Request.Context(), c.Param("project_id"))
	if err != nil {
		projectError(c, err)
		return
	}
	success(c, projectJSON(p))
}

func (h *projectHandler) update(c *gin.Context) {
	name, description, ok := projectInput(c)
	if !ok || (name == nil && description == nil) {
		failure(c, errInvalidArgument, "at least one valid field is required")
		return
	}
	p, err := h.store.Update(c.Request.Context(), c.Param("project_id"), name, description, time.Now().UTC().UnixMilli())
	if err != nil {
		projectError(c, err)
		return
	}
	success(c, projectJSON(p))
}

func (h *projectHandler) delete(c *gin.Context) {
	deleted, err := h.store.Delete(c.Request.Context(), c.Param("project_id"))
	if err != nil {
		internalError(c)
		return
	}
	if !deleted {
		failure(c, errNotFound, "project not found")
		return
	}
	success(c, nil)
}

func projectError(c *gin.Context, err error) {
	if errors.Is(err, database.ErrNotFound) {
		failure(c, errNotFound, "project not found")
		return
	}
	internalError(c)
}
