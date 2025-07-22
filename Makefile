PHONY: run build

build:
	CGO_ENABLED=0 GOOS=linux go build  -ldflags "-s -X main.gitSHA=`git rev-parse --short HEAD` -extldflags '-static'" -o shortly/shortly shortly/main.go 

run:
	go run shortly/main.go

up-dev-db:
	docker compose up db
	# to verify run:
	# PGPASSWORD=shortly psql -h localhost -U shortly -d shortly
