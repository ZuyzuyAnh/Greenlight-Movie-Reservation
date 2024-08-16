CREATE TABLE IF NOT EXISTS movies (
    id bigserial PRIMARY KEY,  
    created_at timestamp(0) with time zone NOT NULL DEFAULT NOW(),
    title text NOT NULL,
    year integer NOT NULL,
    runtime integer NOT NULL,
    genres text[] NOT NULL,
    img text,
    version integer NOT NULL DEFAULT 1
);

ALTER TABLE movies ADD CONSTRAINT movies_runtime_check CHECK (runtime >= 0);
ALTER TABLE movies ADD CONSTRAINT movies_year_check CHECK (year BETWEEN 1888 AND date_part('year', now()));
ALTER TABLE movies ADD CONSTRAINT genres_length_check CHECK (array_length(genres, 1) BETWEEN 1 AND 5);

CREATE INDEX IF NOT EXISTS movies_title_idx ON movies USING GIN (to_tsvector('simple', title));
CREATE INDEX IF NOT EXISTS movies_genres_idx ON movies USING GIN (genres);

CREATE TABLE IF NOT EXISTS users (
    id bigserial PRIMARY KEY,
    created_at timestamp(0) with time zone NOT NULL DEFAULT NOW(),
    name text NOT NULL,
    email citext UNIQUE NOT NULL,
    password_hash bytea NOT NULL,
    activated bool NOT NULL,
    version integer NOT NULL DEFAULT 1
);

CREATE TABLE IF NOT EXISTS tokens (
    hash bytea PRIMARY KEY,
    user_id bigint NOT NULL REFERENCES users ON DELETE CASCADE,
    expiry timestamp(0) with time zone NOT NULL,
    scope text NOT NULL
);

CREATE TABLE IF NOT EXISTS permissions (
    id bigserial PRIMARY KEY,
    code text NOT NULL
);
CREATE TABLE IF NOT EXISTS users_permissions (
    user_id bigint NOT NULL REFERENCES users ON DELETE CASCADE,
    permission_id bigint NOT NULL REFERENCES permissions ON DELETE CASCADE,
    PRIMARY KEY (user_id, permission_id)
);

INSERT INTO permissions (code)
VALUES 
    ('admin'),
    ('user');

CREATE TABLE IF NOT EXISTS theatres (
    id bigserial PRIMARY KEY,
    name text NOT NULL UNIQUE,
    city text NOT NULL
);

CREATE TABLE IF NOT EXISTS screens (
    id bigserial PRIMARY KEY,
    number integer NOT NULL,
    theatre_id bigint NOT NULL REFERENCES theatres ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS shows (
    id bigserial PRIMARY KEY,
    showtime TIMESTAMP NOT NULL,
    movie_id bigint NOT NULL REFERENCES movies ON DELETE CASCADE,
    screen_id bigint NOT NULL REFERENCES screens ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS seats (
    id bigserial PRIMARY KEY,
    row text NOT NULL, 
    number integer NOT NULL,
    price integer NOT NULL,
    screen_id bigint NOT NULL REFERENCES screens ON DELETE CASCADE,
    CONSTRAINT unique_row_number UNIQUE (row, number)
);

CREATE TYPE status AS ENUM ('pending', 'success');

CREATE TABLE IF NOT EXISTS reservations(
    id bigserial PRIMARY KEY,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    user_id BIGINT NOT NULL REFERENCES users ON DELETE CASCADE,
    amount BIGINT DEFAULT 0,
    status status NOT NULL DEFAULT 'pending',
    show_id BIGINT NOT NULL REFERENCES shows ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS reservation_seat(
    id bigserial PRIMARY KEY,
    reservation_id BIGINT NOT NULL REFERENCES reservations ON DELETE CASCADE,
    seat_id BIGINT NOT NULL REFERENCES seats ON DELETE CASCADE
);


CREATE TABLE IF NOT EXISTS seat_status(
    id bigserial PRIMARY KEY,
    seat_id BIGINT NOT NULL REFERENCES seats ON DELETE CASCADE,
    available BOOLEAN NOT NULL DEFAULT TRUE,
    show_id BIGINT NOT NULL REFERENCES shows ON DELETE CASCADE
);

