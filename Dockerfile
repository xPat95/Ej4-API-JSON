FROM golang:1.22-alpine

WORKDIR /app

COPY . .

RUN go build -o server .

EXPOSE 24355

CMD ["./server"]
