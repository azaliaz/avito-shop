.PHONY: deploy rollback test coverage lint

deploy:
	docker compose --file ./deploy/docker/docker-compose.yml  up -d

rollback:
	docker compose --file ./deploy/docker/docker-compose.yml  down
	docker rmi docker-migration docker-avito-shop

unit_test:
	go test ./internal/application/tests ./internal/facade/rest/tests


integration_tests: 
	go test -v ./internal/storage/tests

cover-html:
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out
	rm coverage.out

cover:
	go test -coverprofile=coverage.out ./...
	go tool cover -func=coverage.out
	rm coverage.out

lint:
	go mod vendor
	docker run --rm -v $(shell pwd):/work:ro -w /work golangci/golangci-lint:latest golangci-lint run -v