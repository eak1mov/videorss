FROM golang:1.24 AS build
WORKDIR /go/src/app
COPY settings.html ./
COPY go.mod go.sum ./
RUN go mod download
COPY *.go ./
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /go/bin/app

FROM gcr.io/distroless/static-debian12:nonroot-amd64
COPY --from=build /go/bin/app /app
CMD ["/app"]
