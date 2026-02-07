-- Docker entrypoint runs this as POSTGRES_USER in POSTGRES_DB,
-- so no need for CREATE DATABASE / CREATE USER / GRANT.

CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    email TEXT UNIQUE NOT NULL,
    password_hash TEXT NOT NULL
);

CREATE TABLE posts (
    id SERIAL PRIMARY KEY,
    url TEXT NOT NULL,
    title TEXT NOT NULL,
    body TEXT NOT NULL,
    is_read BOOLEAN DEFAULT false,
    is_liked BOOLEAN DEFAULT false,
    time_added BIGINT,
    user_id INTEGER NOT NULL,
    FOREIGN KEY (user_id) REFERENCES users(id)
);

CREATE INDEX idx_users_email ON users (email);
CREATE INDEX idx_posts_is_read ON posts (is_read);
CREATE INDEX idx_posts_user_id ON posts (user_id);

ALTER TABLE posts
ADD COLUMN tsvector_content tsvector
GENERATED ALWAYS AS (to_tsvector('english', coalesce(title, '') || ' ' || coalesce(url, '') || ' ' || coalesce(body, ''))) STORED;

CREATE INDEX idx_posts_tsvector_content ON posts USING GIN (tsvector_content);

CREATE EXTENSION vector;

ALTER TABLE posts
ADD COLUMN embedding vector(1536);

CREATE INDEX ON posts USING hnsw (embedding vector_ip_ops);
