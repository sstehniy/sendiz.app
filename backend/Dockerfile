FROM golang:latest

ENV PROJECT_DIR=/app \
  GO111MODULE=on \
  CGO_ENABLED=0

WORKDIR /app
RUN mkdir "/build"
COPY . .
EXPOSE 8080
EXPOSE 3036
RUN go get github.com/githubnemo/CompileDaemon
RUN go install github.com/githubnemo/CompileDaemon
ENTRYPOINT CompileDaemon -build="go build -o /build/app" -command="/build/app"