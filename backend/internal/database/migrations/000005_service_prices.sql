CREATE TABLE service_prices (
    id BIGSERIAL PRIMARY KEY,
    service_id BIGINT NOT NULL,
    vehicle_type_id BIGINT NOT NULL,
    price NUMERIC(12, 2) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT service_prices_service_fk
        FOREIGN KEY (service_id) REFERENCES services(id) ON DELETE RESTRICT,
    CONSTRAINT service_prices_vehicle_type_fk
        FOREIGN KEY (vehicle_type_id) REFERENCES vehicle_types(id) ON DELETE RESTRICT,
    CONSTRAINT service_prices_positive_price CHECK (price > 0),
    CONSTRAINT service_prices_service_vehicle_unique UNIQUE (service_id, vehicle_type_id)
);

CREATE INDEX idx_service_prices_vehicle_type_service
    ON service_prices (vehicle_type_id, service_id);

CREATE TRIGGER service_prices_set_updated_at
BEFORE UPDATE ON service_prices
FOR EACH ROW EXECUTE FUNCTION set_updated_at();
