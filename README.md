# Sample Envoy XDS and External Auth Servers

Inspired by https://github.com/stevesloka/envoy-xds-server, this repo is my playground to learn how to write Envoy XDS and external authorization service.

## To Run

```sh
docker-compose up -d
```

After all the containers are running, find out the container IP for `echo-server` by below command, and update `hack/xds-config.yaml` accordingly.
The XDS should automatically reload upon file changes.

```sh
docker ps | grep jmalloc/echo-server | awk '{print $1}' | xargs -L 1 docker inspect -f '{{range .NetworkSettings.Networks}}{{println .IPAddress}}{{end}}'
```

The reason that this is needed is because you cannot use DNS name in EDS. So for now you have to find the IP of `echo-server`

## Test

```sh
curl -v -H "Authorization: Bearer foobar2" http://localhost:9000
```