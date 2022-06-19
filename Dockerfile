# syntax=docker/dockerfile:1

#FROM nginx
#COPY ./nginx/nginx.conf /etc/nginx/conf.d/default.conf

FROM golang:1.18-alpine

#Create directory inside Docker Image
WORKDIR /app

#Copy Go module files
COPY go.mod ./
COPY go.sum ./

#Download go modules
RUN go mod download

#Copy source into Docker Image
COPY *.go ./

RUN go build -o /mta

EXPOSE 8080

CMD ["/mta"]

