CREATE TABLE payments (
    id BIGSERIAL PRIMARY KEY,
    order_id BIGINT NOT NULL,
    amount NUMERIC(12,2) NOT NULL,
    method VARCHAR(30) NOT NULL,
    external_reference VARCHAR(120),
    status VARCHAR(20) NOT NULL DEFAULT 'VALID',
    idempotency_key VARCHAR(100) NOT NULL,
    delivery_requested BOOLEAN NOT NULL DEFAULT FALSE,
    collected_by_user_id BIGINT NOT NULL,
    paid_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    voided_by_user_id BIGINT,
    voided_at TIMESTAMPTZ,
    void_reason VARCHAR(500),
    review_status VARCHAR(30) NOT NULL DEFAULT 'NOT_REQUIRED',
    reviewed_by_user_id BIGINT,
    reviewed_at TIMESTAMPTZ,
    review_note VARCHAR(500),
    replaces_payment_id BIGINT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT payments_order_fk FOREIGN KEY (order_id) REFERENCES orders(id) ON DELETE RESTRICT,
    CONSTRAINT payments_collected_by_fk FOREIGN KEY (collected_by_user_id) REFERENCES users(id) ON DELETE RESTRICT,
    CONSTRAINT payments_voided_by_fk FOREIGN KEY (voided_by_user_id) REFERENCES users(id) ON DELETE RESTRICT,
    CONSTRAINT payments_reviewed_by_fk FOREIGN KEY (reviewed_by_user_id) REFERENCES users(id) ON DELETE RESTRICT,
    CONSTRAINT payments_replaces_fk FOREIGN KEY (replaces_payment_id) REFERENCES payments(id) ON DELETE RESTRICT,
    CONSTRAINT payments_amount_positive CHECK (amount > 0),
    CONSTRAINT payments_method_check CHECK (method IN ('CASH','QR','BANK_TRANSFER','CARD')),
    CONSTRAINT payments_status_check CHECK (status IN ('VALID','VOIDED')),
    CONSTRAINT payments_reference_length CHECK (external_reference IS NULL OR CHAR_LENGTH(BTRIM(external_reference)) BETWEEN 3 AND 120),
    CONSTRAINT payments_idempotency_length CHECK (CHAR_LENGTH(idempotency_key) BETWEEN 16 AND 100),
    CONSTRAINT payments_review_status_check CHECK (review_status IN ('NOT_REQUIRED','PENDING_REVIEW','REVIEWED','OBSERVED')),
    CONSTRAINT payments_void_audit_check CHECK (
        (status = 'VALID' AND voided_by_user_id IS NULL AND voided_at IS NULL AND void_reason IS NULL
            AND review_status = 'NOT_REQUIRED' AND reviewed_by_user_id IS NULL AND reviewed_at IS NULL AND review_note IS NULL)
        OR
        (status = 'VOIDED' AND voided_by_user_id IS NOT NULL AND voided_at IS NOT NULL
            AND CHAR_LENGTH(BTRIM(void_reason)) BETWEEN 5 AND 500
            AND (
                (review_status = 'PENDING_REVIEW' AND reviewed_by_user_id IS NULL AND reviewed_at IS NULL AND review_note IS NULL)
                OR
                (review_status IN ('REVIEWED','OBSERVED') AND reviewed_by_user_id IS NOT NULL AND reviewed_at IS NOT NULL)
            ))
    ),
    CONSTRAINT payments_idempotency_unique UNIQUE (idempotency_key),
    CONSTRAINT payments_replacement_unique UNIQUE (replaces_payment_id)
);

CREATE UNIQUE INDEX payments_one_valid_per_order ON payments (order_id) WHERE status = 'VALID';
CREATE INDEX idx_payments_order_time ON payments (order_id, paid_at DESC, id DESC);
CREATE INDEX idx_payments_pending_review ON payments (voided_at, id) WHERE review_status = 'PENDING_REVIEW';
CREATE TRIGGER payments_set_updated_at
BEFORE UPDATE ON payments
FOR EACH ROW EXECUTE FUNCTION set_updated_at();

CREATE TABLE payment_events (
    id BIGSERIAL PRIMARY KEY,
    payment_id BIGINT NOT NULL,
    order_id BIGINT NOT NULL,
    event_type VARCHAR(30) NOT NULL,
    actor_user_id BIGINT NOT NULL,
    occurred_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    reason VARCHAR(500),
    CONSTRAINT payment_events_payment_fk FOREIGN KEY (payment_id) REFERENCES payments(id) ON DELETE RESTRICT,
    CONSTRAINT payment_events_order_fk FOREIGN KEY (order_id) REFERENCES orders(id) ON DELETE RESTRICT,
    CONSTRAINT payment_events_actor_fk FOREIGN KEY (actor_user_id) REFERENCES users(id) ON DELETE RESTRICT,
    CONSTRAINT payment_events_type_check CHECK (event_type IN ('REGISTERED','VOIDED','CORRECTED','REVIEWED','OBSERVED')),
    CONSTRAINT payment_events_reason_length CHECK (reason IS NULL OR CHAR_LENGTH(BTRIM(reason)) BETWEEN 3 AND 500)
);

CREATE INDEX idx_payment_events_payment_time ON payment_events (payment_id, occurred_at, id);

CREATE OR REPLACE FUNCTION prevent_payment_event_mutation()
RETURNS TRIGGER
LANGUAGE plpgsql
AS $$
BEGIN
    RAISE EXCEPTION 'payment events are append-only';
END;
$$;

CREATE TRIGGER payment_events_append_only
BEFORE UPDATE OR DELETE ON payment_events
FOR EACH ROW EXECUTE FUNCTION prevent_payment_event_mutation();
