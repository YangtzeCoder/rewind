package server

import (
	"context"
	"errors"
	"math"
	"strings"
	"time"

	"rewind/internal/auth"
	"rewind/internal/database"
	"rewind/internal/id"
	"rewind/internal/storage"

	"github.com/gin-gonic/gin"
)

type projectHandler struct {
	store *database.ProjectStore
	files storage.Store
}

func registerProjectRoutes(router *gin.RouterGroup, authService *auth.Service, store *database.ProjectStore, files storage.Store) {
	h := &projectHandler{store: store, files: files}
	group := router.Group("/projects")
	group.Use((&authHandler{service: authService}).authenticate)
	group.POST("", h.create)
	group.GET("", h.list)
	group.GET("/:project_id", h.get)
	group.GET("/:project_id/files", h.listFiles)
	group.PATCH("/:project_id", h.update)
	group.DELETE("/:project_id", h.delete)
}

func (h *projectHandler) listFiles(c *gin.Context) {
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
	filters := c.Request.URL.Query()["key"]
	if len(filters) > 1 {
		failure(c, errInvalidArgument, "key must be specified at most once")
		return
	}
	filter := ""
	if len(filters) == 1 {
		filter = filters[0]
	}
	files, err := h.store.ListFiles(c.Request.Context(), c.Param("project_id"), filter, size, (page-1)*size)
	if err != nil {
		projectError(c, err)
		return
	}
	type latestVersionResponse struct {
		VersionID string `json:"version_id"`
		Size      int64  `json:"size"`
		CreatedAt string `json:"created_at"`
	}
	type fileResponse struct {
		ID            string                 `json:"id"`
		Key           string                 `json:"key"`
		CreatedAt     string                 `json:"created_at"`
		UpdatedAt     string                 `json:"updated_at"`
		VersionCount  int64                  `json:"version_count"`
		LatestVersion *latestVersionResponse `json:"latest_version"`
	}
	result := make([]fileResponse, 0, len(files))
	for _, f := range files {
		item := fileResponse{ID: f.ID, Key: f.Key, CreatedAt: time.UnixMilli(f.CreatedAt).UTC().Format(time.RFC3339Nano),
			UpdatedAt: time.UnixMilli(f.UpdatedAt).UTC().Format(time.RFC3339Nano), VersionCount: f.VersionCount}
		if f.LatestVersionID.Valid {
			item.LatestVersion = &latestVersionResponse{VersionID: f.LatestVersionID.String, Size: f.LatestSize.Int64,
				CreatedAt: time.UnixMilli(f.LatestTime.Int64).UTC().Format(time.RFC3339Nano)}
		}
		result = append(result, item)
	}
	success(c, result)
}

func (h *projectHandler) create(c *gin.Context) {
	type createProjectRequest struct {
		Name        *string `json:"name" binding:"required"`
		Description *string `json:"description"`
	}
	var input createProjectRequest
	if c.ShouldBindJSON(&input) != nil {
		failure(c, errInvalidArgument, "name is required and fields must be valid strings")
		return
	}
	if strings.TrimSpace(*input.Name) == "" {
		failure(c, errInvalidArgument, "name is required and fields must be valid strings")
		return
	}
	projectID, err := id.Project()
	if err != nil {
		internalError(c)
		return
	}
	now := time.Now().UTC().UnixMilli()
	p := database.Project{ID: projectID, Name: *input.Name, CreatedAt: now, UpdatedAt: now}
	if input.Description != nil {
		p.Description = *input.Description
	}
	if err := h.store.Create(c.Request.Context(), p); err != nil {
		internalError(c)
		return
	}
	type createProjectResponse struct {
		ID          string `json:"id"`
		Name        string `json:"name"`
		Description string `json:"description"`
		CreatedAt   string `json:"created_at"`
		UpdatedAt   string `json:"updated_at"`
	}
	success(c, createProjectResponse{ID: p.ID, Name: p.Name, Description: p.Description,
		CreatedAt: time.UnixMilli(p.CreatedAt).UTC().Format(time.RFC3339Nano),
		UpdatedAt: time.UnixMilli(p.UpdatedAt).UTC().Format(time.RFC3339Nano)})
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
	type listProjectResponse struct {
		ID          string `json:"id"`
		Name        string `json:"name"`
		Description string `json:"description"`
		CreatedAt   string `json:"created_at"`
		UpdatedAt   string `json:"updated_at"`
	}
	result := make([]listProjectResponse, 0, len(projects))
	for _, p := range projects {
		result = append(result, listProjectResponse{ID: p.ID, Name: p.Name, Description: p.Description,
			CreatedAt: time.UnixMilli(p.CreatedAt).UTC().Format(time.RFC3339Nano),
			UpdatedAt: time.UnixMilli(p.UpdatedAt).UTC().Format(time.RFC3339Nano)})
	}
	success(c, result)
}

func (h *projectHandler) get(c *gin.Context) {
	p, err := h.store.Get(c.Request.Context(), c.Param("project_id"))
	if err != nil {
		projectError(c, err)
		return
	}
	type getProjectResponse struct {
		ID          string `json:"id"`
		Name        string `json:"name"`
		Description string `json:"description"`
		CreatedAt   string `json:"created_at"`
		UpdatedAt   string `json:"updated_at"`
	}
	success(c, getProjectResponse{ID: p.ID, Name: p.Name, Description: p.Description,
		CreatedAt: time.UnixMilli(p.CreatedAt).UTC().Format(time.RFC3339Nano),
		UpdatedAt: time.UnixMilli(p.UpdatedAt).UTC().Format(time.RFC3339Nano)})
}

func (h *projectHandler) update(c *gin.Context) {
	type updateProjectRequest struct {
		Name        *string `json:"name"`
		Description *string `json:"description"`
	}
	var input updateProjectRequest
	if c.ShouldBindJSON(&input) != nil {
		failure(c, errInvalidArgument, "at least one valid field is required")
		return
	}
	if input.Name == nil && input.Description == nil {
		failure(c, errInvalidArgument, "at least one valid field is required")
		return
	}
	if input.Name != nil && strings.TrimSpace(*input.Name) == "" {
		failure(c, errInvalidArgument, "at least one valid field is required")
		return
	}
	p, err := h.store.Update(c.Request.Context(), c.Param("project_id"), input.Name, input.Description, time.Now().UTC().UnixMilli())
	if err != nil {
		projectError(c, err)
		return
	}
	type updateProjectResponse struct {
		ID          string `json:"id"`
		Name        string `json:"name"`
		Description string `json:"description"`
		CreatedAt   string `json:"created_at"`
		UpdatedAt   string `json:"updated_at"`
	}
	success(c, updateProjectResponse{ID: p.ID, Name: p.Name, Description: p.Description,
		CreatedAt: time.UnixMilli(p.CreatedAt).UTC().Format(time.RFC3339Nano),
		UpdatedAt: time.UnixMilli(p.UpdatedAt).UTC().Format(time.RFC3339Nano)})
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
	if err := h.files.DeleteProject(context.WithoutCancel(c.Request.Context()), c.Param("project_id")); err != nil {
		_ = c.Error(err)
		internalError(c)
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
