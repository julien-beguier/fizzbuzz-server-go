FROM golang:alpine as build

ARG bin_name=fizzbuzz-server
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o $bin_name -ldflags="-w -s" main.go

FROM scratch

COPY --from=build /app/$bin_name /usr/bin/

ENTRYPOINT ["fizzbuzz-server"]
