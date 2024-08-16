-- Drop tables in reverse order of creation
-- Ensure to drop tables that reference other tables first

-- Drop tables with foreign key constraints
DROP TABLE IF EXISTS reservation_seat;
DROP TABLE IF EXISTS reservations;
DROP TABLE IF EXISTS movie_screen;
DROP TABLE IF EXISTS booked_seats;

-- Drop tables with foreign keys
DROP TABLE IF EXISTS seats;
DROP TABLE IF EXISTS shows;
DROP TABLE IF EXISTS screens;
DROP TABLE IF EXISTS theatres;

-- Drop tables that are not referenced by others
DROP TABLE IF EXISTS tokens;
DROP TABLE IF EXISTS users_permissions;
DROP TABLE IF EXISTS permissions;
DROP TABLE IF EXISTS users;
DROP TABLE IF EXISTS movies;

-- Drop indexes if they were created
DROP INDEX IF EXISTS movies_title_idx;
DROP INDEX IF EXISTS movies_genres_idx;

-- Drop constraints if they were added
ALTER TABLE movies DROP CONSTRAINT IF EXISTS movies_runtime_check;
ALTER TABLE movies DROP CONSTRAINT IF EXISTS movies_year_check;
ALTER TABLE movies DROP CONSTRAINT IF EXISTS genres_length_check;
