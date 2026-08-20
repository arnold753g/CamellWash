INSERT INTO vehicle_types (name)
VALUES
    ('Moto'),
    ('Torito/Mototaxi'),
    ('Automóvil'),
    ('SUV/Camioneta'),
    ('Furgoneta'),
    ('Camión')
ON CONFLICT (normalized_name) DO NOTHING;

INSERT INTO service_categories (name)
VALUES
    ('Lavado'),
    ('Detailing')
ON CONFLICT (normalized_name) DO NOTHING;

WITH initial_services(category_name, service_name) AS (
    VALUES
        ('Lavado', 'Chasis'),
        ('Lavado', 'Básico'),
        ('Lavado', 'Premium con aspirado'),
        ('Detailing', 'Raibanizada Americana'),
        ('Detailing', 'Raibanizada China'),
        ('Detailing', 'PPF')
)
INSERT INTO services (category_id, name)
SELECT categories.id, initial_services.service_name
FROM initial_services
JOIN service_categories AS categories
  ON categories.normalized_name = normalize_catalog_name(initial_services.category_name)
ON CONFLICT (category_id, normalized_name) DO NOTHING;
