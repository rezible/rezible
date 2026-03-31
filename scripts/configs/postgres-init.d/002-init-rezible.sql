CREATE USER rez_migrator WITH LOGIN CREATEDB PASSWORD 'rez_migrator';
CREATE USER rez_app WITH LOGIN PASSWORD 'rez_app';
CREATE USER rez_documents WITH LOGIN PASSWORD 'rez_documents';

CREATE DATABASE rezible OWNER rez_migrator;
REVOKE ALL ON DATABASE rezible FROM PUBLIC;

GRANT CONNECT ON DATABASE rezible TO rez_app;
GRANT CONNECT ON DATABASE rezible TO rez_documents;

/*
\connect rezible

CREATE SCHEMA IF NOT EXISTS rezible AUTHORIZATION rez_migrator;

GRANT USAGE ON SCHEMA rezible TO rez_app;
GRANT USAGE ON SCHEMA rezible TO rez_documents;

-- Default privileges: rez_migrator-created objects in rezible schema
ALTER DEFAULT PRIVILEGES FOR USER rez_migrator IN SCHEMA rezible
    GRANT SELECT, INSERT, UPDATE, DELETE ON TABLES TO rez_app;
ALTER DEFAULT PRIVILEGES FOR USER rez_migrator IN SCHEMA rezible
    GRANT USAGE, SELECT ON SEQUENCES TO rez_app;

ALTER DEFAULT PRIVILEGES FOR USER rez_migrator IN SCHEMA rezible
    GRANT SELECT, INSERT, UPDATE, DELETE ON TABLES TO rez_documents;
ALTER DEFAULT PRIVILEGES FOR USER rez_migrator IN SCHEMA rezible
    GRANT USAGE, SELECT ON SEQUENCES TO rez_documents;

-- Set default search paths
ALTER ROLE rez_migrator SET search_path TO rezible, public;
ALTER ROLE rez_app SET search_path TO rezible, public;
ALTER ROLE rez_documents SET search_path TO rezible, public;
*/