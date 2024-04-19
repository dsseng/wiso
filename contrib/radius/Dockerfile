FROM freeradius/freeradius-server:latest-3.2-alpine

RUN apk add --no-cache libpq freeradius-utils

COPY ./raddb /etc/raddb/
