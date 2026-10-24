package space

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/Dimonchester/ttk-room-map/backend/internal/logger"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrMapNotFound      = errors.New("map not found")
	ErrVersionConflict  = errors.New("map version conflict")
	ErrPanoramaNotFound = errors.New("panorama not found")
)

type Repository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

func (r *Repository) CreateMap(ctx context.Context, ownerID, title string) (*MapDoc, error) {
	var mapID, floorID string
	err := r.pool.QueryRow(ctx, "SELECT gen_random_uuid()::text, gen_random_uuid()::text").Scan(&mapID, &floorID)
	if err != nil {
		logger.Error(logger.CatDB, "ошибка генерации UUID для новой карты: %v", err)
		return nil, err
	}

	now := time.Now().UTC()
	doc := &MapDoc{
		ID:          mapID,
		Title:       title,
		IsPublished: false,
		Access:      "public",
		Version:     1,
		UpdatedAt:   now,
		Floors: []Floor{
			{
				ID:   floorID,
				Name: "Этаж 1",
				Plan: Plan{
					Width:        1000,
					Height:       800,
					ImageOpacity: 1,
				},
				Rooms: []Room{},
				Paths: []Path{},
				Pois:  []Poi{},
			},
		},
		Panoramas: []Panorama{},
	}

	docBytes, err := json.Marshal(doc)
	if err != nil {
		return nil, err
	}

	const query = `
		INSERT INTO maps (id, owner_id, title, version, doc, created_at, updated_at)
		VALUES ($1::uuid, $2::uuid, $3, 1, $4::jsonb, $5, $5)
	`
	if _, err := r.pool.Exec(ctx, query, mapID, ownerID, title, string(docBytes), now); err != nil {
		logger.Error(logger.CatDB, "ошибка создания карты в БД owner_id=%s: %v", ownerID, err)
		return nil, err
	}

	logger.Info(logger.CatMap, "создана новая карта map_id=%s owner_id=%s title=%q", mapID, ownerID, title)
	return doc, nil
}

func (r *Repository) ListByOwner(ctx context.Context, ownerID string) ([]MapSummary, error) {
	const query = `
		SELECT id::text, title, is_published, slug, version, updated_at, thumbnail_url
		FROM maps
		WHERE owner_id = $1::uuid
		ORDER BY updated_at DESC
	`

	rows, err := r.pool.Query(ctx, query, ownerID)
	if err != nil {
		logger.Error(logger.CatDB, "ошибка получения списка карт owner_id=%s: %v", ownerID, err)
		return nil, err
	}
	defer rows.Close()

	items := []MapSummary{}
	for rows.Next() {
		var m MapSummary
		if err := rows.Scan(&m.ID, &m.Title, &m.IsPublished, &m.Slug, &m.Version, &m.UpdatedAt, &m.ThumbnailURL); err != nil {
			logger.Error(logger.CatDB, "ошибка чтения строки карты owner_id=%s: %v", ownerID, err)
			return nil, err
		}
		items = append(items, m)
	}
	return items, nil
}

func (r *Repository) GetMap(ctx context.Context, mapID, ownerID string) (*MapDoc, error) {
	const query = `
		SELECT id::text, title, slug, is_published, access, version, updated_at, doc
		FROM maps
		WHERE id = $1::uuid AND owner_id = $2::uuid
	`

	var (
		id, title, access string
		slug              *string
		isPublished       bool
		version           int64
		updatedAt         time.Time
		docBytes          []byte
	)

	err := r.pool.QueryRow(ctx, query, mapID, ownerID).Scan(
		&id, &title, &slug, &isPublished, &access, &version, &updatedAt, &docBytes,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrMapNotFound
		}
		logger.Error(logger.CatDB, "ошибка получения карты map_id=%s: %v", mapID, err)
		return nil, err
	}

	var doc MapDoc
	if err := json.Unmarshal(docBytes, &doc); err != nil {
		logger.Error(logger.CatMap, "ошибка десериализации JSON документа карты map_id=%s: %v", mapID, err)
		return nil, err
	}

	doc.ID = id
	doc.Title = title
	doc.Slug = slug
	doc.IsPublished = isPublished
	doc.Access = access
	doc.Version = version
	doc.UpdatedAt = updatedAt

	r.enrichPanoramas(ctx, &doc)
	return &doc, nil
}

func (r *Repository) UpdateMap(ctx context.Context, mapID, ownerID string, expectedVersion int64, input MapDoc) (*MapDoc, error) {
	now := time.Now().UTC()
	newVersion := expectedVersion + 1
	input.ID = mapID
	input.Version = newVersion
	input.UpdatedAt = now

	docBytes, err := json.Marshal(input)
	if err != nil {
		return nil, err
	}

	const query = `
		UPDATE maps
		SET title = $1, version = $2, doc = $3::jsonb, updated_at = $4
		WHERE id = $5::uuid AND owner_id = $6::uuid AND version = $7
	`

	tag, err := r.pool.Exec(ctx, query, input.Title, newVersion, string(docBytes), now, mapID, ownerID, expectedVersion)
	if err != nil {
		logger.Error(logger.CatDB, "ошибка обновления карты map_id=%s: %v", mapID, err)
		return nil, err
	}

	if tag.RowsAffected() == 0 {
		var exists bool
		_ = r.pool.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM maps WHERE id = $1::uuid AND owner_id = $2::uuid)", mapID, ownerID).Scan(&exists)
		if !exists {
			return nil, ErrMapNotFound
		}
		logger.Warn(logger.CatMap, "конфликт версий при сохранении карты map_id=%s (ожидалась версия %d)", mapID, expectedVersion)
		return nil, ErrVersionConflict
	}

	logger.Info(logger.CatMap, "сохранена карта map_id=%s version=%d", mapID, newVersion)
	return r.GetMap(ctx, mapID, ownerID)
}

func (r *Repository) DeleteMap(ctx context.Context, mapID, ownerID string) error {
	tag, err := r.pool.Exec(ctx, "DELETE FROM maps WHERE id = $1::uuid AND owner_id = $2::uuid", mapID, ownerID)
	if err != nil {
		logger.Error(logger.CatDB, "ошибка удаления карты map_id=%s: %v", mapID, err)
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrMapNotFound
	}
	logger.Info(logger.CatMap, "удалена карта map_id=%s", mapID)
	return nil
}

func (r *Repository) PublishMap(ctx context.Context, mapID, ownerID, access, password string) (string, error) {
	var slug *string
	err := r.pool.QueryRow(ctx, "SELECT slug FROM maps WHERE id = $1::uuid AND owner_id = $2::uuid", mapID, ownerID).Scan(&slug)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", ErrMapNotFound
		}
		return "", err
	}

	var finalSlug string
	if slug != nil && *slug != "" {
		finalSlug = *slug
	} else {
		buf := make([]byte, 5)
		_, _ = rand.Read(buf)
		finalSlug = hex.EncodeToString(buf)
	}

	const query = `
		UPDATE maps
		SET is_published = true, slug = $1, access = $2, password_hash = $3, updated_at = now()
		WHERE id = $4::uuid AND owner_id = $5::uuid
	`
	if _, err := r.pool.Exec(ctx, query, finalSlug, access, password, mapID, ownerID); err != nil {
		logger.Error(logger.CatDB, "ошибка публикации карты map_id=%s: %v", mapID, err)
		return "", err
	}

	logger.Info(logger.CatMap, "карта опубликована map_id=%s slug=%s access=%s", mapID, finalSlug, access)
	return finalSlug, nil
}

func (r *Repository) GetPublicBySlug(ctx context.Context, slug string) (*MapDoc, string, error) {
	const query = `
		SELECT id::text, title, slug, is_published, access, COALESCE(password_hash, ''), version, updated_at, doc
		FROM maps
		WHERE slug = $1 AND is_published = true
	`

	var (
		id, title, access, pass string
		slugPtr                 *string
		isPublished             bool
		version                 int64
		updatedAt               time.Time
		docBytes                []byte
	)

	err := r.pool.QueryRow(ctx, query, slug).Scan(
		&id, &title, &slugPtr, &isPublished, &access, &pass, &version, &updatedAt, &docBytes,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, "", ErrMapNotFound
		}
		logger.Error(logger.CatDB, "ошибка чтения публичной карты slug=%s: %v", slug, err)
		return nil, "", err
	}

	var doc MapDoc
	if err := json.Unmarshal(docBytes, &doc); err != nil {
		return nil, "", err
	}

	doc.ID = id
	doc.Title = title
	doc.Slug = slugPtr
	doc.IsPublished = isPublished
	doc.Access = access
	doc.Version = version
	doc.UpdatedAt = updatedAt

	r.enrichPanoramas(ctx, &doc)
	return &doc, pass, nil
}

func (r *Repository) UpsertPanoramaStatus(ctx context.Context, mapID string, media PanoramaMedia) error {
	const query = `
		INSERT INTO panorama_media (panorama_id, map_id, status, preview_url, full_url, width, height, error, updated_at)
		VALUES ($1::uuid, $2::uuid, $3, $4, $5, $6, $7, $8, now())
		ON CONFLICT (panorama_id) DO UPDATE SET
			status = EXCLUDED.status,
			preview_url = EXCLUDED.preview_url,
			full_url = EXCLUDED.full_url,
			width = EXCLUDED.width,
			height = EXCLUDED.height,
			error = EXCLUDED.error,
			updated_at = now()
	`
	_, err := r.pool.Exec(ctx, query,
		media.PanoramaID, mapID, media.Status, media.PreviewURL, media.FullURL, media.Width, media.Height, media.Error,
	)
	if err != nil {
		logger.Error(logger.CatDB, "ошибка сохранения статуса панорамы panorama_id=%s: %v", media.PanoramaID, err)
		return fmt.Errorf("upsert panorama status: %w", err)
	}
	return nil
}

func (r *Repository) GetPanoramaStatus(ctx context.Context, panoramaID string) (*PanoramaMedia, error) {
	const query = `
		SELECT panorama_id::text, status, preview_url, full_url, width, height, error
		FROM panorama_media
		WHERE panorama_id = $1::uuid
	`
	var m PanoramaMedia
	err := r.pool.QueryRow(ctx, query, panoramaID).Scan(
		&m.PanoramaID, &m.Status, &m.PreviewURL, &m.FullURL, &m.Width, &m.Height, &m.Error,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrPanoramaNotFound
		}
		logger.Error(logger.CatDB, "ошибка запроса статуса панорамы panorama_id=%s: %v", panoramaID, err)
		return nil, err
	}
	return &m, nil
}

func (r *Repository) enrichPanoramas(ctx context.Context, doc *MapDoc) {
	if len(doc.Panoramas) == 0 {
		return
	}

	rows, err := r.pool.Query(ctx, `
		SELECT panorama_id::text, status, preview_url, full_url, width, height
		FROM panorama_media
		WHERE map_id = $1::uuid
	`, doc.ID)
	if err != nil {
		logger.Error(logger.CatDB, "ошибка загрузки медиа-данных панорам для map_id=%s: %v", doc.ID, err)
		return
	}
	defer rows.Close()

	mediaMap := make(map[string]PanoramaMedia)
	for rows.Next() {
		var m PanoramaMedia
		if err := rows.Scan(&m.PanoramaID, &m.Status, &m.PreviewURL, &m.FullURL, &m.Width, &m.Height); err == nil {
			mediaMap[m.PanoramaID] = m
		}
	}

	for i := range doc.Panoramas {
		if m, ok := mediaMap[doc.Panoramas[i].ID]; ok {
			doc.Panoramas[i].Status = m.Status
			doc.Panoramas[i].PreviewURL = m.PreviewURL
			doc.Panoramas[i].FullURL = m.FullURL
			doc.Panoramas[i].Width = m.Width
			doc.Panoramas[i].Height = m.Height
		}
	}
}
