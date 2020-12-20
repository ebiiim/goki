FROM golang:1.15-buster as builder
WORKDIR /go/src/app
COPY . .
COPY server/static/ dist/static/
COPY server/views/ dist/views/
COPY config/config.json.sample dist/config.json
RUN CGO_ENABLED=0 go build "-ldflags=-s -w" -trimpath -o dist/main cmd/pod/main.go


FROM alpine:3.12
COPY --from=builder /go/src/app/dist/ ./
ENTRYPOINT [ "./main" ]
