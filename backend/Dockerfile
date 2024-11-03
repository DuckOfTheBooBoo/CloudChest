# backend/Dockerfile
# Use the Go image to build the application
FROM golang:1.22-alpine AS build

# Set working directory and copy files
WORKDIR /app
COPY . .

# Build the Go app
RUN go mod download
RUN go build -o server ./cmd/server/main.go
RUN go build -o migrator ./internal/database/migrations/migrate.go

# Use a smaller image for running the app
FROM alpine:latest

# Install ffmpeg
RUN apk add --no-cache ffmpeg
WORKDIR /root/
COPY --from=build /app/migrator .
COPY --from=build /app/server .


# Expose the port used by the Go app
EXPOSE 3000
CMD ["/bin/sh", "-c", "./migrator && ./server"]
