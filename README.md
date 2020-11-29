# Sample Envoy XDS and External Auth Servers

Inspired by https://github.com/stevesloka/envoy-xds-server, this repo is my playground to learn how to write Envoy XDS and external authorization service.

## To Run

```sh
make cert
docker-compose up -d
```

## Test

```sh
curl -k -v -H "x-route: echo-server-1" -H "Authorization: Bearer echo-server-1-password" https://localhost:9000
curl -k -v -H "x-route: envoy-web" -H "Authorization: Bearer envoy-web-password" https://localhost:9000
```