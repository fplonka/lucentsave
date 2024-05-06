-- Create a new database
CREATE DATABASE lucentsave2;

-- Switch to the new database (commands below will execute in this database)
\c lucentsave2;

-- Create a new user
CREATE USER ls2user WITH ENCRYPTED PASSWORD '$LS2_DB_PASSWORD';

-- Grant privileges to the user
GRANT ALL PRIVILEGES ON DATABASE lucentsave2 TO ls2user;

-- Set default privileges for future tables, sequences, and other objects
ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT ALL PRIVILEGES ON TABLES TO ls2user;
ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT ALL PRIVILEGES ON SEQUENCES TO ls2user;
ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT ALL PRIVILEGES ON FUNCTIONS TO ls2user;

-- Create the 'users' table
CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    email TEXT UNIQUE NOT NULL,
    password_hash TEXT NOT NULL
);

-- Create the 'posts' table
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

-- -- Insert a default user into the 'users' table; password is 123
-- INSERT INTO users (email, password_hash) VALUES
-- ('123@123.com', '$2a$12$PDbM/laBF8gyD5ECPY6AdOw2lkzFyPeaqOD29XsbTjaG28cZD.KLa');

-- -- Insert some sample data into the 'posts' table
-- INSERT INTO posts (url, title, body, is_read, is_liked, time_added, user_id) VALUES
-- ('http://example.com/post1', 'First Post', 'This is the first post', false, false, EXTRACT(EPOCH FROM NOW() - INTERVAL '1 second'), 1),
-- ('http://example.com/post2', 'Second Post', 'This is the second post', false, false, EXTRACT(EPOCH FROM NOW()), 1);

CREATE INDEX idx_users_email ON users (email);

CREATE INDEX idx_posts_is_read ON posts (is_read);

CREATE INDEX idx_posts_user_id ON posts (user_id);

-- Modify the 'posts' table to include a generated tsvector column
ALTER TABLE posts
ADD COLUMN tsvector_content tsvector
GENERATED ALWAYS AS (to_tsvector('english', coalesce(title, '') || ' ' || coalesce(url, '') || ' ' || coalesce(body, ''))) STORED;

-- Create a GIN index on the generated tsvector column
CREATE INDEX idx_posts_tsvector_content ON posts USING GIN (tsvector_content);


-- pgvector bs
CREATE EXTENSION vector;

ALTER TABLE posts
ADD COLUMN embedding vector(1536);

CREATE INDEX ON posts USING hnsw (embedding vector_ip_ops);
