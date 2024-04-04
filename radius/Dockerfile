FROM fedora:39

RUN dnf update -y && dnf install -y freeradius freeradius-postgresql freeradius-utils && rm -rf /var/cache/dnf
COPY raddb /etc/raddb
