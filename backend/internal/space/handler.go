package space

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"image"
	"net/http"
	"strconv"
	"strings"

	"github.com/Dimonchester/ttk-room-map/backend/internal/auth"
	"github.com/Dimonchester/ttk-room-map/backend/internal/logger"
)

type Handler struct {
	repo      *Repository
	processor *MediaProcessor
	auth      *auth.Handler
}

func NewHandler(repo *Repository, processor *MediaProcessor, authHandler *auth.Handler) *Handler {
	return &Handler{
		repo:      repo,
		processor: processor,
		auth:      authHandler,
	}
}

func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.Handle("GET /api/maps", h.auth.RequireAuth(http.HandlerFunc(h.ListMaps)))
	mux.Handle("POST /api/maps", h.auth.RequireAuth(http.HandlerFunc(h.CreateMap)))
	mux.Handle("GET /api/maps/{mapId}", h.auth.RequireAuth(http.HandlerFunc(h.GetMap)))
	mux.Handle("PUT /api/maps/{mapId}", h.auth.RequireAuth(http.HandlerFunc(h.UpdateMap)))
	mux.Handle("DELETE /api/maps/{mapId}", h.auth.RequireAuth(http.HandlerFunc(h.DeleteMap)))
	mux.Handle("POST /api/maps/{mapId}/publish", h.auth.RequireAuth(http.HandlerFunc(h.PublishMap)))
	mux.Handle("POST /api/maps/{mapId}/plan-image", h.auth.RequireAuth(http.HandlerFunc(h.UploadPlanImage)))
	mux.Handle("POST /api/maps/{mapId}/panoramas", h.auth.RequireAuth(http.HandlerFunc(h.UploadPanorama)))
	mux.Handle("POST /api/maps/{mapId}/panorama-jobs", h.auth.RequireAuth(http.HandlerFunc(h.CreatePanoramaJob)))
	mux.Handle("GET /api/panoramas/{panoramaId}", h.auth.RequireAuth(http.HandlerFunc(h.GetPanoramaStatus)))
	mux.HandleFunc("GET /api/public/maps/{slug}", h.GetPublicMap)
}

func (h *Handler) ListMaps(w http.ResponseWriter, r *http.Request) {
	u, _ := auth.UserFromContext(r.Context())
	items, err := h.repo.ListByOwner(r.Context(), u.ID)
	if err != nil {
		writeProblem(w, http.StatusInternalServerError, "db_error", "Ошибка получения списка карт")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": items})
}

func (h *Handler) CreateMap(w http.ResponseWriter, r *http.Request) {
	u, _ := auth.UserFromContext(r.Context())
	var body struct {
		Title string `json:"title"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || strings.TrimSpace(body.Title) == "" {
		logger.Warn(logger.CatMap, "ошибка создания карты: пустой title от user_id=%s", u.ID)
		writeProblem(w, http.StatusUnprocessableEntity, "validation_error", "Название карты обязательно")
		return
	}

	doc, err := h.repo.CreateMap(r.Context(), u.ID, strings.TrimSpace(body.Title))
	if err != nil {
		writeProblem(w, http.StatusInternalServerError, "db_error", "Не удалось создать карту")
		return
	}

	w.Header().Set("ETag", fmt.Sprintf(`"%d"`, doc.Version))
	writeJSON(w, http.StatusCreated, doc)
}

func (h *Handler) GetMap(w http.ResponseWriter, r *http.Request) {
	u, _ := auth.UserFromContext(r.Context())
	mapID := r.PathValue("mapId")

	doc, err := h.repo.GetMap(r.Context(), mapID, u.ID)
	if err != nil {
		if errors.Is(err, ErrMapNotFound) {
			writeProblem(w, http.StatusNotFound, "not_found", "Карта не найдена")
			return
		}
		writeProblem(w, http.StatusInternalServerError, "db_error", "Ошибка загрузки карты")
		return
	}

	w.Header().Set("ETag", fmt.Sprintf(`"%d"`, doc.Version))
	writeJSON(w, http.StatusOK, doc)
}

func (h *Handler) UpdateMap(w http.ResponseWriter, r *http.Request) {
	u, _ := auth.UserFromContext(r.Context())
	mapID := r.PathValue("mapId")

	ifMatch := strings.Trim(r.Header.Get("If-Match"), `"`)
	expectedVersion, err := strconv.ParseInt(ifMatch, 10, 64)
	if err != nil {
		expectedVersion = 1
	}

	var input MapDoc
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		logger.Warn(logger.CatMap, "некорректный JSON при сохранении карты map_id=%s: %v", mapID, err)
		writeProblem(w, http.StatusBadRequest, "invalid_json", "Некорректный документ карты")
		return
	}

	updated, err := h.repo.UpdateMap(r.Context(), mapID, u.ID, expectedVersion, input)
	if err != nil {
		if errors.Is(err, ErrMapNotFound) {
			writeProblem(w, http.StatusNotFound, "not_found", "Карта не найдена")
			return
		}
		if errors.Is(err, ErrVersionConflict) {
			writeProblem(w, http.StatusConflict, "version_conflict", "Карта была изменена в другой сессии")
			return
		}
		writeProblem(w, http.StatusInternalServerError, "db_error", "Ошибка сохранения карты")
		return
	}

	w.Header().Set("ETag", fmt.Sprintf(`"%d"`, updated.Version))
	writeJSON(w, http.StatusOK, updated)
}

func (h *Handler) DeleteMap(w http.ResponseWriter, r *http.Request) {
	u, _ := auth.UserFromContext(r.Context())
	mapID := r.PathValue("mapId")

	if err := h.repo.DeleteMap(r.Context(), mapID, u.ID); err != nil {
		if errors.Is(err, ErrMapNotFound) {
			writeProblem(w, http.StatusNotFound, "not_found", "Карта не найдена")
			return
		}
		writeProblem(w, http.StatusInternalServerError, "db_error", "Ошибка удаления карты")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) PublishMap(w http.ResponseWriter, r *http.Request) {
	u, _ := auth.UserFromContext(r.Context())
	mapID := r.PathValue("mapId")

	var req struct {
		Access   string `json:"access"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeProblem(w, http.StatusBadRequest, "invalid_json", "Некорректный JSON")
		return
	}
	if req.Access == "" {
		req.Access = "public"
	}

	slug, err := h.repo.PublishMap(r.Context(), mapID, u.ID, req.Access, req.Password)
	if err != nil {
		if errors.Is(err, ErrMapNotFound) {
			writeProblem(w, http.StatusNotFound, "not_found", "Карта не найдена")
			return
		}
		writeProblem(w, http.StatusInternalServerError, "db_error", "Ошибка публикации карты")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{
		"slug":   slug,
		"url":    "/m/" + slug,
		"access": req.Access,
	})
}

func (h *Handler) UploadPlanImage(w http.ResponseWriter, r *http.Request) {
	u, _ := auth.UserFromContext(r.Context())
	mapID := r.PathValue("mapId")

	if _, err := h.repo.GetMap(r.Context(), mapID, u.ID); err != nil {
		writeProblem(w, http.StatusNotFound, "not_found", "Карта не найдена")
		return
	}

	if err := r.ParseMultipartForm(20 << 20); err != nil {
		logger.Warn(logger.CatMedia, "ошибка чтения multipart/form-data плана map_id=%s: %v", mapID, err)
		writeProblem(w, http.StatusBadRequest, "invalid_multipart", "Ожидался файл в multipart/form-data")
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		writeProblem(w, http.StatusUnprocessableEntity, "file_required", "Поле file обязательно")
		return
	}
	defer file.Close()

	planImg, err := h.processor.ProcessPlanImage(mapID, header.Filename, file)
	if err != nil {
		writeProblem(w, http.StatusUnsupportedMediaType, "invalid_image", "Не удалось обработать изображение плана")
		return
	}

	writeJSON(w, http.StatusCreated, planImg)
}

func (h *Handler) UploadPanorama(w http.ResponseWriter, r *http.Request) {
	u, _ := auth.UserFromContext(r.Context())
	mapID := r.PathValue("mapId")

	if _, err := h.repo.GetMap(r.Context(), mapID, u.ID); err != nil {
		writeProblem(w, http.StatusNotFound, "not_found", "Карта не найдена")
		return
	}

	if err := r.ParseMultipartForm(32 << 20); err != nil {
		writeProblem(w, http.StatusBadRequest, "invalid_multipart", "Ошибка чтения multipart/form-data")
		return
	}

	panoramaID := strings.TrimSpace(r.FormValue("panoramaId"))
	if panoramaID == "" {
		writeProblem(w, http.StatusUnprocessableEntity, "panorama_id_required", "Поле panoramaId обязательно")
		return
	}

	file, _, err := r.FormFile("file")
	if err != nil {
		writeProblem(w, http.StatusUnprocessableEntity, "file_required", "Поле file обязательно")
		return
	}
	defer file.Close()

	srcImg, _, err := image.Decode(file)
	if err != nil {
		logger.Warn(logger.CatMedia, "ошибка декодирования файла панорамы panorama_id=%s: %v", panoramaID, err)
		writeProblem(w, http.StatusUnsupportedMediaType, "invalid_image", "Неподдерживаемый формат изображения")
		return
	}

	initial := PanoramaMedia{PanoramaID: panoramaID, Status: "processing"}
	if err := h.repo.UpsertPanoramaStatus(r.Context(), mapID, initial); err != nil {
		writeProblem(w, http.StatusInternalServerError, "db_error", "Не удалось зарегистрировать задачу панорамы")
		return
	}

	go func() {
		bgCtx := context.Background()
		fullURL, prevURL, wPx, hPx, err := h.processor.ProcessUploadedPanorama(panoramaID, srcImg)
		if err != nil {
			errMsg := err.Error()
			_ = h.repo.UpsertPanoramaStatus(bgCtx, mapID, PanoramaMedia{
				PanoramaID: panoramaID,
				Status:     "failed",
				Error:      &errMsg,
			})
			return
		}
		_ = h.repo.UpsertPanoramaStatus(bgCtx, mapID, PanoramaMedia{
			PanoramaID: panoramaID,
			Status:     "ready",
			PreviewURL: &prevURL,
			FullURL:    &fullURL,
			Width:      &wPx,
			Height:     &hPx,
		})
	}()

	writeJSON(w, http.StatusAccepted, initial)
}

func (h *Handler) CreatePanoramaJob(w http.ResponseWriter, r *http.Request) {
	u, _ := auth.UserFromContext(r.Context())
	mapID := r.PathValue("mapId")

	if _, err := h.repo.GetMap(r.Context(), mapID, u.ID); err != nil {
		writeProblem(w, http.StatusNotFound, "not_found", "Карта не найдена")
		return
	}

	if err := r.ParseMultipartForm(64 << 20); err != nil {
		logger.Warn(logger.CatMedia, "ошибка чтения серии кадров панорамы map_id=%s: %v", mapID, err)
		writeProblem(w, http.StatusBadRequest, "invalid_multipart", "Ошибка чтения multipart/form-data")
		return
	}

	panoramaID := strings.TrimSpace(r.FormValue("panoramaId"))
	if panoramaID == "" {
		writeProblem(w, http.StatusUnprocessableEntity, "panorama_id_required", "Поле panoramaId обязательно")
		return
	}

	files := r.MultipartForm.File["frames"]
	if len(files) == 0 {
		writeProblem(w, http.StatusUnprocessableEntity, "frames_required", "Не переданы кадры в поле frames")
		return
	}

	var decodedFrames []image.Image
	for _, fh := range files {
		f, err := fh.Open()
		if err != nil {
			continue
		}
		img, _, err := image.Decode(f)
		_ = f.Close()
		if err != nil {
			logger.Warn(logger.CatMedia, "пропущен поврежденный кадр %s для panorama_id=%s: %v", fh.Filename, panoramaID, err)
			continue
		}
		decodedFrames = append(decodedFrames, img)
	}

	if len(decodedFrames) == 0 {
		writeProblem(w, http.StatusUnprocessableEntity, "invalid_frames", "Ни один кадр не удалось декодировать")
		return
	}

	initial := PanoramaMedia{PanoramaID: panoramaID, Status: "processing"}
	if err := h.repo.UpsertPanoramaStatus(r.Context(), mapID, initial); err != nil {
		writeProblem(w, http.StatusInternalServerError, "db_error", "Не удалось создать задачу сшивки панорамы")
		return
	}

	go func(frames []image.Image) {
		bgCtx := context.Background()
		fullURL, prevURL, wPx, hPx, err := h.processor.StitchFramesToPanorama(panoramaID, frames)
		if err != nil {
			errMsg := err.Error()
			_ = h.repo.UpsertPanoramaStatus(bgCtx, mapID, PanoramaMedia{
				PanoramaID: panoramaID,
				Status:     "failed",
				Error:      &errMsg,
			})
			return
		}
		_ = h.repo.UpsertPanoramaStatus(bgCtx, mapID, PanoramaMedia{
			PanoramaID: panoramaID,
			Status:     "ready",
			PreviewURL: &prevURL,
			FullURL:    &fullURL,
			Width:      &wPx,
			Height:     &hPx,
		})
	}(decodedFrames)

	writeJSON(w, http.StatusAccepted, initial)
}

func (h *Handler) GetPanoramaStatus(w http.ResponseWriter, r *http.Request) {
	panoramaID := r.PathValue("panoramaId")
	status, err := h.repo.GetPanoramaStatus(r.Context(), panoramaID)
	if err != nil {
		if errors.Is(err, ErrPanoramaNotFound) {
			writeProblem(w, http.StatusNotFound, "not_found", "Панорама не найдена")
			return
		}
		writeProblem(w, http.StatusInternalServerError, "db_error", "Ошибка получения статуса панорамы")
		return
	}
	writeJSON(w, http.StatusOK, status)
}

func (h *Handler) GetPublicMap(w http.ResponseWriter, r *http.Request) {
	slug := r.PathValue("slug")
	doc, _, err := h.repo.GetPublicBySlug(r.Context(), slug)
	if err != nil {
		if errors.Is(err, ErrMapNotFound) {
			writeProblem(w, http.StatusNotFound, "not_found", "Публичная карта не найдена")
			return
		}
		writeProblem(w, http.StatusInternalServerError, "db_error", "Ошибка загрузки публичной карты")
		return
	}

	etag := fmt.Sprintf(`"%d"`, doc.Version)
	if r.Header.Get("If-None-Match") == etag {
		w.WriteHeader(http.StatusNotModified)
		return
	}

	w.Header().Set("ETag", etag)
	writeJSON(w, http.StatusOK, doc)
}

func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}

func writeProblem(w http.ResponseWriter, status int, code, detail string) {
	w.Header().Set("Content-Type", "application/problem+json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(auth.Problem{
		Type:    "about:blank",
		Title:   http.StatusText(status),
		Status:  status,
		Code:    code,
		Detail:  detail,
		Message: detail,
	})
}
