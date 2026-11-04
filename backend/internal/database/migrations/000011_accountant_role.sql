-- Extend the role constraint without changing existing users or their roles.
ALTER TABLE users DROP CONSTRAINT users_role_check;
ALTER TABLE users
    ADD CONSTRAINT users_role_check
    CHECK (role IN ('admin', 'manager', 'wash_staff', 'accountant'));
