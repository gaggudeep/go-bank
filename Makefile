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

proto:
	rm -f pb/*.go
	rm -f doc/swagger/*.swagger.json
	protoc --proto_path=proto --go_out=pb --go_opt=paths=source_relative \
        --go-grpc_out=pb --go-grpc_opt=paths=source_relative \
        --grpc-gateway_out=pb --grpc-gateway_opt=paths=source_relative \
        --openapiv2_out=doc/swagger --openapiv2_opt=allow_merge=true,merge_file_name=bank\
        proto/*.proto
	statik -src=./doc/swagger -dest=./doc

evans:
	evans --host localhost --port 9090 -r repl

.PHONY: postgres createdb dropdb migrate-up migrate-down migrate-up1 migrate-down1 db-docs db-schema sqlc test server mock proto