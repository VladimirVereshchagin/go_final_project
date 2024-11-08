
# scheduler

![Build Status](https://github.com/VladimirVereshchagin/scheduler/workflows/Go%20CI/CD/badge.svg)
![Go Version](https://img.shields.io/badge/go-1.22-blue)
![Docker Image Size](https://img.shields.io/docker/image-size/vladimirvereschagin/scheduler/latest)

[Docker Hub repository for scheduler](https://hub.docker.com/r/vladimirvereschagin/scheduler)

## Project Description

**scheduler** is a web application for task scheduling developed in Go. The application allows users to create, view, edit, and delete tasks, as well as mark them as completed. It uses **SQLite** as the database with a pure Go driver [`modernc.org/sqlite`](https://gitlab.com/cznic/sqlite), which simplifies building and deploying the application on different architectures. The application provides a RESTful API and includes a frontend for convenient interaction.

### All Starred Tasks Implemented, Including

- **Authentication:** Implemented authentication mechanism using JWT tokens. Access to the application is protected by a password set through the `TODO_PASSWORD` environment variable.
- **Docker Image Creation:** A `Dockerfile` has been developed to build the Docker image of the application, simplifying its deployment and scaling. The ready-made image is available on [Docker Hub](https://hub.docker.com/r/vladimirvereschagin/scheduler).
- **Cross-compilation and Multi-architecture Support:** Thanks to the pure Go driver for SQLite, the application supports cross-compilation and building multi-architecture Docker images, allowing it to run on various platforms, including `linux/amd64` and `linux/arm64`.

## Requirements

- **Go** version **1.22** or higher
- **Git**
- **Docker** (for running in a container)

## Installation and Running

### Clone the Repository

```bash
git clone https://github.com/VladimirVereshchagin/scheduler.git
cd scheduler
```

### Set Environment Variables

Create a `.env` file at the root of the project with the following content:

```bash
TODO_PORT=7540
TODO_DBFILE=data/scheduler.db
TODO_PASSWORD=your_password_here
```

- `TODO_PORT` — Port to run the web server (default is 7540).
- `TODO_DBFILE` — SQLite database file name.
- `TODO_PASSWORD` — Password for accessing the application. Leave empty if authentication is not required.

### Install Dependencies

```bash
go mod download
```

### Database Initialization

No explicit database initialization is required. The application will automatically create the database in the `data` directory upon first launch.

### Build the Application

```bash
go build -o app ./cmd
```

### Run the Application

```bash
./app
```

### Access the Application

Open your browser and go to:

```bash
http://localhost:7540/
```

If a password is set (the `TODO_PASSWORD` variable is not empty), you will be redirected to the login page:

```bash
http://localhost:7540/login.html
```

Enter the configured password to access the application.

## Quick Start via Docker Hub

For quick deployment of the application, you can use the ready-made Docker image available on [Docker Hub](https://hub.docker.com/r/vladimirvereschagin/scheduler). The image supports `linux/amd64` and `linux/arm64` architectures, making it compatible with various platforms.

### Run the Container

## Important

Before running the container, make sure you set the correct password in the `TODO_PASSWORD` environment variable, or leave it empty if you want to run the application without a password. These values should match those specified in the `.env` file. This is necessary for proper authentication in the application. The `data` directory, which already exists in the project, is used to store the database. Use the following command to run the container:

```bash
docker run -d \
  -p 7540:7540 \
  --name scheduler \
  --env TODO_PORT=7540 \
  --env TODO_DBFILE=data/scheduler.db \
  --env TODO_PASSWORD=your_password_here \
  -v $(pwd)/data:/app/data \
  vladimirvereschagin/scheduler:latest
```

### Explanation

- `-p 7540:7540` — Maps port 7540 to access the application at `http://localhost:7540/`.
- `--env TODO_PORT=7540` — Specifies the application’s port.
- `--env TODO_DBFILE=data/scheduler.db` — Connects the database file.
- `--env TODO_PASSWORD=your_password_here` — Sets the password to access the application (leave empty to run without a password).
- `-v $(pwd)/data:/app/data` — Mounts the database on the host to persist data outside the container.

### Access via Browser

After starting the container, the application will be available at:

```bash
http://localhost:7540/
```

### Stop and Remove the Container

To stop and remove the container, execute the following commands:

```bash
docker stop scheduler
docker rm scheduler
```

## Running Tests

### Before Running Tests

Ensure the application is not running or is using a different database to avoid conflicts.

### Run Tests via Script

The tests use a separate test database `test_data/test_scheduler.db` to avoid conflicts with the main application database.
Use the `run-tests.sh` script to automatically run the tests. The script automatically handles cases with and without a set password.

```bash
./run-tests.sh
```

## How the `run-tests.sh` Script Works

- Starts the application in the background with the specified `TODO_PASSWORD`.
- Sets the `TODO_DBFILE` environment variable to `$(pwd)/test_data/test_scheduler.db`.
- Creates the `test_data` directory if it does not exist.
- Starts the application in the background, using the test database.
- Retrieves a JWT token for authentication (if a password is set) and sets the `TOKEN` environment variable.
- Runs tests using the configured environment variables.
- Stops the application after completing the tests.
- Deletes the test database `test_data/test_scheduler.db`.
- Removes the `test_data` directory if it is empty.

### Test Settings

In the `tests/settings.go` file, you can configure the following parameters:

- `Port`: Port on which the application runs (default is 7540).
- `DBFile`: Path to the database file for testing.
- `Token`: JWT token for authentication, typically set automatically by the `run-tests.sh` script.

## Additional Information

### CI/CD

GitHub Actions is set up in the project for automatic building and testing on pushes to the `main` and `new-feature` branches.
Upon successful build, a multi-architecture Docker image is created and pushed to Docker Hub.

### Pre-commit Hooks

`pre-commit` is used for automatic code checking before committing. Install the hooks with the following command:

```bash
pre-commit install
```

To manually check all code, run:

```bash
pre-commit run --all-files --verbose
```

## Project Structure

- `cmd/` — Entry point of the application (`main.go`).
- `internal/` — Internal packages of the application:
  - `app/` — Setup of routes and handlers.
  - `auth/` — Authentication and JWT handling.
  - `config/` — Configuration loading and management.
  - `models/` — Data models.
  - `repository/` — Database interactions.
  - `services/` — Business logic of the application.
  - `timeutils/` — Date and time utility functions.
- `tests/` — Unit and integration tests.
- `web/` — Frontend files (HTML, CSS, JavaScript).

## Feedback

If you have any questions or suggestions, please create an [issue](https://github.com/VladimirVereshchagin/scheduler/issues) or [pull request](https://github.com/VladimirVereshchagin/scheduler/pulls) in the project repository.

## License

This project is distributed under the MIT License. See the [LICENSE](LICENSE) file for details.
