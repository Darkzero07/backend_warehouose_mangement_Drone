# Use the official Golang image as a base image
FROM golang:1.24-alpine

# Set the current working directory inside the container
WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download all dependencies
RUN go mod download

# Copy the source code
COPY . .

# Expose the port the app runs on
EXPOSE 8000

# Run the application using go run
CMD ["go", "run", "main.go"]