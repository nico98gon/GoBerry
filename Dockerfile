# use oficical golang image
FROM golang:1.22.4-alpine3.20

WORKDIR /app

# copy the source code
COPY . .

# download and install dependencies
RUN go get -d -v ./...

# build the go aplication
RUN go build -o api .

EXPOSE 8080

# run the executable
CMD ["./api"]