FROM golang:1.22.2 AS build
WORKDIR /app
COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go mod tidy
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go mod download
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o cloudrun ./cmd/stresstest

FROM scratch
WORKDIR /app
COPY --from=build /app/cloudrun .
EXPOSE 8080
ENTRYPOINT ["./cloudrun"]