CREATE TABLE IF NOT EXISTS movies (
    id bigserial PRIMARY KEY,
    title text NOT NULL,
    release_date date NOT NULL,
    runtime integer NOT NULL,
    genres text[] NOT NULL
);