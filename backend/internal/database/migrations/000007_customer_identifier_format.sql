ALTER TABLE clients
    ADD CONSTRAINT clients_ci_format
    CHECK (normalized_ci IS NULL OR normalized_ci ~ '^[A-Z0-9]{4,20}$');

ALTER TABLE vehicles
    ADD CONSTRAINT vehicles_plate_format
    CHECK (normalized_plate IS NULL OR normalized_plate ~ '^[A-Z0-9]{3,15}$');
