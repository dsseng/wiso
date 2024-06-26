# Minimalist captive portal with advanced features

Wiso project aims to build a network user management system including a captive portal and admin tools for network authorization. Goals of the project include maintainability, robustness, minimal and clear design and support for modern authentication technologies like SSO with OpenID Connect. The design goal is to allow customizing authentication methods and adding new ones.

## Build

```bash
docker build -t ghcr.io/dsseng/wiso-freeradius:latest contrib/radius
docker build -t ghcr.io/dsseng/wiso-postgres:latest contrib/postgres
docker build -t ghcr.io/dsseng/wiso:latest .
# To run wiso locally:
go build .
```

## Run

- `192.168.88.235` - IP of the server RADIUS and Wiso run on
- `gitea.example.com` - sample OIDC provider
- `internal_pass` - PostgreSQL password

```bash
docker network create wiso-net
docker run --net=wiso-net --name wiso-postgres -e POSTGRES_PASSWORD=internal_pass -d ghcr.io/dsseng/wiso-postgres:latest
docker run --net=wiso-net --name wiso-radius -e POSTGRES_PASSWORD=internal_pass -e RADIUS_SECRET=mikrotik -p 1812-1813:1812-1813/udp --tmpfs /opt/var/log/radius -d ghcr.io/dsseng/wiso-freeradius:latest
docker run --net=wiso-net --name wiso -v .:/conf:ro -p 8989:8989 -d ghcr.io/dsseng/wiso:latest
```

Running server natively (with PostgreSQL and FreeRADIUS running in Docker):
```bash
./wiso web -c config.yaml
```

# Stop

```bash
docker stop wiso wiso-postgres wiso-radius; docker rm wiso wiso-postgres wiso-radius
```

This project is developed, tested and deployed with MikroTik RouterOS-based hardware.
Refer to `contrib/mikrotik/hotspot.rsc` for example on how to configure the hotspot for authentication.

Support for other devices should not be problematic thanks to RADIUS being widely accepted standard.
