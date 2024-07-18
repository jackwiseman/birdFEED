FROM gocv/opencv:4.10.0

# Build the go app
WORKDIR /app
COPY main.go go.mod go.sum .
COPY --from=build /app/frontend/dist /app/frontend/dist
RUN go mod download
RUN go build -o birdFEED .

EXPOSE 8080
CMD ["./birdFEED"]
