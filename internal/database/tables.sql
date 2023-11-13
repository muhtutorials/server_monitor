CREATE TABLE IF NOT EXISTS users (
    id BIGSERIAL PRIMARY KEY,
    first_name VARCHAR(255) NOT NULL,
    last_name VARCHAR(255) NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    password VARCHAR(60) NOT NULL,
    is_active INTEGER NOT NULL DEFAULT 0,
    access_level INTEGER NOT NULL DEFAULT 3,
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL,
    deleted_at TIMESTAMPTZ
);

CREATE TABLE IF NOT EXISTS preferences (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    value TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL
);

CREATE TABLE IF NOT EXISTS remember_me_tokens (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL,
    token VARCHAR(100) NOT NULL
);

CREATE TABLE IF NOT EXISTS sessions (
    token TEXT PRIMARY KEY,
    data BYTEA NOT NULL,
    expiry TIMESTAMPTZ NOT NULL
);
CREATE INDEX sessions_expiry_idx ON sessions (expiry);

CREATE TABLE IF NOT EXISTS hosts (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    full_name VARCHAR(255) NOT NULL,
    url VARCHAR(255) NOT NULL,
    ip VARCHAR(255) NOT NULL,
    ipv6 VARCHAR(255) NOT NULL,
    location VARCHAR(255) NOT NULL,
    os VARCHAR(255) NOT NULL,
    is_active INTEGER NOT NULL DEFAULT 1,
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL
);

CREATE TABLE IF NOT EXISTS services (
    id BIGSERIAL PRIMARY KEY,
    host_id BIGINT NOT NULL,
    host_name VARCHAR(255) NOT NULL,
    name VARCHAR(255) NOT NULL,
    icon VARCHAR(255) NOT NULL,
    schedule_number INTEGER NOT NULL DEFAULT 10,
    schedule_unit VARCHAR(2) NOT NULL DEFAULT 's',
    status VARCHAR(50) NOT NULL DEFAULT 'pending',
    is_active INTEGER NOT NULL DEFAULT 0,
    last_check TIMESTAMPTZ NOT NULL DEFAULT '0001-01-01 00:00:00',
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL,
    CONSTRAINT fk_host
        FOREIGN KEY(host_id)
            REFERENCES hosts(id)
            ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS events (
    id BIGSERIAL PRIMARY KEY,
    host_id BIGINT NOT NULL,
    service_id BIGINT NOT NULL,
    host_name VARCHAR(255) NOT NULL,
    service_name VARCHAR(255) NOT NULL,
    event_type VARCHAR(255) NOT NULL,
    message VARCHAR(512) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL
);