FROM postgres
ENV POSTGRES_DB radius
COPY schema.sql /docker-entrypoint-initdb.d
COPY setup.sql /docker-entrypoint-initdb.d
EXPOSE 5432
