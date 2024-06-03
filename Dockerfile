# FROM golang:alpine AS builder

# WORKDIR /var/app

# COPY . .

# RUN go build cmd/stresstest/main.go

# FROM scratch

# WORKDIR /var/app

# COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
# COPY --from=builder /var/app/main .

# ENTRYPOINT [ "./main" ]

FROM golang:1.21 as build
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