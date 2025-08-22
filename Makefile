default:
	GOOS=linux GOARCH=amd64 go build -o ./aicommit ./main.go

windows:
	GOOS=windows GOARCH=amd64 go build -o ./aicommit ./main.go

run:
	go run ./main.go

test:
	go test ./...

test-cover:
	go test ./... -cover

symlink:
	sudo cp -f ./aicommit /usr/local/bin
