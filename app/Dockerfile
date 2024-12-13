FROM golang:1.22

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .


COPY templates/ /app/templates/
COPY static/ /app/static/


RUN go build -o main .

EXPOSE 8080

CMD ["./main", "0.0.0.0"]
