default:
	GOOS=linux GOARCH=amd64 go build -o ./aicommit ./main.go

windows:
	GOOS=windows GOARCH=amd64 go build -o ./aicommit ./main.go

release:
	GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -trimpath -o ./aicommit ./main.go

run:
	go run ./main.go

test:
	go test ./...

test-cover:
	go test ./... -cover

symlink:
	sudo cp -f ./aicommit /usr/local/bin
