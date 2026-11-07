package server

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"math"
	"mime"
	"net/http"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"rewind/internal/auth"
	"rewind/internal/database"
	"rewind/internal/id"
	"rewind/internal/storage"

	"github.com/gin-gonic/gin"
)

type backupHandler struct {
	store *database.BackupStore
	files storage.Store
}

func registerBackupRoutes(router *gin.RouterGroup, authService *auth.Service, tokens *database.TokenStore, store *database.BackupStore, files storage.Store) {
	h := &backupHandler{store: store, files: files}
	group := router.Group("/backups")
	group.POST("", backupAuth(tokens, nil), h.create)
	group.POST("/history", backupAuth(tokens, authService), h.history)
	group.GET("/download/:project_id/:version_id", backupAuth(tokens, authService), h.download)
}

func backupAuth(tokens *database.TokenStore, authService *auth.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		parts := strings.Fields(c.GetHeader("Authorization"))
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			failure(c, errUnauthorized, "unauthorized")
			c.Abort()
			return
		}
		if authService != nil && strings.HasPrefix(strings.ToLower(parts[1]), "login_") {
			if err := authService.Authenticate(c.Request.Context(), parts[1]); err != nil {
				authError(c, err, errUnauthorized, "unauthorized")
				c.Abort()
				return
			}
			c.Next()
			return
		}
		credential := strings.ToLower(parts[1])
		if !strings.HasPrefix(credential, "api_") {
			failure(c, errUnauthorized, "unauthorized")
			c.Abort()
			return
		}
		hash := sha256.Sum256([]byte(credential))
		exists, err := tokens.ExistsHash(c.Request.Context(), hex.EncodeToString(hash[:]))
		if err != nil {
			internalError(c)
			c.Abort()
			return
		}
		if !exists {
			failure(c, errUnauthorized, "unauthorized")
			c.Abort()
			return
		}
		c.Next()
	}
}

func (h *backupHandler) history(c *gin.Context) {
	type backupHistoryRequest struct {
		ProjectID string          `json:"project_id"`
		Key       string          `json:"key"`
		Page      json.RawMessage `json:"page"`
		PageSize  json.RawMessage `json:"page_size"`
	}
	var input backupHistoryRequest
	if c.ShouldBindJSON(&input) != nil {
		failure(c, errInvalidArgument, "project_id and key are required")
		return
	}
	if input.ProjectID == "" || input.Key == "" {
		failure(c, errInvalidArgument, "project_id and key are required")
		return
	}
	page, size := int64(1), int64(20)
	if (input.Page != nil && (json.Unmarshal(input.Page, &page) != nil || string(input.Page) == "null")) ||
		(input.PageSize != nil && (json.Unmarshal(input.PageSize, &size) != nil || string(input.PageSize) == "null")) ||
		page < 1 || size < 1 || size > 100 || page-1 > math.MaxInt64/size {
		failure(c, errInvalidArgument, "invalid pagination parameters")
		return
	}
	versions, err := h.store.History(c.Request.Context(), strings.ToLower(input.ProjectID), input.Key, size, (page-1)*size)
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			failure(c, errNotFound, "file not found")
		} else {
			internalError(c)
		}
		return
	}
	type versionResponse struct {
		VersionID   string `json:"version_id"`
		Size        int64  `json:"size"`
		Description string `json:"description"`
		CreatedAt   string `json:"created_at"`
	}
	result := make([]versionResponse, 0, len(versions))
	for _, version := range versions {
		result = append(result, versionResponse{version.VersionID, version.Size, version.Description, time.UnixMilli(version.CreatedAt).UTC().Format(time.RFC3339Nano)})
	}
	success(c, result)
}

func (h *backupHandler) download(c *gin.Context) {
	projectID, versionID := strings.ToLower(c.Param("project_id")), strings.ToLower(c.Param("version_id"))
	b, err := h.store.Version(c.Request.Context(), projectID, versionID)
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			failure(c, errNotFound, "version not found")
		} else {
			internalError(c)
		}
		return
	}
	f, err := h.files.Open(c.Request.Context(), storage.VersionKey{ProjectID: projectID, FileID: b.FileID, VersionID: versionID})
	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			failure(c, errNotFound, "version file not found")
		} else {
			internalError(c)
		}
		return
	}
	defer f.Close()
	contentType := b.ContentType
	if _, _, err := mime.ParseMediaType(contentType); err != nil {
		contentType = "application/octet-stream"
	}
	disposition := mime.FormatMediaType("attachment", map[string]string{"filename": b.OriginalFilename})
	if disposition == "" {
		disposition = `attachment; filename="download"`
	}
	c.Header("Content-Type", contentType)
	c.Header("Content-Disposition", disposition)
	c.Header("Content-Length", strconv.FormatInt(b.Size, 10))
	written, err := io.Copy(c.Writer, f)
	if err == nil && written != b.Size {
		err = io.ErrUnexpectedEOF
	}
	if err != nil {
		_ = c.Error(err)
		if !c.Writer.Written() {
			c.Writer.Header().Del("Content-Length")
			c.Writer.Header().Del("Content-Disposition")
			c.Writer.Header().Del("Content-Type")
			internalError(c)
		} else {
			c.Abort()
			panic(http.ErrAbortHandler)
		}
		return
	}
	c.Writer.WriteHeaderNow()
}

func (h *backupHandler) create(c *gin.Context) {
	reader, err := c.Request.MultipartReader()
	if err != nil {
		failure(c, errInvalidArgument, "multipart/form-data is required")
		return
	}
	ctx := c.Request.Context()
	cleanupCtx := context.WithoutCancel(ctx)
	var staged storage.StagedFile
	defer func() {
		if staged != nil {
			if err := staged.Abort(cleanupCtx); err != nil {
				_ = c.Error(err)
			}
		}
	}()
	var info storage.FileInfo
	type createBackupRequest struct {
		ProjectID   string
		Key         string
		Description string
		Filename    string
		ContentType string
	}
	var input createBackupRequest
	fields := make(map[string]struct{})
	for {
		part, err := reader.NextPart()
		if err == io.EOF {
			break
		}
		if err != nil {
			failure(c, errInvalidArgument, "invalid multipart body")
			return
		}
		name := part.FormName()
		if name == "file" {
			if staged != nil || part.FileName() == "" || !utf8.ValidString(part.FileName()) {
				failure(c, errInvalidArgument, "exactly one file with a filename is required")
				return
			}
			input.Filename, input.ContentType = part.FileName(), part.Header.Get("Content-Type")
			if input.ContentType == "" {
				input.ContentType = "application/octet-stream"
			}
			staged, info, err = h.files.Stage(ctx, part)
			if err != nil {
				if errors.Is(err, io.ErrUnexpectedEOF) {
					failure(c, errInvalidArgument, "incomplete file upload")
				} else {
					internalError(c)
				}
				return
			}
			continue
		}
		_, duplicate := fields[name]
		if (name != "project_id" && name != "key" && name != "description") || duplicate || part.FileName() != "" {
			failure(c, errInvalidArgument, "invalid or duplicate form field")
			return
		}
		// File content is streamed; only text fields are buffered, up to 1 MiB each.
		value, err := io.ReadAll(io.LimitReader(part, (1<<20)+1))
		if err != nil || len(value) > 1<<20 || !utf8.Valid(value) {
			failure(c, errInvalidArgument, "form fields must be valid UTF-8 and at most 1 MiB each")
			return
		}
		fields[name] = struct{}{}
		switch name {
		case "project_id":
			input.ProjectID = strings.ToLower(string(value))
		case "key":
			input.Key = string(value)
		case "description":
			input.Description = string(value)
		}
	}
	if input.ProjectID == "" || input.Key == "" || staged == nil {
		failure(c, errInvalidArgument, "project_id, key and file are required; description must be at most 1024 characters")
		return
	}
	if utf8.RuneCountInString(input.Description) > 1024 {
		failure(c, errInvalidArgument, "project_id, key and file are required; description must be at most 1024 characters")
		return
	}
	fileID, err := id.File()
	if err != nil {
		internalError(c)
		return
	}
	versionID, err := id.Version()
	if err != nil {
		internalError(c)
		return
	}
	b := database.Backup{
		ProjectID: input.ProjectID, Key: input.Key, FileID: fileID, VersionID: versionID,
		Size: info.Size, SHA256: info.SHA256, Description: input.Description,
		OriginalFilename: input.Filename, ContentType: input.ContentType, CreatedAt: time.Now().UnixMilli(),
	}
	var published bool
	var storageKey storage.VersionKey
	b, err = h.store.Create(ctx, b, func(fileID string) error {
		storageKey = storage.VersionKey{ProjectID: input.ProjectID, FileID: fileID, VersionID: versionID}
		if err := staged.Commit(ctx, storageKey); err != nil {
			return err
		}
		published = true
		return nil
	})
	if err != nil {
		if published {
			if cleanupErr := h.files.Delete(cleanupCtx, storageKey); cleanupErr != nil {
				_ = c.Error(cleanupErr)
			}
		}
		if errors.Is(err, database.ErrNotFound) {
			failure(c, errNotFound, "project not found")
		} else {
			internalError(c)
		}
		return
	}
	success(c, gin.H{
		"project_id": b.ProjectID, "key": b.Key, "file_id": b.FileID, "version_id": b.VersionID,
		"size": b.Size, "sha256": b.SHA256, "description": b.Description,
		"original_filename": b.OriginalFilename, "content_type": b.ContentType,
		"created_at": time.UnixMilli(b.CreatedAt).UTC().Format(time.RFC3339Nano),
	})
}
