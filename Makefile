
up:
	docker rmi -f pushpin-playground-go-service || true
	docker rmi -f pushpin-playground-pushpin || true
	docker-compose up
