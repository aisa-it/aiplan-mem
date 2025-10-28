build:
	GOOS=linux CGO_ENABLED=0 go build -o bin/app -ldflags="-s -w" cmd/aiplan-mem/main.go
