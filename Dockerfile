FROM golang:1.24.1-alpine AS builder

WORKDIR /app

RUN apk add --no-cache gcc musl-dev git

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN apk add --no-cache protobuf
RUN go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
RUN go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
RUN mkdir -p proto/gen
RUN protoc -I=./proto --go_out=./proto/gen --go_opt=paths=source_relative --go-grpc_out=./proto/gen --go-grpc_opt=paths=source_relative ./proto/*.proto

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags="-w -s" -o /go/bin/dice-game ./cmd

FROM alpine:3.18

WORKDIR /app

RUN apk --no-cache add ca-certificates tzdata

COPY --from=builder /go/bin/dice-game .

COPY --from=builder /app/config.yaml /app/config/config.yaml
COPY --from=builder /app/config.yaml /app/

COPY --from=builder /app/migrations /app/migrations

RUN mkdir -p /app/migrations
COPY --from=builder /app/migrations/* /app/migrations/

RUN adduser -D -g '' appuser
RUN chown -R appuser:appuser /app
USER appuser

EXPOSE 8080 9090

CMD ["./dice-game"]