
up:
	docker rmi -f pushpin-playground-go-service || true
	docker rmi -f pushpin-playground-pushpin || true
	docker rmi -f pushpin-playground-python-service || true
	docker-compose up
