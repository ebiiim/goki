DIST_DIR=dist/server
DIST_LI64=${DIST_DIR}/linux-amd64
# DIST_LA64=${DIST_DIR}/linux-arm64
DIST_DI64=${DIST_DIR}/darwin-amd64
# DIST_DA64=${DIST_DIR}/darwin-arm64

all: clean test build

test:
	go test -race -cover ./...

build: build-linux-amd64 build-darwin-amd64

build-linux-amd64:
	GOOS=linux GOARCH=amd64 go build "-ldflags=-s -w" -trimpath -o ${DIST_LI64}/goki cmd/server/main.go
	cp -r server/views ${DIST_LI64}
	cp -r server/static ${DIST_LI64}
	cp config/config.json.sample ${DIST_LI64}/config.json.sample
	mkdir -p ${DIST_LI64}/sessions

build-darwin-amd64:
	GOOS=darwin GOARCH=amd64 go build "-ldflags=-s -w" -trimpath -o ${DIST_DI64}/goki cmd/server/main.go
	cp -r server/views ${DIST_DI64}
	cp -r server/static ${DIST_DI64}
	cp config/config.json.sample ${DIST_DI64}/config.json.sample
	mkdir -p ${DIST_DI64}/sessions

clean:
	rm -rf dist
