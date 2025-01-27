# Use the official Golang image as the base image
FROM golang:latest

# Set the working directory inside the container
WORKDIR /app

# Copy Go module files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy the entire project into the container
COPY . .

# Build the Go application from the cmd directory
RUN go build -o main ./cmd

# Expose the port the app runs on
EXPOSE 8080

# Command to run the application
CMD ["./main"]
