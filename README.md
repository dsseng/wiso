# Minimalist captive portal with advanced features

```
docker build -t ghcr.io/dsseng/wiso-freeradius:latest contrib/radius
docker build -t ghcr.io/dsseng/wiso-postgres:latest contrib/postgres

docker run --name some-postgres -e POSTGRES_PASSWORD=mysecretpassword -d ghcr.io/dsseng/wiso-postgres:latest
docker run --name some-radius -p 1812-1813:1812-1813/udp --tmpfs /var/log/radius -d ghcr.io/dsseng/wiso-freeradius:latest
```

This project is developed, tested and deployed with MikroTik RouterOS-based hardware.
Refer to `mikrotik-hotspot.rsc` for example on how to configure the hotspot for authentication.
