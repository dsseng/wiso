# Minimalist captive portal with advanced features

```
docker build -t ghcr.io/dsseng/freeradius:latest .

docker run --name radius-postgres -e POSTGRES_PASSWORD=mysecretpassword -d postgres
docker run --name radius -p 1812-1813:1812-1813/udp --rm --tmpfs /var/log/radius ghcr.io/dsseng/freeradius:latest
```

This project is developed, tested and deployed with MikroTik RouterOS-based hardware.
Refer to `mikrotik-hotspot.rsc` for example on how to configure the hotspot for authentication.
