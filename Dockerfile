FROM golang:1.21 as builder

COPY . /app
WORKDIR /app
RUN CGO_ENABLED=0 GOOS=linux go build .

FROM alpine

COPY --from=builder /app/overtrakt /app/overtrakt
RUN mkdir /app/database
COPY migrate /app/migrate

EXPOSE 6868

CMD ["/app/overtrakt"]
