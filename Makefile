.PHONY: up
up:
	docker compose up -d

.PHONY: down
down:
	docker-compose down --volumes --remove-orphans

.PHONY: run
run:
	go run main.go