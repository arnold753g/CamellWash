#!/bin/sh
# Run after migrations, using the database owner's PG* connection variables.
# The backend must use DB_APP_USER / DB_APP_PASSWORD instead of that owner.
set -eu

: "${PGHOST:?PGHOST is required}"
: "${PGPORT:?PGPORT is required}"
: "${PGDATABASE:?PGDATABASE is required}"
: "${PGUSER:?PGUSER is required}"
: "${PGPASSWORD:?PGPASSWORD is required}"
: "${DB_APP_USER:?DB_APP_USER is required}"
: "${DB_APP_PASSWORD:?DB_APP_PASSWORD is required}"

# PostgreSQL identifiers are limited to 63 bytes. Reject truncation, and never
# remove privileges from the connection owner through a misconfigured name.
LC_ALL=C
export LC_ALL
if [ "$DB_APP_USER" = "$PGUSER" ]; then
    printf '%s\n' 'DB_APP_USER must differ from the database owner.' >&2
    exit 1
fi
if [ "${#DB_APP_USER}" -gt 63 ]; then
    printf '%s\n' 'DB_APP_USER must contain at most 63 bytes.' >&2
    exit 1
fi

# psql 16 reads credentials directly from the environment. In particular, the
# runtime password is never passed as a command-line argument or echoed.
psql --no-psqlrc --no-password --quiet \
    --set=ON_ERROR_STOP=1 --set=ECHO=none \
    --set=VERBOSITY=terse --set=SHOW_CONTEXT=never <<'SQL'
\getenv app_user DB_APP_USER
\getenv app_password DB_APP_PASSWORD

BEGIN;

SELECT format('CREATE ROLE %I', :'app_user')
WHERE NOT EXISTS (SELECT 1 FROM pg_catalog.pg_roles WHERE rolname = :'app_user')
\gexec

SELECT format(
    'ALTER ROLE %I WITH LOGIN NOSUPERUSER NOCREATEDB NOCREATEROLE NOREPLICATION NOBYPASSRLS NOINHERIT PASSWORD %L',
    :'app_user', :'app_password'
)
\gexec

SELECT format('GRANT CONNECT ON DATABASE %I TO %I', current_database(), :'app_user')
\gexec
SELECT format('REVOKE CREATE ON SCHEMA public FROM %I', :'app_user')
\gexec
SELECT format('GRANT USAGE ON SCHEMA public TO %I', :'app_user')
\gexec
SELECT format('GRANT SELECT, INSERT, UPDATE, DELETE ON ALL TABLES IN SCHEMA public TO %I', :'app_user')
\gexec
SELECT format('GRANT USAGE, SELECT ON ALL SEQUENCES IN SCHEMA public TO %I', :'app_user')
\gexec

-- Migrations and this one-shot must connect as the same owner so that newly
-- created application tables and sequences inherit the runtime grants.
SELECT format(
    'ALTER DEFAULT PRIVILEGES FOR ROLE %I IN SCHEMA public GRANT SELECT, INSERT, UPDATE, DELETE ON TABLES TO %I',
    current_user, :'app_user'
)
\gexec
SELECT format(
    'ALTER DEFAULT PRIVILEGES FOR ROLE %I IN SCHEMA public GRANT USAGE, SELECT ON SEQUENCES TO %I',
    current_user, :'app_user'
)
\gexec

-- The runtime can read migration status but must never rewrite migration state.
SELECT format(
    'REVOKE INSERT, UPDATE, DELETE, TRUNCATE, REFERENCES, TRIGGER ON TABLE public.schema_migrations FROM %I',
    :'app_user'
)
WHERE to_regclass('public.schema_migrations') IS NOT NULL
\gexec

COMMIT;
SQL

printf '%s\n' 'Runtime database permissions configured.'
