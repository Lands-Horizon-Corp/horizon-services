To add prerequisites to your README for your project, mentioning the required versions of Go and Docker CLI is essential for others to replicate your setup successfully. Here's how you can phrase it:

### README.md

#### Prerequisites

Make sure you have the following installed before proceeding:

* **Go 1.24.2**: Ensure you have Go version 1.24.2 or later installed. You can download it from [here](https://golang.org/dl/).
* **Docker CLI**: Install Docker and Docker Compose. You can download Docker Desktop from [here](https://www.docker.com/products/docker-desktop).

#### Running Tests

To run tests inside Docker using `docker exec`, execute the following command:

```bash
go test ./services/horizon_test
```

#### Starting Docker

To start Docker with your project:

```bash
docker compose up --build -d
```

#### Starting the Server

To start the server locally:

```bash
go run main.go
```

Including these instructions will help ensure that anyone trying to work with your project has the necessary tools and versions installed to get started quickly.
