# Stage 1: Build Vue.js app
FROM node:16 AS build

WORKDIR /app

COPY frontend/package*.json ./frontend/
RUN cd frontend && npm install
COPY frontend /app/frontend
RUN cd frontend && npm run build

# Stage 2: Build Go server
FROM golang:1.22.4

# Install OpenCV 4.10.0
RUN apt-get update && apt-get install -y cmake g++ wget unzip
RUN wget -O opencv.zip https://github.com/opencv/opencv/archive/4.10.0.zip
RUN wget -O opencv_contrib.zip https://github.com/opencv/opencv_contrib/archive/4.10.0.zip
RUN unzip opencv.zip
RUN unzip opencv_contrib.zip
RUN mkdir -p build
WORKDIR ./build
RUN ls
RUN cmake -DOPENCV_EXTRA_MODULES_PATH=../opencv_contrib-4.10.0/modules ../opencv-4.10.0
RUN cmake --build .

# Build the go app
WORKDIR /app
COPY main.go go.mod go.sum .
COPY --from=build /app/frontend/dist /app/frontend/dist
RUN go build -o server main.go

EXPOSE 8080
CMD ["./server"]
