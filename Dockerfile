FROM golang:alpine AS builder
WORKDIR /usr/src/app
COPY . .
RUN go build -o updatebot cmd/updatebot/main.go

FROM alpine AS runtime
COPY --from=builder /usr/src/app/updatebot /bin
ENTRYPOINT ["updatebot"]
