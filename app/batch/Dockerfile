FROM golang:1.22 AS build

WORKDIR /go/src/app
COPY . .

RUN go mod download && go mod verify

RUN CGO_ENABLED=0 go build -o /go/bin/app

FROM gcr.io/distroless/static-debian12

COPY --from=build /go/bin/app /
CMD ["/app"]