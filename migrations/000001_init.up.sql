-- Pulseira QR: schema do MVP (PostgreSQL 13+)
-- Tabelas: "user", category, tag
-- Ordem de execucao: este arquivo, depois seed_categories.sql

CREATE TYPE tag_status AS ENUM ('GENERATED', 'ACTIVATED', 'LOST', 'SUSPENDED');

-- "user" e palavra reservada no PostgreSQL: sempre entre aspas.
CREATE TABLE "user" (
  id          uuid         PRIMARY KEY DEFAULT gen_random_uuid(),
  name        varchar(120) NOT NULL,
  email       varchar(254) NOT NULL UNIQUE,
  phone       varchar(20),                       -- formato E.164
  consent_at  timestamptz,                       -- consentimento LGPD
  created_at  timestamptz  NOT NULL DEFAULT now()
);

CREATE TABLE category (
  id            uuid        PRIMARY KEY DEFAULT gen_random_uuid(),
  slug          varchar(40) NOT NULL UNIQUE,     -- pet | elderly | neurodivergent
  name          varchar(80) NOT NULL,
  fields_schema jsonb,                           -- formulario da categoria
  active        boolean     NOT NULL DEFAULT true,
  CONSTRAINT category_fields_schema_object
    CHECK (fields_schema IS NULL OR jsonb_typeof(fields_schema) = 'object')
);

CREATE TABLE tag (
  -- ID aleatorio de 10 caracteres, alfabeto sem 0/O/1/I/L (impresso no QR)
  id            varchar(10) PRIMARY KEY
                CHECK (id ~ '^[2-9A-HJKMNP-Z]{10}$'),
  status        tag_status  NOT NULL DEFAULT 'GENERATED',
  batch         varchar(50),
  pin_hash      varchar(255),                    -- nulo ate existir a ativacao
  user_id       uuid        REFERENCES "user"(id),
  category_id   uuid        REFERENCES category(id),
  profile_data  jsonb,                           -- dados de quem usa a pulseira
  activated_at  timestamptz,
  created_at    timestamptz NOT NULL DEFAULT now(),

  CONSTRAINT tag_profile_data_object
    CHECK (profile_data IS NULL OR jsonb_typeof(profile_data) = 'object'),

  -- GENERATED: sem dono, sem categoria, sem perfil
  CONSTRAINT tag_generated_is_empty
    CHECK (status <> 'GENERATED' OR (
      user_id IS NULL AND category_id IS NULL
      AND profile_data IS NULL AND activated_at IS NULL)),

  -- ACTIVATED: dono, categoria e perfil obrigatorios
  CONSTRAINT tag_activated_requires_profile
    CHECK (status <> 'ACTIVATED' OR (
      user_id IS NOT NULL AND category_id IS NOT NULL
      AND profile_data IS NOT NULL))
);

CREATE INDEX tag_batch_idx       ON tag (batch);
CREATE INDEX tag_status_idx      ON tag (status);
CREATE INDEX tag_user_id_idx     ON tag (user_id);
CREATE INDEX tag_category_id_idx ON tag (category_id);
