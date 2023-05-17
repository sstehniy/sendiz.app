# Set the base image to the official Go image
FROM golang:latest

# Set the working directory
WORKDIR /app

# Install Air
RUN go install github.com/cosmtrek/air@latest

# Copy the source code into the container
COPY . /app

# Copy the .air.toml configuration file into the container
COPY .air.toml /app/.air.toml

# Expose the port that the application is running on
EXPOSE 8080

# Start Air
CMD ["air"]
