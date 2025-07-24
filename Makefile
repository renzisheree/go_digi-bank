postgres:
	docker run --name golang-postgres -p 5432:5432 -e POSTGRES_USER=root -e POSTGRES_PASSWORD=secret -d postgres:13.21-alpine3.21

createdb: 
	docker exec -it golang-postgres createdb --username=root --owner=root digi-bank

dropdb: 
	docker exec -it golang-postgres dropdb digi-bank

migrateup:
	migrate -path db/migration -database "postgresql://root:secret@localhost:5432/digi-bank?sslmode=disable" -verbose up

migrateup1:
	migrate -path db/migration -database "postgresql://root:secret@localhost:5432/digi-bank?sslmode=disable" -verbose up 1

migratedown:
	migrate -path db/migration -database "postgresql://root:secret@localhost:5432/digi-bank?sslmode=disable" -verbose down

migratedown1:
	migrate -path db/migration -database "postgresql://root:secret@localhost:5432/digi-bank?sslmode=disable" -verbose down 1

sqlc:
	sqlc generate
test: 
	go test -v -cover ./...

server: 
	go run main.go

mock:
	mockgen -package mockdb -destination db/mock/store.go tutorial.sqlc.dev/app/db/sqlc Store

.PHONY: postgres createdb dropdb migrateup migratedown sqlc test server migrateup1 migratedown1