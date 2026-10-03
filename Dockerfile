FROM golang:1.22-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/todo-server main.go

FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/todo-server .
COPY ./web ./web
ENV TODO_PORT=7540
ENV TODO_DBFILE=/app/scheduler.db
EXPOSE 7540
CMD ["./todo-server"]
