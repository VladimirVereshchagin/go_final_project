
# Contributing to `scheduler`

Thank you for your interest in the **scheduler** project! We’re excited to see your ideas, improvements, and bug fixes.

## General Guidelines

Please review these guidelines before you start contributing to keep the process simple and efficient.

1. **Fork**: First, create a fork of the project in your GitHub account.
2. **Clone**: Clone your fork of the project to your local machine.

    ```bash
    git clone https://github.com/VladimirVereshchagin/scheduler.git
    cd scheduler
    ```

3. **Create a New Branch**: Always create a new branch for your changes. Use descriptive names, for example, `feature/add-authentication` or `fix/login-bug`.

    ```bash
    git checkout -b feature/your-feature-name
    ```

4. **Development**: Make your changes and add commits with clear messages.

    ```bash
    git add .
    git commit -m "Add feature for user authentication"
    ```

5. **Code Style**: Make sure your code adheres to the project’s style. Use pre-commit hooks configured in the project.

## Pull Requests (PR)

When your changes are ready, create a Pull Request following these guidelines:

1. **Describe Changes**: In your Pull Request, explain what was changed and why.
2. **Issue Links**: If your PR relates to an existing Issue in the project, link to the Issue using `#issue_number`.
3. **Testing**: Ensure your code passes all tests locally.

## Pre-Pull Request Checklist

- Ensure all changes are in the new branch.
- Check that tests are passing:

    ```bash
    ./run-tests.sh
    ```

- If new functionality was added, write tests for it in the `tests/` folder.
- Review your code and ensure it follows the project’s style.

## Writing Tests

The project uses a separate test database to avoid conflicts with the main database. Basic testing guidelines:

1. **Write Tests for New Features**: Ensure that each new feature has corresponding tests in the `tests/` folder.
2. **Test Coverage**: Aim for good test coverage, including both positive and negative cases.
3. **Run Tests Before Commit**: Ensure that tests pass successfully before committing.

## Local Build

To build and test the project on your local machine:

1. **Build the Application**:

    ```bash
    go build -o app ./cmd
    ```

2. **Run the Docker Container**: If you need to test the build in Docker, use the following command:

    ```bash
    docker run -d -p 7540:7540 --name scheduler --env-file .env vladimirvereschagin/scheduler:latest
    ```

## Code Style and Pre-commit Hooks

The project uses pre-commit hooks to automate code checks. Set them up before starting:

```bash
pre-commit install
pre-commit run --all-files --verbose
```

## Questions and Feedback

If you have questions or suggestions, create an [Issue](https://github.com/VladimirVereshchagin/scheduler/issues) on GitHub or participate in the [Discussions](https://github.com/VladimirVereshchagin/scheduler/discussions) section. We’re always open to your ideas and suggestions for improving the project.

Thank you for contributing to the project! 🙌 We appreciate each contributor and strive to make `scheduler` better together!
