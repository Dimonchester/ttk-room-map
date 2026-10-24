package space

import "time"

type Pt []float64

type Plan struct {
	Width        float64 `json:"width"`
	Height       float64 `json:"height"`
	ImageURL     *string `json:"imageUrl,omitempty"`
	ImageOpacity float64 `json:"imageOpacity"`
}

type Room struct {
	ID      string `json:"id"`
	Name    string `json:"name,omitempty"`
	Polygon []Pt   `json:"polygon"`
}

type Path struct {
	ID     string  `json:"id"`
	Points []Pt    `json:"points"`
	Width  float64 `json:"width"`
}

type Poi struct {
	ID          string  `json:"id"`
	Title       string  `json:"title"`
	Category    string  `json:"category"`
	Description string  `json:"description,omitempty"`
	X           float64 `json:"x"`
	Y           float64 `json:"y"`
}

type Floor struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Plan  Plan   `json:"plan"`
	Rooms []Room `json:"rooms"`
	Paths []Path `json:"paths"`
	Pois  []Poi  `json:"pois"`
}

type Hotspot struct {
	ID               string  `json:"id"`
	Kind             string  `json:"kind"`
	Yaw              float64 `json:"yaw"`
	Pitch            float64 `json:"pitch"`
	PoiID            string  `json:"poiId,omitempty"`
	TargetPanoramaID string  `json:"targetPanoramaId,omitempty"`
}

type Panorama struct {
	ID         string    `json:"id"`
	FloorID    string    `json:"floorId"`
	Title      string    `json:"title,omitempty"`
	X          float64   `json:"x"`
	Y          float64   `json:"y"`
	North      float64   `json:"north"`
	Hotspots   []Hotspot `json:"hotspots"`
	Status     string    `json:"status,omitempty"`
	PreviewURL *string   `json:"previewUrl,omitempty"`
	FullURL    *string   `json:"fullUrl,omitempty"`
	Width      *int      `json:"width,omitempty"`
	Height     *int      `json:"height,omitempty"`
}

type MapDoc struct {
	ID          string     `json:"id"`
	Title       string     `json:"title"`
	Slug        *string    `json:"slug,omitempty"`
	IsPublished bool       `json:"isPublished"`
	Access      string     `json:"access"`
	Version     int64      `json:"version"`
	UpdatedAt   time.Time  `json:"updatedAt"`
	Floors      []Floor    `json:"floors"`
	Panoramas   []Panorama `json:"panoramas"`
}

type MapSummary struct {
	ID           string    `json:"id"`
	Title        string    `json:"title"`
	IsPublished  bool      `json:"isPublished"`
	Slug         *string   `json:"slug,omitempty"`
	Version      int64     `json:"version"`
	UpdatedAt    time.Time `json:"updatedAt"`
	ThumbnailURL *string   `json:"thumbnailUrl,omitempty"`
}

type PlanImage struct {
	ImageURL string `json:"imageUrl"`
	Width    int    `json:"width"`
	Height   int    `json:"height"`
}

type PanoramaMedia struct {
	PanoramaID string  `json:"panoramaId"`
	Status     string  `json:"status"`
	PreviewURL *string `json:"previewUrl,omitempty"`
	FullURL    *string `json:"fullUrl,omitempty"`
	Width      *int    `json:"width,omitempty"`
	Height     *int    `json:"height,omitempty"`
	Error      *string `json:"error,omitempty"`
}
