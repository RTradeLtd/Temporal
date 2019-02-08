#! /bin/bash


# before running setup a password for the posgres linux user

# setup password for posgres database user
createdb temporal
psql -c "ALTER USER postgres with PASSWORD 'password123';"
psql -c "ALTER DATABASE temporal OWNER TO temporal;"