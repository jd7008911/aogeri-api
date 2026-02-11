// go.mod
module github.com/jd7008911/aogeri-api

go 1.21

require (
	github.com/go-chi/chi/v5 v5.0.10
	github.com/go-chi/cors v1.2.1
	github.com/go-playground/validator/v10 v10.15.5
	github.com/golang-jwt/jwt/v5 v5.0.0
	github.com/google/uuid v1.3.1
	github.com/jackc/pgx/v5 v5.4.3
	github.com/redis/go-redis/v9 v9.2.1
	golang.org/x/crypto v0.14.0
)

require (
	github.com/cespare/xxhash/v2 v2.2.0 // indirect
	github.com/dgryski/go-rendezvous v0.0.0-20200823014737-9f7001d12a5f // indirect
	github.com/gabriel-vasile/mimetype v1.4.2 // indirect
	github.com/go-playground/locales v0.14.1 // indirect
	github.com/go-playground/universal-translator v0.18.1 // indirect
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20221227161230-091c0ba34f0a // indirect
	github.com/jackc/puddle/v2 v2.2.1 // indirect
	github.com/leodido/go-urn v1.2.4 // indirect
	golang.org/x/net v0.10.0 // indirect
	golang.org/x/sync v0.1.0 // indirect
	golang.org/x/sys v0.13.0 // indirect
	golang.org/x/text v0.13.0 // indirect
)

// Local replace for placeholder imports used in the codebase
replace github.com/yourproject => ./
