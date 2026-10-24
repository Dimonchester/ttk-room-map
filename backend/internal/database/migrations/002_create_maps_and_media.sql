CREATE TABLE IF NOT EXISTS maps (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    owner_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    title VARCHAR(255) NOT NULL,
    slug VARCHAR(64) UNIQUE,
    is_published BOOLEAN NOT NULL DEFAULT false,
    access VARCHAR(16) NOT NULL DEFAULT 'public' CHECK (access IN ('public', 'password')),
    password_hash TEXT,
    version BIGINT NOT NULL DEFAULT 1,
    thumbnail_url TEXT,
    doc JSONB NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_maps_owner_id ON maps (owner_id);
CREATE INDEX IF NOT EXISTS idx_maps_slug ON maps (slug);

CREATE TABLE IF NOT EXISTS panorama_media (
    panorama_id UUID PRIMARY KEY,
    map_id UUID NOT NULL REFERENCES maps(id) ON DELETE CASCADE,
    status VARCHAR(16) NOT NULL DEFAULT 'processing' CHECK (status IN ('processing', 'ready', 'failed')),
    preview_url TEXT,
    full_url TEXT,
    width INT,
    height INT,
    error TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_panorama_media_map_id ON panorama_media (map_id);