# Build stage
FROM golang:1.26-alpine AS build

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 go build -o /app/authz-server ./cmd/server

# Runtime stage — image kecil & aman
FROM alpine:3.20

WORKDIR /app
COPY --from=build /app/authz-server /app/authz-server
COPY --from=build /app/migrations /app/migrations

ENV AUTHZ_DB_DRIVER=postgres
ENV AUTHZ_ADDR=:8080
ENV AUTHZ_AUTO_MIGRATE=true

EXPOSE 8080

CMD ["/app/authz-server"]