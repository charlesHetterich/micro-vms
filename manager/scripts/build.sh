GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o ./scripts/bin/guest-go-init.amd64 ./guest
go build -o ./dist/microvm .

cp ./dist/microvm /usr/local/bin/vm