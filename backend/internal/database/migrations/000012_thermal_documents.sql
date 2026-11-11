CREATE TABLE documents (
    id BIGSERIAL PRIMARY KEY,
    document_key TEXT NOT NULL UNIQUE,
    kind TEXT NOT NULL CHECK (kind IN ('RECEIPT', 'RECEPTION')),
    order_id BIGINT NOT NULL REFERENCES orders(id),
    payment_id BIGINT REFERENCES payments(id),
    snapshot JSONB NOT NULL,
    issued_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    issued_by_user_id BIGINT NOT NULL REFERENCES users(id),
    CHECK ((kind = 'RECEIPT' AND payment_id IS NOT NULL) OR (kind = 'RECEPTION' AND payment_id IS NULL))
);
CREATE UNIQUE INDEX documents_payment_unique ON documents(payment_id) WHERE payment_id IS NOT NULL;
CREATE INDEX documents_order_idx ON documents(order_id, id);
CREATE TABLE document_print_attempts (
    id BIGSERIAL PRIMARY KEY,
    document_id BIGINT NOT NULL REFERENCES documents(id),
    request_key TEXT NOT NULL,
    copy_number INTEGER NOT NULL CHECK (copy_number >= 0),
    actor_user_id BIGINT NOT NULL REFERENCES users(id),
    requested_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(document_id, request_key),
    UNIQUE(document_id, copy_number)
);
