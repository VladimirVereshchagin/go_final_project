# syntax=docker/dockerfile:1.10

# Build stage
FROM --platform=$BUILDPLATFORM golang:1.22 AS builder

# Set the working directory inside the container
WORKDIR /app

# Set build arguments for passing environment variables
ARG TARGETOS
ARG TARGETARCH

# Set environment variables for cross-compilation
ENV GOOS=$TARGETOS
ENV GOARCH=$TARGETARCH
ENV CGO_ENABLED=0

# Copy go.mod and go.sum files to download dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy all the source code of the application
COPY . .

# Build the Go application
RUN go build -o app ./cmd

# Final stage
FROM ubuntu:22.04

# Add metadata to the image
LABEL org.opencontainers.image.source="https://github.com/vladimirvereshchagin/scheduler"
LABEL maintainer="Vladimir Vereshchagin <vlvereschagin06@gmail.com>"

# Set the working directory for the final container
WORKDIR /app

# Copy the compiled application from the build stage
COPY --from=builder /app/app .

# Copy the web directory for the frontend
COPY --from=builder /app/web ./web

# Set environment variable for the port
ENV TODO_PORT=7540

# Run the compiled application
CMD ["./app"]