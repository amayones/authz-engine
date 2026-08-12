FROM golang:1.26 AS build

WORKDIR /app
COPY . .

RUN go mod tidy
RUN go build -o /app/authz-server ./cmd/server
RUN go build -o /app/authz-genkey ./cmd/genkey

FROM golang:1.26

WORKDIR /app
COPY --from=build /app/authz-server /app/authz-server
COPY --from=build /app/authz-genkey /app/authz-genkey
COPY --from=build /app/migrations /app/migrations

ENV AUTHZ_DB_DRIVER=postgres
ENV AUTHZ_ADDR=:8080
ENV AUTHZ_AUTO_MIGRATE=true

EXPOSE 8080

CMD ["/app/authz-server"]