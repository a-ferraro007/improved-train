# syntax=docker/dockerfile:1

#FROM nginx
#COPY ./nginx/nginx.conf /etc/nginx/conf.d/default.conf

FROM golang:1.18-alpine AS builder

WORKDIR /app

COPY . /app

RUN cd /app/cmd/mta && go build -o main.go



FROM alpine:latest

WORKDIR /app

COPY static_transit/ ./static_transit
COPY --from=builder /app/cmd/mta /app

EXPOSE 8080
CMD ["./main.go", "run"]