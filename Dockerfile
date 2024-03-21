# Use a Go base image
FROM golang:1.22 AS builder

# Set the working directory inside the container
WORKDIR /app

# Copy the Go modules manifests
COPY go.mod go.sum ./

# Download the module dependencies
RUN go mod download

# Copy the source code into the container
COPY . .

# Build the Go application
RUN CGO_ENABLED=1 GOOS=linux go build -a -ldflags '-linkmode external -extldflags "-static"' .

# Start a new stage for the final lightweight container
FROM alpine:latest

RUN apk --no-cache add sqlite

# Set the working directory inside the container
WORKDIR /app

# Copy the compiled executable from the builder stage
COPY --from=builder /app/summarize .

# EXPOSE 8080

# Command to run the executable
ENTRYPOINT ["./summarize"]
