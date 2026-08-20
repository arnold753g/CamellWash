CREATE OR REPLACE FUNCTION normalize_client_identifier(value TEXT)
RETURNS TEXT
LANGUAGE SQL
IMMUTABLE
PARALLEL SAFE
RETURN NULLIF(UPPER(REGEXP_REPLACE(BTRIM(value), '[[:space:]-]+', '', 'g')), '');

CREATE TABLE clients (
    id BIGSERIAL PRIMARY KEY,
    full_name VARCHAR(150) NOT NULL,
    ci VARCHAR(30),
    normalized_ci VARCHAR(30) GENERATED ALWAYS AS (normalize_client_identifier(ci)) STORED,
    phone VARCHAR(16) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT clients_full_name_length CHECK (CHAR_LENGTH(BTRIM(full_name)) BETWEEN 2 AND 150),
    CONSTRAINT clients_ci_length CHECK (normalized_ci IS NULL OR CHAR_LENGTH(normalized_ci) BETWEEN 4 AND 20),
    CONSTRAINT clients_ci_placeholder_reserved CHECK (normalized_ci IS NULL OR normalized_ci <> '00000000'),
    CONSTRAINT clients_phone_format CHECK (phone ~ '^\+?[0-9]{7,15}$')
);

CREATE UNIQUE INDEX clients_normalized_ci_unique
    ON clients (normalized_ci)
    WHERE normalized_ci IS NOT NULL;
CREATE INDEX idx_clients_full_name ON clients (LOWER(full_name));
CREATE INDEX idx_clients_phone ON clients (phone);
CREATE TRIGGER clients_set_updated_at
BEFORE UPDATE ON clients
FOR EACH ROW EXECUTE FUNCTION set_updated_at();

CREATE TABLE vehicles (
    id BIGSERIAL PRIMARY KEY,
    client_id BIGINT NOT NULL,
    vehicle_type_id BIGINT NOT NULL,
    plate VARCHAR(20),
    normalized_plate VARCHAR(20) GENERATED ALWAYS AS (normalize_client_identifier(plate)) STORED,
    brand VARCHAR(100) NOT NULL DEFAULT '',
    model VARCHAR(100) NOT NULL DEFAULT '',
    color VARCHAR(100) NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT vehicles_client_fk FOREIGN KEY (client_id) REFERENCES clients(id) ON DELETE RESTRICT,
    CONSTRAINT vehicles_vehicle_type_fk FOREIGN KEY (vehicle_type_id) REFERENCES vehicle_types(id) ON DELETE RESTRICT,
    CONSTRAINT vehicles_plate_length CHECK (normalized_plate IS NULL OR CHAR_LENGTH(normalized_plate) BETWEEN 3 AND 15),
    CONSTRAINT vehicles_brand_length CHECK (CHAR_LENGTH(brand) <= 100),
    CONSTRAINT vehicles_model_length CHECK (CHAR_LENGTH(model) <= 100),
    CONSTRAINT vehicles_color_length CHECK (CHAR_LENGTH(color) <= 100)
);

CREATE UNIQUE INDEX vehicles_normalized_plate_unique
    ON vehicles (normalized_plate)
    WHERE normalized_plate IS NOT NULL;
CREATE INDEX idx_vehicles_client_created ON vehicles (client_id, created_at DESC);
CREATE INDEX idx_vehicles_vehicle_type ON vehicles (vehicle_type_id);
CREATE TRIGGER vehicles_set_updated_at
BEFORE UPDATE ON vehicles
FOR EACH ROW EXECUTE FUNCTION set_updated_at();
