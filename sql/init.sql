-- Run this script to initialize the database in "postgres" database.
-- Disconnect all connections to the database "secnex_core"
SELECT pg_terminate_backend(pg_stat_activity.pid)
FROM pg_stat_activity
WHERE pg_stat_activity.datname = 'secnex_core'
  AND pid <> pg_backend_pid();

-- Drop the database "secnex_core"
DROP DATABASE IF EXISTS "secnex_core";
CREATE DATABASE "secnex_core";

-- Reconnect to the database "secnex_core" and run the following script
DROP TYPE IF EXISTS "method" CASCADE;
DROP TYPE IF EXISTS "action" CASCADE;

CREATE TYPE "method" AS ENUM ('GET', 'POST', 'PUT', 'DELETE', 'PATCH', 'OPTIONS', 'HEAD', 'CONNECT', 'TRACE');
CREATE TYPE "action" AS ENUM ('ALLOW', 'BLOCK');

CREATE TABLE servers (
    "id" UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    "name" TEXT NOT NULL,
    "address" TEXT NOT NULL,
    "port" INT NOT NULL,
    "base_path" TEXT NOT NULL DEFAULT '/api/v1',
    "created_at" TIMESTAMPTZ NOT NULL DEFAULT now(),
    "updated_at" TIMESTAMPTZ NOT NULL DEFAULT now(),
    "deleted_at" TIMESTAMPTZ
);

CREATE TABLE firewalls (
    "id" UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    "name" TEXT NOT NULL,
    "allow_all" BOOLEAN NOT NULL DEFAULT TRUE,
    "require_auth" BOOLEAN NOT NULL DEFAULT FALSE,
    "created_at" TIMESTAMPTZ NOT NULL DEFAULT now(),
    "updated_at" TIMESTAMPTZ NOT NULL DEFAULT now(),
    "deleted_at" TIMESTAMPTZ
);

CREATE TABLE "routes" (
    "id" UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    "name" TEXT NOT NULL,
    "route" TEXT NOT NULL,
    "target" TEXT NOT NULL,
    "firewall_id" UUID NOT NULL,
    "server_id" UUID,
    "global_available" BOOLEAN NOT NULL DEFAULT FALSE,
    "include_subroutes" BOOLEAN NOT NULL DEFAULT FALSE,
    "created_at" TIMESTAMPTZ NOT NULL DEFAULT now(),
    "updated_at" TIMESTAMPTZ NOT NULL DEFAULT now(),
    "deleted_at" TIMESTAMPTZ,
    FOREIGN KEY ("firewall_id") REFERENCES "firewall" ("id") ON DELETE CASCADE,
    FOREIGN KEY ("server_id") REFERENCES "servers" ("id") ON DELETE CASCADE
);

CREATE TABLE "methods" (
    "firewall_id" UUID NOT NULL,
    "route_id" UUID NOT NULL,
    "method" "method" NOT NULL,
    "action" "action" NOT NULL DEFAULT 'ALLOW',
    "created_at" TIMESTAMPTZ NOT NULL DEFAULT now(),
    "updated_at" TIMESTAMPTZ NOT NULL DEFAULT now(),
    "deleted_at" TIMESTAMPTZ,
    PRIMARY KEY ("firewall_id", "route_id", "method"),
    FOREIGN KEY ("firewall_id") REFERENCES "firewall" ("id") ON DELETE CASCADE,
    FOREIGN KEY ("route_id") REFERENCES "routes" ("id") ON DELETE CASCADE
);

CREATE TABLE "ips" (
    "firewall_id" UUID NOT NULL,
    "route_id" UUID NOT NULL,
    "ip" TEXT NOT NULL,
    "action" "action" NOT NULL DEFAULT 'ALLOW',
    "created_at" TIMESTAMPTZ NOT NULL DEFAULT now(),
    "updated_at" TIMESTAMPTZ NOT NULL DEFAULT now(),
    "deleted_at" TIMESTAMPTZ,
    PRIMARY KEY ("firewall_id", "route_id", "ip"),
    FOREIGN KEY ("firewall_id") REFERENCES "firewall" ("id") ON DELETE CASCADE,
    FOREIGN KEY ("route_id") REFERENCES "routes" ("id") ON DELETE CASCADE
);

CREATE TABLE "useragents" (
    "firewall_id" UUID NOT NULL,
    "route_id" UUID NOT NULL,
    "useragent" TEXT NOT NULL,
    "action" "action" NOT NULL DEFAULT 'ALLOW',
    "created_at" TIMESTAMPTZ NOT NULL DEFAULT now(),
    "updated_at" TIMESTAMPTZ NOT NULL DEFAULT now(),
    "deleted_at" TIMESTAMPTZ,
    PRIMARY KEY ("firewall_id", "route_id", "useragent"),
    FOREIGN KEY ("firewall_id") REFERENCES "firewall" ("id") ON DELETE CASCADE,
    FOREIGN KEY ("route_id") REFERENCES "routes" ("id") ON DELETE CASCADE
);

CREATE TABLE "auths" (
    "id" UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    "api_key" TEXT NOT NULL,
    "created_at" TIMESTAMPTZ NOT NULL DEFAULT now(),
    "updated_at" TIMESTAMPTZ NOT NULL DEFAULT now(),
    "deleted_at" TIMESTAMPTZ
);

CREATE TABLE "route_auths" (
    "route_id" UUID NOT NULL,
    "auth_id" UUID NOT NULL,
    "created_at" TIMESTAMPTZ NOT NULL DEFAULT now(),
    "updated_at" TIMESTAMPTZ NOT NULL DEFAULT now(),
    "deleted_at" TIMESTAMPTZ,
    PRIMARY KEY ("route_id", "auth_id"),
    FOREIGN KEY ("route_id") REFERENCES "routes" ("id") ON DELETE CASCADE,
    FOREIGN KEY ("auth_id") REFERENCES "auths" ("id") ON DELETE CASCADE
);