DB_URL=postgresql://root:password@localhost:5432/bank?sslmode=disable

postgres:
	docker run --name postgres-15 -p 5432:5432 -e POSTGRES_USER=root -e POSTGRES_PASSWORD=password -d postgres:15-alpine

create-db:
	docker exec -it postgres-15 createdb --username=root --owner=root bank

drop-db:
	docker exec -it postgres-15 dropdb bank

migrate-up:
	migrate -path db/migration  -database "${DB_URL}" -verbose up

migrate-up1:
	migrate -path db/migration  -database "${DB_URL}" -verbose up 1

migrate-down:
	migrate -path db/migration  -database "${DB_URL}" -verbose down

migrate-down1:
	migrate -path db/migration  -database "postgresql://root:password@localhost:5432/bank?sslmode=disable" -verbose down 1

db-docs:
	dbdocs build doc/db.dbml

db-schema:
	 dbml2sql --postgres -o doc/schema.sql doc/db.dbml

sqlc:
	sqlc generate

test:
	go test -v -cover ./...

server:
	go run main.go

mock:
	mockgen -destination db/mock/store.go --build_flags=--mod=mod -package mockdb  github.com/gaggudeep/bank_go/db/sqlc Store

.PHONY: postgres createdb dropdb migrate-up migrate-down migrate-up1 migrate-down1 db-docs db-schema sqlc test server mock