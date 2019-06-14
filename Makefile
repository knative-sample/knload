all: knload

knload:
	@echo "build knload"
	go build -o bin/knload cmd/main.go

