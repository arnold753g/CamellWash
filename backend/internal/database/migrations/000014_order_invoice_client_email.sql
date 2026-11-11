-- Reconcile the local invoice migration formerly numbered 11 with the published role migration.
ALTER TABLE users DROP CONSTRAINT users_role_check;
ALTER TABLE users ADD CONSTRAINT users_role_check
    CHECK (role IN ('admin', 'manager', 'wash_staff', 'accountant'));

ALTER TABLE payments ADD COLUMN IF NOT EXISTS invoice_requested BOOLEAN NOT NULL DEFAULT FALSE;
ALTER TABLE orders ADD COLUMN invoice_requested BOOLEAN NOT NULL DEFAULT FALSE;
UPDATE orders o SET invoice_requested = p.invoice_requested
FROM (
    SELECT DISTINCT ON (order_id) order_id, invoice_requested
    FROM payments
    ORDER BY order_id, CASE WHEN status = 'VALID' THEN 0 ELSE 1 END, paid_at DESC, id DESC
) p WHERE p.order_id = o.id;

ALTER TABLE clients ADD COLUMN email VARCHAR(254);
ALTER TABLE clients ADD CONSTRAINT clients_email_format CHECK (
    email IS NULL OR (email = LOWER(BTRIM(email)) AND email ~ '^[^[:space:]@]+@[^[:space:]@]+$')
);
