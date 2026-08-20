CREATE INDEX idx_orders_received_at ON orders (received_at DESC, id DESC);
CREATE INDEX idx_order_status_history_status_time ON order_status_history (to_status, changed_at DESC, order_id);
CREATE INDEX idx_payments_status_paid_at ON payments (status, paid_at DESC, id DESC);
CREATE INDEX idx_payments_voided_at ON payments (voided_at DESC, id DESC) WHERE status = 'VOIDED';
