package handlers

import (
	"encoding/json"
	"fmt"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"

	"example.com/go-master-web-sample/internal/auth"
	"example.com/go-master-web-sample/internal/httpx"
	appmw "example.com/go-master-web-sample/internal/middleware"
	"example.com/go-master-web-sample/internal/models"
	"example.com/go-master-web-sample/internal/store"
)

type Handler struct {
	store *store.MemoryStore
	auth  *auth.Manager
}

func New(store *store.MemoryStore, auth *auth.Manager) *Handler {
	return &Handler{store: store, auth: auth}
}

func (h *Handler) Health(w http.ResponseWriter, r *http.Request) {
	httpx.OK(w, map[string]any{
		"status":     "ok",
		"request_id": chimw.GetReqID(r.Context()),
		"time":       time.Now().UTC(),
	})
}

func (h *Handler) Ready(w http.ResponseWriter, r *http.Request) {
	httpx.OK(w, map[string]string{"status": "ready"})
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var req models.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.Error(w, http.StatusBadRequest, "invalid json body", err.Error())
		return
	}

	user, ok := h.store.Authenticate(strings.TrimSpace(req.Username), req.Password)
	if !ok {
		httpx.Error(w, http.StatusUnauthorized, "invalid credentials", nil)
		return
	}

	token, err := h.auth.IssueToken(user)
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, "failed to issue token", err.Error())
		return
	}

	httpx.OK(w, map[string]any{
		"token": token,
		"user": map[string]any{
			"id":       user.ID,
			"username": user.Username,
			"role":     user.Role,
		},
	})
}

func (h *Handler) Me(w http.ResponseWriter, r *http.Request) {
	claims, ok := appmw.ClaimsFromContext(r.Context())
	if !ok {
		httpx.Error(w, http.StatusUnauthorized, "missing auth context", nil)
		return
	}

	httpx.OK(w, map[string]any{
		"username": claims.Subject,
		"role":     claims.Role,
		"expires":  claims.ExpiresAt.Time,
	})
}

func (h *Handler) AdminAudit(w http.ResponseWriter, r *http.Request) {
	httpx.OK(w, map[string]any{
		"message": "admin endpoint reached",
		"at":      time.Now().UTC(),
	})
}

func (h *Handler) ListItems(w http.ResponseWriter, r *http.Request) {
	qs := r.URL.Query()
	limit := parseInt(qs.Get("page_size"), 10)
	page := parseInt(qs.Get("page"), 1)
	if limit < 1 || limit > 100 {
		httpx.Error(w, http.StatusBadRequest, "page_size must be between 1 and 100", nil)
		return
	}
	if page < 1 {
		httpx.Error(w, http.StatusBadRequest, "page must be at least 1", nil)
		return
	}

	items, total := h.store.ListItems(qs.Get("q"), qs.Get("status"), page, limit)
	httpx.OK(w, map[string]any{
		"items": items,
		"pagination": map[string]int{
			"page":      page,
			"page_size": limit,
			"total":     total,
		},
	})
}

func (h *Handler) GetItem(w http.ResponseWriter, r *http.Request) {
	id, ok := itemIDFromRequest(w, r)
	if !ok {
		return
	}

	item, found := h.store.GetItem(id)
	if !found {
		httpx.Error(w, http.StatusNotFound, "item not found", nil)
		return
	}

	httpx.OK(w, item)
}

func (h *Handler) CreateItem(w http.ResponseWriter, r *http.Request) {
	var req models.CreateItemRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.Error(w, http.StatusBadRequest, "invalid json body", err.Error())
		return
	}

	if err := validateRequiredFields(req.Title, req.Description); err != nil {
		httpx.Error(w, http.StatusBadRequest, err.Error(), nil)
		return
	}
	if req.Status == "" {
		req.Status = "draft"
	}

	item, err := h.store.CreateItem(req)
	if err != nil {
		httpx.Error(w, http.StatusBadRequest, err.Error(), nil)
		return
	}

	httpx.Created(w, item)
}

func (h *Handler) ReplaceItem(w http.ResponseWriter, r *http.Request) {
	id, ok := itemIDFromRequest(w, r)
	if !ok {
		return
	}

	var req models.ReplaceItemRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.Error(w, http.StatusBadRequest, "invalid json body", err.Error())
		return
	}
	if err := validateRequiredFields(req.Title, req.Description); err != nil {
		httpx.Error(w, http.StatusBadRequest, err.Error(), nil)
		return
	}

	item, err := h.store.ReplaceItem(id, req)
	if err != nil {
		handleItemWriteError(w, err)
		return
	}

	httpx.OK(w, item)
}

func (h *Handler) PatchItem(w http.ResponseWriter, r *http.Request) {
	id, ok := itemIDFromRequest(w, r)
	if !ok {
		return
	}

	var req models.PatchItemRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.Error(w, http.StatusBadRequest, "invalid json body", err.Error())
		return
	}

	item, err := h.store.PatchItem(id, req)
	if err != nil {
		handleItemWriteError(w, err)
		return
	}

	httpx.OK(w, item)
}

func (h *Handler) DeleteItem(w http.ResponseWriter, r *http.Request) {
	id, ok := itemIDFromRequest(w, r)
	if !ok {
		return
	}

	if !h.store.DeleteItem(id) {
		httpx.Error(w, http.StatusNotFound, "item not found", nil)
		return
	}

	httpx.JSON(w, http.StatusOK, map[string]any{
		"data": map[string]any{
			"deleted": true,
			"id":      id,
		},
	})
}

func (h *Handler) EchoForm(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		httpx.Error(w, http.StatusBadRequest, "failed to parse form", err.Error())
		return
	}

	httpx.OK(w, map[string]any{
		"name":     r.FormValue("name"),
		"email":    r.FormValue("email"),
		"comments": r.FormValue("comments"),
	})
}

func (h *Handler) UploadFile(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseMultipartForm(5 << 20); err != nil {
		httpx.Error(w, http.StatusBadRequest, "failed to parse multipart form", err.Error())
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		httpx.Error(w, http.StatusBadRequest, "missing file field", err.Error())
		return
	}
	defer file.Close()

	size, err := uploadedSize(header)
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, "failed to inspect upload", err.Error())
		return
	}

	httpx.OK(w, map[string]any{
		"filename": header.Filename,
		"size":     size,
		"mime":     header.Header.Get("Content-Type"),
		"ext":      filepath.Ext(header.Filename),
	})
}

func (h *Handler) RequestInfo(w http.ResponseWriter, r *http.Request) {
	httpx.OK(w, map[string]any{
		"method":     r.Method,
		"path":       r.URL.Path,
		"user_agent": r.UserAgent(),
		"client_ip":  r.RemoteAddr,
		"headers": map[string]string{
			"accept":      r.Header.Get("Accept"),
			"contentType": r.Header.Get("Content-Type"),
		},
	})
}

func itemIDFromRequest(w http.ResponseWriter, r *http.Request) (int, bool) {
	id, err := strconv.Atoi(chi.URLParam(r, "itemID"))
	if err != nil {
		httpx.Error(w, http.StatusBadRequest, "item id must be an integer", nil)
		return 0, false
	}
	return id, true
}

func parseInt(raw string, fallback int) int {
	if raw == "" {
		return fallback
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return -1
	}
	return value
}

func validateRequiredFields(title, description string) error {
	if strings.TrimSpace(title) == "" {
		return fmt.Errorf("title is required")
	}
	if strings.TrimSpace(description) == "" {
		return fmt.Errorf("description is required")
	}
	return nil
}

func handleItemWriteError(w http.ResponseWriter, err error) {
	switch err {
	case store.ErrItemNotFound:
		httpx.Error(w, http.StatusNotFound, "item not found", nil)
	case store.ErrInvalidItemState:
		httpx.Error(w, http.StatusBadRequest, err.Error(), nil)
	default:
		httpx.Error(w, http.StatusInternalServerError, "unexpected storage error", err.Error())
	}
}

func uploadedSize(header *multipart.FileHeader) (int64, error) {
	if header.Size > 0 {
		return header.Size, nil
	}
	file, err := header.Open()
	if err != nil {
		return 0, err
	}
	defer file.Close()

	var total int64
	buf := make([]byte, 1024)
	for {
		n, readErr := file.Read(buf)
		total += int64(n)
		if readErr != nil {
			if readErr.Error() == "EOF" {
				return total, nil
			}
			return 0, readErr
		}
	}
}
