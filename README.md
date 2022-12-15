### App to broadcast events to websocket clients via Pushpin   

--------
#### Architecture Diagram :

![](websocket-pushpin.png)


#### To start the origin server

```
docker-compose build
docker-compose up
```
-------
#### Pushpin routes file:

```
*,debug go-service:8080,over_http
```
-------
#### Sample client repo at : https://github.com/PankhudiB/websocket-client

Client request to subscribe to `test` channel:

> websocat -v ws://localhost:7999/subscribe
-------

#### To publish event to `test` channel for clients connected through `websocket` protocol :


Through Terminal :

```
curl -d '{ "items": [ { "channel": "test", "formats": {
    "ws-message": { "content": "hello there\n" } } } ] }' \
    http://localhost:5561/publish/
```

OR 

Trigger through origin-server : 
```
curl -v --data "updated_state" http://localhost:8080/publish
```
-------