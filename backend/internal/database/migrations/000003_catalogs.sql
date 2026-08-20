CREATE OR REPLACE FUNCTION normalize_catalog_name(value TEXT)
RETURNS TEXT
LANGUAGE SQL
IMMUTABLE
PARALLEL SAFE
RETURN LOWER(REGEXP_REPLACE(BTRIM(value), '[[:space:]]+', ' ', 'g'));

CREATE TABLE vehicle_types (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    normalized_name VARCHAR(100) GENERATED ALWAYS AS (normalize_catalog_name(name)) STORED,
    active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT vehicle_types_name_length CHECK (CHAR_LENGTH(normalized_name) BETWEEN 2 AND 100),
    CONSTRAINT vehicle_types_normalized_name_unique UNIQUE (normalized_name)
);

CREATE INDEX idx_vehicle_types_active_name ON vehicle_types (active, normalized_name);
CREATE TRIGGER vehicle_types_set_updated_at
BEFORE UPDATE ON vehicle_types
FOR EACH ROW EXECUTE FUNCTION set_updated_at();

CREATE TABLE service_categories (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    normalized_name VARCHAR(100) GENERATED ALWAYS AS (normalize_catalog_name(name)) STORED,
    active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT service_categories_name_length CHECK (CHAR_LENGTH(normalized_name) BETWEEN 2 AND 100),
    CONSTRAINT service_categories_normalized_name_unique UNIQUE (normalized_name)
);

CREATE INDEX idx_service_categories_active_name ON service_categories (active, normalized_name);
CREATE TRIGGER service_categories_set_updated_at
BEFORE UPDATE ON service_categories
FOR EACH ROW EXECUTE FUNCTION set_updated_at();

CREATE TABLE services (
    id BIGSERIAL PRIMARY KEY,
    category_id BIGINT NOT NULL,
    name VARCHAR(100) NOT NULL,
    normalized_name VARCHAR(100) GENERATED ALWAYS AS (normalize_catalog_name(name)) STORED,
    active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT services_category_fk FOREIGN KEY (category_id) REFERENCES service_categories(id) ON DELETE RESTRICT,
    CONSTRAINT services_name_length CHECK (CHAR_LENGTH(normalized_name) BETWEEN 2 AND 100),
    CONSTRAINT services_category_normalized_name_unique UNIQUE (category_id, normalized_name)
);

CREATE INDEX idx_services_active_category_name ON services (active, category_id, normalized_name);
CREATE TRIGGER services_set_updated_at
BEFORE UPDATE ON services
FOR EACH ROW EXECUTE FUNCTION set_updated_at();
