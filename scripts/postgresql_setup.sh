#! /bin/bash


# before running setup a password for the posgres linux user

# setup password for posgres database user
psql -c "ALTER USER postgres with PASSWORD 'password123';"
psql -c "ALTER DATABASE temporal OWNER TO temporal;"
createdb temporal

#CREATE TABLE uploads (id intcreated_at timestamptz, updated_at timestamptz, deleted_at timestamptz, hash varchar, type varchar, hold_time_in_months int, upload_address varchar)