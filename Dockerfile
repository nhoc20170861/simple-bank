# Build stage
FROM golang:1.23-alpine3.19 AS builder
WORKDIR /app
COPY . .
RUN go build -o main main.go
# download go-migrate
RUN apk add curl
RUN curl -L https://github.com/golang-migrate/migrate/releases/download/v4.18.1/migrate.linux-amd64.tar.gz | tar xvz

# Run stage
FROM alpine:3.19
WORKDIR /app
COPY --from=builder /app/main .
COPY --from=builder /app/migrate ./migrate
 
COPY app.env .
COPY start.sh  .
COPY wait-for.sh .
RUN chmod +x /app/wait-for.sh
RUN chmod +x /app/start.sh
COPY db/migration ./migration


EXPOSE 4000
# CMD [ "/app/main" ]
# ENTRYPOINT [ "/app/start.sh" ]
