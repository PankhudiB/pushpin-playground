# Pushpin Playground

To start the origin server

```
docker-compose build
docker-compose up
```

Pushpin routes file:

```
*,debug go-service:8080,over_http
```

Client request to subscribe to test channel:

> websocat -v ws://localhost:7999/subscribe


To publish event from terminal to test channel:

```
curl -d '{ "items": [ { "channel": "test", "formats": {
    "ws-message": { "content": "hello there\n" } } } ] }' \
    http://localhost:5561/publish/
```

To publish event through externally triggering origin-server on `test` channel:
```
curl -v --data "updated_state" http://localhost:8080/publish
```
