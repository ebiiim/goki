all: clean test build

test:
	go test -race -cover ./...

build:
	go build "-ldflags=-s -w" -trimpath -o dist/server/goki cmd/server/main.go
	cp -r api/views dist/server
	cp -r api/static dist/server
	cp config/config.json.sample dist/server/config.json.sample
	mkdir dist/server/sessions

clean:
	rm -rf dist
