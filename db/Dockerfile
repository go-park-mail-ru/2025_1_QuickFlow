FROM postgres:15
COPY roles/*.sql /docker-entrypoint-initdb.d/
COPY ../deploy/sql/*.sql /docker-entrypoint-initdb.d/