db_login:
	psql ${DB_URL}

db_create_migration:
	migrate create -ext sql -dir db/migrations -seq ${NAME}

db_migrate:
	migrate -database ${DB_URL} -path db/migrations up