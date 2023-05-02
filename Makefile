postgres:
	docker run --name postgres-15 -p 5432:5432 -e POSTGRES_USER=root -e POSTGRES_PASSWORD=password -d postgres:15.2-alpine

create-db:
	docker exec -it postgres-15 createdb --username=root --owner=root bank

drop-db:
	docker exec -it postgres-15 dropdb bank

migrate-up:
	migrate -path db/migration  -database "postgresql://root:password@localhost:5432/bank?sslmode=disable" -verbose up

migrate-down:
	migrate -path db/migration  -database "postgresql://root:password@localhost:5432/bank?sslmode=disable" -verbose down

sqlc:
	sqlc generate

test:
	go test -v -cover ./...

server:
	go run main.go

.PHONY: postgres createdb dropdb migrate-up migrate-down sqlc test server