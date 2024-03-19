# Remote continainers

## Description

Remote Containers is a tool that facilitates the management of Docker containers across multiple machines. It enables remote communication and control of Docker containers via a Go-based API.

## Prerequisites

- Docker installed on both machines: [Docker Installation Guide](https://docs.docker.com/get-docker/)
- Go installed on both machines: [Go Installation Guide](https://golang.org/doc/install)

## Usage

### Run the Application

1. Install Go modules:

    ```bash
    go mod download
    ```

3. Run the application:

    ```bash
    go run main.go
    ```

4. Follow the prompts:

    - When prompted, enter empty value to generate a connection key for libp2p on the first machine.
    - Use the generated key to connect the second machine to the first one.

### Accessing the API

Once the application is running, you can access the API and explore the available routes using Swagger documentation:

- Open [Swagger Documentation](http://localhost:8080/swagger/index.html) in your web browser.