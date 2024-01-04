all:	format tidy build test

build: tidy
	env GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" ./...

format:
	gofmt -s -w -l .

test:
	go test ./...

tidy:
	go mod tidy
