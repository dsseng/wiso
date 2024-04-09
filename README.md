# Minimalist captive portal with advanced features

## Build

```bash
docker build -t ghcr.io/dsseng/wiso-freeradius:latest contrib/radius
docker build -t ghcr.io/dsseng/wiso-postgres:latest contrib/postgres
```

## Run

- `192.168.88.235` - IP of the server RADIUS and Wiso run on

```bash
docker network create wiso-net
docker run --net=wiso-net --name wiso-postgres -e POSTGRES_PASSWORD=internal_pass -d ghcr.io/dsseng/wiso-postgres:latest
docker run --net=wiso-net --name wiso-radius -e POSTGRES_PASSWORD=internal_pass -e RADIUS_SECRET=mikrotik -p 1812-1813:1812-1813/udp --tmpfs /var/log/radius -d ghcr.io/dsseng/wiso-freeradius:latest
```

```bash
export CLIENT_ID=UUID CLIENT_SECRET=gto_ ISSUER=https://gitea.example.com/
go run main.go -d "host=172.18.0.2 user=postgres password=internal_pass dbname=radius port=5432 sslmode=disable" web -u http://192.168.88.235:8989/
```

# Stop

```bash
docker stop wiso-postgres wiso-radius; docker rm wiso-postgres wiso-radius
```

This project is developed, tested and deployed with MikroTik RouterOS-based hardware.
Refer to `contrib/mikrotik/hotspot.rsc` for example on how to configure the hotspot for authentication.

Support for other devices should not be problematic thanks to RADIUS being widely accepted standard.
