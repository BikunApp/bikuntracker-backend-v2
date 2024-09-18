up:
	docker compose up -d

down:
	docker compose down

migration:
	migrate create -seq -ext sql -dir db/migrations $(filter-out $@,$(MAKECMDGOALS))

migrate:
	go run db/migrations/migrate.go $(filter-out $@,$(MAKECMDGOALS))

migrate-down:
	go run db/migrations/migrate.go -action down $(filter-out $@,$(MAKECMDGOALS))

run:
	go run main.go