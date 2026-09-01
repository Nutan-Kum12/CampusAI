-- =============================================================================
-- Migration: 000001_init_schema.up.sql
-- Description: Initial schema for CampusAI


-- Enable the pgcrypto extension for gen_random_uuid().
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

CREATE TABLE tenants (
    id         UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    name       VARCHAR(255) NOT NULL,
    slug       VARCHAR(100) UNIQUE NOT NULL, 
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- USERS
-- THREE ROLES:
--   "super_admin" — platform owner, no college scope (tenant_id IS NULL)
--   "admin"       — college staff, scoped to one college (uploads official docs)
--   "student"     — scoped to one college (asks questions, uploads personal docs)
--
-- SELF-SERVICE SIGNUP:
--   Admin registers with college name → college auto-created → admin account created
--   Student registers with college name → looks up existing college → joins it
--   (College must exist first — admin registers before students)
CREATE TABLE users (
    id            UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id     UUID        REFERENCES tenants(id) ON DELETE CASCADE,
    name          VARCHAR(255) NOT NULL,
    email         VARCHAR(255) NOT NULL,
    password_hash TEXT        NOT NULL,
    role          VARCHAR(20)  NOT NULL DEFAULT 'student'
                               CHECK (role IN ('super_admin', 'admin', 'student')),
    department    VARCHAR(100),
    semester      SMALLINT    CHECK (semester BETWEEN 1 AND 10),
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (tenant_id, email)
);

-- super_admin email must also be globally unique (they have no tenant_id)
CREATE UNIQUE INDEX idx_super_admin_email
    ON users(email)
    WHERE role = 'super_admin';


-- LOGIN: every auth request queries users by email within a tenant
CREATE INDEX idx_users_email    ON users(email);
CREATE INDEX idx_users_tenant   ON users(tenant_id);

-- COLLEGE NAME SEARCH: used at signup/login to find college by name.
CREATE INDEX idx_tenants_name_lower ON tenants(lower(name));
CREATE INDEX idx_tenants_slug       ON tenants(slug);
