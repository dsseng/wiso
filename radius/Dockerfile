FROM fedora:39

RUN dnf update -y && dnf install -y freeradius freeradius-postgresql freeradius-utils && rm -rf /var/cache/dnf
RUN rm -rf /etc/raddb
COPY ./raddb /etc/raddb/

ENTRYPOINT ["/usr/sbin/radiusd", "-X"]
