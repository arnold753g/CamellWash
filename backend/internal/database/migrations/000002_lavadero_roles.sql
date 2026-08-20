DO $$
BEGIN
    IF EXISTS (
        SELECT 1
        FROM users
        WHERE role = 'user'
          AND LOWER(email) <> 'user@example.test'
    ) THEN
        RAISE EXCEPTION 'migration 000002 found an unclassified legacy user';
    END IF;
END;
$$;

ALTER TABLE users DROP CONSTRAINT users_role_check;

UPDATE users
SET role = 'manager',
    updated_at = NOW()
WHERE role = 'user'
  AND LOWER(email) = 'user@example.test';

ALTER TABLE users ALTER COLUMN role DROP DEFAULT;
ALTER TABLE users
    ADD CONSTRAINT users_role_check
    CHECK (role IN ('admin', 'manager', 'wash_staff'));
