CREATE TABLE orders (
    id BIGSERIAL PRIMARY KEY,
    vehicle_id BIGINT NOT NULL,
    registered_by_user_id BIGINT NOT NULL,
    received_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    estimated_delivery_at TIMESTAMPTZ NOT NULL,
    status VARCHAR(30) NOT NULL DEFAULT 'PENDING',
    total NUMERIC(12,2) NOT NULL,
    version BIGINT NOT NULL DEFAULT 1,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT orders_vehicle_fk FOREIGN KEY (vehicle_id) REFERENCES vehicles(id) ON DELETE RESTRICT,
    CONSTRAINT orders_registered_by_fk FOREIGN KEY (registered_by_user_id) REFERENCES users(id) ON DELETE RESTRICT,
    CONSTRAINT orders_estimated_delivery CHECK (estimated_delivery_at >= received_at),
    CONSTRAINT orders_status_check CHECK (status IN ('PENDING','IN_PROGRESS','QUALITY_CONTROL','FINISHED','DELIVERED','CANCELLED')),
    CONSTRAINT orders_total_positive CHECK (total > 0),
    CONSTRAINT orders_version_positive CHECK (version > 0)
);

CREATE INDEX idx_orders_status_received ON orders (status, received_at DESC);
CREATE INDEX idx_orders_vehicle_received ON orders (vehicle_id, received_at DESC);
CREATE INDEX idx_orders_registered_by ON orders (registered_by_user_id, received_at DESC);
CREATE INDEX idx_orders_estimated_delivery ON orders (estimated_delivery_at);
CREATE TRIGGER orders_set_updated_at
BEFORE UPDATE ON orders
FOR EACH ROW EXECUTE FUNCTION set_updated_at();

CREATE TABLE order_services (
    id BIGSERIAL PRIMARY KEY,
    order_id BIGINT NOT NULL,
    service_id BIGINT NOT NULL,
    service_name_snapshot VARCHAR(100) NOT NULL,
    price_snapshot NUMERIC(12,2) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT order_services_order_fk FOREIGN KEY (order_id) REFERENCES orders(id) ON DELETE RESTRICT,
    CONSTRAINT order_services_service_fk FOREIGN KEY (service_id) REFERENCES services(id) ON DELETE RESTRICT,
    CONSTRAINT order_services_unique UNIQUE (order_id, service_id),
    CONSTRAINT order_services_name_snapshot CHECK (CHAR_LENGTH(BTRIM(service_name_snapshot)) BETWEEN 2 AND 100),
    CONSTRAINT order_services_price_positive CHECK (price_snapshot > 0)
);
CREATE INDEX idx_order_services_service ON order_services (service_id, order_id);

CREATE TABLE order_staff (
    order_id BIGINT NOT NULL,
    user_id BIGINT NOT NULL,
    assigned_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (order_id, user_id),
    CONSTRAINT order_staff_order_fk FOREIGN KEY (order_id) REFERENCES orders(id) ON DELETE RESTRICT,
    CONSTRAINT order_staff_user_fk FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE RESTRICT
);
CREATE INDEX idx_order_staff_user ON order_staff (user_id, order_id);

CREATE TABLE order_status_history (
    id BIGSERIAL PRIMARY KEY,
    order_id BIGINT NOT NULL,
    from_status VARCHAR(30),
    to_status VARCHAR(30) NOT NULL,
    changed_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    actor_user_id BIGINT NOT NULL,
    reason VARCHAR(500),
    CONSTRAINT order_status_history_order_fk FOREIGN KEY (order_id) REFERENCES orders(id) ON DELETE RESTRICT,
    CONSTRAINT order_status_history_actor_fk FOREIGN KEY (actor_user_id) REFERENCES users(id) ON DELETE RESTRICT,
    CONSTRAINT order_status_history_from_check CHECK (from_status IS NULL OR from_status IN ('PENDING','IN_PROGRESS','QUALITY_CONTROL','FINISHED','DELIVERED','CANCELLED')),
    CONSTRAINT order_status_history_to_check CHECK (to_status IN ('PENDING','IN_PROGRESS','QUALITY_CONTROL','FINISHED','DELIVERED','CANCELLED')),
    CONSTRAINT order_status_history_transition CHECK (from_status IS NULL OR from_status <> to_status),
    CONSTRAINT order_status_history_reason_length CHECK (reason IS NULL OR CHAR_LENGTH(BTRIM(reason)) BETWEEN 3 AND 500)
);
CREATE INDEX idx_order_status_history_order_time ON order_status_history (order_id, changed_at, id);

CREATE OR REPLACE FUNCTION prevent_order_history_mutation()
RETURNS TRIGGER
LANGUAGE plpgsql
AS $$
BEGIN
    RAISE EXCEPTION 'order status history is append-only';
END;
$$;

CREATE TRIGGER order_status_history_append_only
BEFORE UPDATE OR DELETE ON order_status_history
FOR EACH ROW EXECUTE FUNCTION prevent_order_history_mutation();
