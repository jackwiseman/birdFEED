FROM gocv/opencv:4.10.0

# Build the go app
WORKDIR /app
COPY * .
RUN go mod download
RUN go build -o birdFEED .

EXPOSE 8080
CMD ["./birdFEED"]
