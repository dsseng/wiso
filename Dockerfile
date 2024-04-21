# syntax=docker/dockerfile:1
FROM golang:1.22 as build
COPY . /src
WORKDIR /src
RUN go mod tidy
RUN CGO_ENABLED=0 GOOS=linux go build -o /wiso .

FROM scratch
WORKDIR /
COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=build /wiso /wiso
ENV GIN_MODE=release
ENTRYPOINT ["/wiso", "web", "-c", "/conf/config.yaml"]
