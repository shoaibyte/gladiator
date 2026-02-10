.PHONY: dev dev-build dev-down dev-clean install-deps generate build test test-coverage lint docker-build

dev:
	docker-compose up

dev-build:
	docker-compose up --build

dev-down:
	docker-compose down

dev-clean:
	docker-compose down -v

install-deps:
	go mod download
	cd frontend && npm install

generate:
	go generate ./ent

build:
	cd frontend && npm run build
	rm -rf cmd/server/frontend_dist && cp -r frontend/dist cmd/server/frontend_dist
	go build -o bin/server ./cmd/server

test:
	go test -v ./...

test-coverage:
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out

lint:
	golangci-lint run ./...

docker-build:
	docker build -f docker/Dockerfile -t gladiator:latest .
