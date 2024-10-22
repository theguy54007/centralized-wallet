
# Centralized Wallet API

# Table of Contents

- [Overview](#overview)
- [Installation and Setup](#installation-and-setup)
- [API Workflow](#api-workflow)
- [Project Structure](#project-structure)
  - [Domain Structure](#domain-structure)
  - [Authentication Mechanism](#authentication-mechanism)
  - [Project Boostrap](#project-bootstrap)
  - [API Endpoints](#api-endpoints)
  - [Error Handling](#error-handling)
  - [Logging](#logging)
  - [Middleware](#middleware)
  - [Migration](#migration)
  - [Database Schema Overview](#database-schema-overview)
- [Testing](#testing)
- [Redis Usage](#redis-usage)
- [Interview Related](#interview-related)
  - [Satisfying Functional and Non-Functional Requirements](#satisfying-functional-and-non-functional-requirements)
  - [Explaining Any Decisions You Made](#explaining-any-decisions-you-made)
  - [Features Not Included in the Submission](#features-not-included-in-the-submission)
  - [Areas for Future Improvement](#areas-for-future-improvement)
- [Conclusion](#conclusion)

## Overview

This project implements a centralized wallet system where users can perform the following operations via a RESTful API:

- Deposit money into their wallet
- Withdraw money from their wallet
- Transfer money to another user
- Check their wallet balance
- View their transaction history

The project is built using Go and follows best practices for error handling, logging, and test coverage. The wallet system keeps track of users and wallets, and logs all transactions securely.

## Installation and Setup

  1. Clone the repository:

      ```bash
      git clone https://github.com/theguy54007/centralized-wallet
      cd centralized-wallet
      ```
  2. Setup Go Environment:
      Ensure that you are using Go version 1.22. It’s recommended to use gvm (Go Version Manager) to manage Go versions easily.

  3. Setup environment variables:
      Copy the `.env.example` and rename it as `.env` in the root directory. Then modify the following variables as per your environment setup:

      ```
      PORT=8080
      APP_ENV=local
      DB_HOST=localhost
      DB_PORT=5433
      DB_DATABASE=centralized_wallet
      DB_USERNAME=postgres
      DB_PASSWORD=password1234
      DB_SCHEMA=public
      DB_VOLUME_PATH=localpath
      JWT_SECRET=pleasemakesureitissecret
      REDIS_PORT=6379
      REDIS_ADDRESS=localhost
      REDIS_PASSWORD=password1234
      REDIS_DATABASE=0
      ```

  4. Install dependencies:

      ```bash
      go mod download
      ```

  5. Start Postgres and Redis with Docker

      ```bash
      make docker-up
      ```

  6. After Docker posgtres is ready, run migration

      ```bash
      make db-migrate
      ```

  7. Generate User Seed data

      ```bash
      make db-migrate
      ```

  8. Run the API server:

      ```bash
      make run
      ```

## API Workflow

### 1. User Registration and Login

- Before accessing wallet-related operations, users need to register and login to get a valid JWT token, which is required for authentication on all wallet-related API requests.
- There are two ways to set up the user data:
  1. Manual Registration:
    - Use the POST /register endpoint to create a new user.
    - After registration, use the POST /login endpoint to get the JWT token.
  2. Seed Data:
    - Alternatively, you can generate seed data by running the command:

    ```bash
    make seed
    ```

    - This will create pre-defined user accounts and wallets for testing.

    - If you want to reseed with a clean database, use the command:
    ```bash
    make seed-truncate
    ```

- Important: Login to get the JWT token by using the POST /login endpoint and provide the email and password. The generated token will be needed in the Authorization header of every wallet-related API request.
- For the seed data, here are the default credentials:
  - User 1:
    - Email: jack@example.com
    - Password: password1
  - User 2:
    - Email: david@example.com
    - Password: password2


### 2. JWT Token Authentication

- Once logged in, the server will return a JWT token, which must be included in the Authorization header in the format Bearer <JWT token> for all wallet-related operations (deposit, withdraw, transfer, balance check, etc.).
- For example:

```
Authorization: Bearer <your-jwt-token-here>
```

### 3. Wallet-Related API Endpoints

 After obtaining the JWT token, you can interact with all wallet-related API endpoints by including the token in the Authorization header.

### API Workflow Overview:

1. Register or Seed Data:
  - First, create a new user using the POST /register endpoint or run the make seed command to generate seed data.
2. Login:
  - Use the POST /login endpoint to get the JWT token, which will authenticate all wallet-related requests.
3. Create a wallet:
  - Use the token to manually create a wallet using the /wallets/create API.
  - This is required before making any deposit, withdrawal, or transfer requests.
4. Authenticated Requests:
  - Use the token to authenticate requests to wallet-related endpoints:
  - POST /wallets/deposit: Deposit money into your wallet.
  - POST /wallets/withdraw: Withdraw money from your wallet.
  - POST /wallets/transfer: Transfer money to another user.
  - GET /wallets/balance: Check your wallet balance.
  - GET /wallets/transactions: View your transaction history.
5. Logout:
  - Use the POST /logout endpoint to invalidate the token and log out the user. After logging out, the token will be blacklisted and no longer valid for future requests.


## Project Structure

The project follows a clean, modular architecture with separation of concerns across different layers. Here's an overview of the directory structure:

```bash
├── cmd
│   └── api
│       └── main.go       # Application entry point, sets up the server
│   ├── migration         # Authentication middleware and JWT handling
│       └── main.go
│   ├── seed         # Authentication middleware and JWT handling
│       └── main.go
├── internal
│   ├── auth              # Authentication middleware and JWT handling
│   ├── apperrors         # Custom error handling logic
│   ├── database          # Database connection and setup
│   ├── redis             # Redis connection and operations
│   ├── server            # Server setup and routes registration
│   ├── transaction       # Transaction domain (service, repo)
│   ├── user              # User domain (handler, service, repo)
│   ├── wallet            # Wallet domain (handler, service, repo)
│   ├── utils             # Utility functions, API response helpers,Custom error handling logic and middlewares
│   └── tests             # Unit and integration test utilities, mock services and repositories
├── logs
│    └── app.logs         # append log to log file
└── Makefile              # Commands for seeding data, running tests, and other utilities
```

### Domain Structure

Each domain (e.g., `wallet`, `user`, `transaction`) is modularized into three layers: **handler**, **service**, and **repository**.

#### 1. **Handler Layer**

- **Purpose**: Handles HTTP requests, validates input, and returns responses.
- **Example**: `wallet/handler.go`
- **Flow**:
  - Receives requests, validates the request payload.
  - Calls the service layer to handle the business logic.
  - Returns structured responses via helpers (`utils.SuccessResponse`, `utils.ErrorResponse`).

#### 2. **Service Layer**

- **Purpose**: Contains business logic and coordinates between handlers and repositories. Can also communicate with other services when necessary.
- **Example**: `wallet/service.go`
- **Flow**:
  - Handles business logic (e.g., transferring funds between wallets).
  - Interacts with other services as needed (e.g., `WalletService` interacts with `TransactionService` to record transactions).
  - Returns data to the handler for response.

#### 3. **Repository Layer**

- **Purpose**: Direct database interactions. Isolates data access logic.
- **Example**: `wallet/repository.go`
- **Flow**:
  - Defines methods for interacting with the database.
  - Service layer depends on repositories to handle data persistence and querying.

### Authentication Mechanism

The project uses **JWT-based authentication**:

1. **JWT Token Creation**:
   - Users need to register and log in to receive a JWT token.
   - The token includes the user ID and expires in 72 hours.

2. **JWT Middleware**:
   - All wallet-related routes require a valid JWT token in the `Authorization` header.
   - The middleware validates the token and adds the user ID to the request context.

3. **Token Expiration and Invalidation**:
   - Tokens expire after 72 hours.
   - On logout, the token is blacklisted and stored in Redis. Redis keeps the token invalid until it expires.

4. **Redis Integration**:
   - Redis is used to store blacklisted tokens with expiration times.

### Project Bootstrap

This project was bootstrapped using **go-blueprint**, which provides a well-structured Go project starter template. The template helped set up the architecture, environment configuration, testing, and database handling. More information about go-blueprint can be found [here](https://github.com/Melkeydev/go-blueprint).

---

### API Endpoints

Here is a list of available API endpoints:

- **POST /register**: Register a new user.
  - **Request**: `{ "email": "user@example.com", "password": "password" }`
  - **Response**:
    - Success: `201 Created`

    ```json
    {
      "status": "success",
      "message": "User registered successfully",
      "data": {
        "id": 1,
        "email": "user@example.com"
      }
    }
    ```

    - Error: `400 Bad Request`

    ```json
    {
      "status": "error",
      "message": "Email already in use"
    }
    ```

- **POST /login**: Login a user.
  - **Request**: `{ "email": "user@example.com", "password": "password" }`
  - **Response**:
    - Success: `200 OK`

    ```json
    {
      "status": "success",
      "message": "Login successful",
      "data": {
        "token": "jwt_token_here",
        "user": {
          "id": 1,
          "email": "user@example.com"
        }
      }
    }
    ```

    - Error: `401 Unauthorized`

    ```json
    {
      "status": "error",
      "message": "Invalid email or password"
    }
    ```

- **POST /logout**: Logout the current user.
  - **Headers**: `{ "Authorization": "Bearer jwt_token_here" }`
  - **Response**: `200 OK`

    ```json
    {
      "status": "success",
      "message": "Logged out successfully"
    }
    ```

- **POST /wallets/create**: Create a new wallet for the logged-in user.
  - **Headers**: `{ "Authorization": "Bearer jwt_token_here" }`
  - **Response**:
    - Success: `201 Created`

    ```json
    {
      "status": "success",
      "message": "Wallet created successfully",
      "data": { "wallet_number": "WAL-17-41022114743-YYQYKO" }
    }
    ```

    - Error: `400 Bad Request`

    ```json
    {
      "status": "error",
      "message": "Wallet already exists for this user"
    }
    ```

    - Error: `401 Unauthorized`

    ```json
    {
      "status": "error",
      "message": "Authorization token is required"
    }
    ```

- **POST /wallets/deposit**: Deposit money into the user's wallet.
  - **Headers**: `{ "Authorization": "Bearer jwt_token_here" }`
  - **Request**: `{ "amount": 100 }`
  - **Response**:
    - Success: `200 OK`

    ```json
    {
      "status": "success",
      "message": "Deposit successful",
      "data": {
        "balance": 200,
        "updated_at": "2024-10-22T03:54:13.521945Z"
      }
    }
    ```

    - Error: `401 Unauthorized`

    ```json
    {
      "status": "error",
      "message": "Authorization token is required"
    }
    ```

    - Error: `400 Bad Request`

    ```json
    {
      "status": "error",
      "message": "Invalid request data"
    }
    ```

- **POST /wallets/withdraw**: Withdraw money from the user's wallet.
  - **Headers**: `{ "Authorization": "Bearer jwt_token_here" }`
  - **Request**: `{ "amount": 50 }`
  - **Response**:
    - Success: `200 OK`

    ```json
    {
      "status": "success",
      "message": "Withdrawal successful",
      "data": {
        "balance": 1123752,
        "updated_at": "2024-10-22T03:57:24.434923Z"
      }
    }
    ```
    - Error: `401 Unauthorized`

    ```json
    {
      "status": "error",
      "message": "Authorization token is required"
    }
    ```

    - Error: `400 Bad Request`

    ```json
    {
      "status": "error",
      "message": "Insufficient funds"
    }
    ```

- **POST /wallets/transfer**: Transfer money to another user's wallet.
  - **Headers**: `{ "Authorization": "Bearer jwt_token_here" }`
  - **Request**: `{ "to_wallet_number": "WAL-654321", "amount": 50 }`
  - **Response**:
    - Success: `200 OK`

    ```json
    {
      "status": "success",
      "message": "Transfer successful",
      "data": {
        "balance": 11800,
        "updated_at": "2024-10-22T04:04:06.175189Z"
      }
    }
    ```

    - Error: `401 Unauthorized`

    ```json
    {
      "status": "error",
      "message": "Authorization token is required"
    }
    ```

    - Error: `400 Bad Request`

    ```json
    {
      "status": "error",
      "message": "Insufficient funds"
    }
    ```

    - Error: `400 Bad Request`

    ```json
    {
      "status": "error",
      "message": "Wallet not found"
    }
    ```

    - Error: `400 Bad Request`

    ```json
    {
      "status": "error",
      "message": "Invalid request data"
    }
    ```

- **GET /wallets/balance**: Retrieve the balance of the user's wallet.
  - **Headers**: `{ "Authorization": "Bearer jwt_token_here" }`
  - **Response**:
    - Success: `200 OK`

    ```json
    {
      "status": "success",
      "message": "Balance retrieved successfully",
      "data": {
        "balance": 100,
        "updated_at": "2024-10-22T11:47:43.241007Z",
        "wallet_number": "WAL-17-41022114743-YYQYKO"
      }
    }
    ```

    - Error: `400 Bad Request`

    ```json
    {
      "status": "error",
      "message": "Wallet not found"
    }
    ```

    - Error: `401 Unauthorized`

    ```json
    {
      "status": "error",
      "message": "Authorization token is required"
    }
    ```

- **GET /wallets/transactions**: View the user's transaction history.
  - **Headers**: `{ "Authorization": "Bearer jwt_token_here" }`
  - **Response**:
    - Success: `200 OK`

    ```json
    {
      "status": "success",
      "message": "Transaction history retrieved successfully",
      "data": {
        "transactions": [
          {
            "transaction_type": "withdraw",
            "amount": 110,
            "direction": "outgoing"
          },
          {
            "transaction_type": "deposit",
            "amount": 12,
            "direction": "incoming"
          },
          {
            "transaction_type": "transfer",
            "amount": 200,
            "direction": "outgoing",
            "to_wallet_number": "WAL-7-20241020173819-OX5POR",
            "to_email": "test3@test.com"
          },
          {
            "transaction_type": "transfer",
            "amount": 200,
            "direction": "incoming",
            "from_wallet_number": "WAL-7-20241020173819-OX5POR",
            "from_email": "test3@test.com"
          },
        ],
        "wallet_number": "WAL-5-20241020173643-5NUVLI"
      }
    }
    ```

    - Error: `400 Bad Request`

    ```json
    {
      "status": "error",
      "message": "Invalid order, must be 'asc' or 'desc'"
    }

    ```

    - Error: `400 Bad Request`

    ```json
    {
      "status": "error",
      "message": "Invalid limit, must be between 1 and 100"
    }
    ```

    - Error: `400 Bad Request`

    ```json
    {
      "status": "error",
      "message": "Invalid offset, must be a non-negative integer"
    }
    ```

---

## Error Handling

In this project, I consolidated the error handling and response structure into a single utility package (`utils`). This approach allows for unified handling of both success and error responses, making the code more maintainable and consistent across the application.

### Key Aspects

1. **Centralized Error Definitions:**
   - I have defined and organized common error messages centrally within the `utils` package. This makes it easy to track and maintain error messages across different parts of the application.
   - Error codes and messages are stored centrally to ensure consistency and avoid duplication.

2. **Unified Response Handling:**
   - Both success and error responses use a shared structure. Whether it's a successful response or an error, the structure of the response is always predictable and consistent, which simplifies testing and debugging.
   - Error responses also use this structure, ensuring the message is clear to the client while keeping the codebase clean.

3. **Internal Error Logging:**
   - For internal errors (e.g., database or Redis errors), the error message is logged to a log file. This ensures that sensitive details are not exposed to the client, while still capturing enough information for developers to debug the issue.
   - The logging system captures the error context and details, allowing developers to track down issues without overexposing information to end users.

### Logging

To ensure that developers have visibility into the application's internal workings, I implemented a robust logging system that captures crucial information like errors, HTTP requests, and database interactions.

Key aspects of logging:

- **Error Logging**: Whenever an error occurs (e.g., database errors, request validation failures), the error details are logged using a centralized logging utility. This helps in tracing the root cause of issues quickly.
- **Log Output**: Logs are written to both the console (for real-time monitoring) and a log file for historical analysis. This helps in both development and production environments, where logs can be analyzed for troubleshooting.

### Middleware

The project uses middleware to handle cross-cutting concerns such as authentication, logging, and request validation. Middleware simplifies the architecture and ensures consistency across the codebase.

#### Key Middleware

1. **JWT Authentication Middleware**:
   This middleware is responsible for validating JWT tokens attached to incoming requests for protected routes. The middleware checks the token for validity, expiration, and blacklisting (e.g., in cases where the user logs out). Once the token is validated, the user's `user_id` is stored in the request context, making it accessible for downstream handlers.

2. **Logging Middleware**:
   Instead of using an error middleware, the project implements a **logging middleware** that logs requests and responses with HTTP status codes of `400` or greater. This ensures that all error responses are captured in the logs, aiding in debugging and monitoring without overwhelming the logs with success responses.

   - For example, if a `500 Internal Server Error` or `400 Bad Request` occurs, the error details (such as the error message, request path, and time) will be logged. This logging is done only for responses that require developer attention.
   - The middleware is designed to log errors in a structured format, which can later be extended to log to a file, external logging service, or monitoring system.

3. **WalletNumber Middleware**:
   This middleware is used for caching and fetching wallet numbers to optimize operations that frequently access wallet information. When a user’s transaction history is requested, the wallet number is fetched from Redis if available. If not, it's retrieved from the database and then cached in Redis. This reduces database load and improves performance when querying transaction histories.

### Migration

For database migrations, I use the **Golang Migrate** library, a widely used tool for applying schema changes in a consistent and reliable manner.

#### How Migration Works

- **Migrations Directory**: A `migrations` folder contains all migration scripts, written in SQL, that modify the database schema (e.g., creating tables, adding indexes).
- **Migration Tool**: The migration tool can be invoked using the provided `Makefile` commands:
  - `make db-migrate`: Applies any pending migrations to bring the database schema up to date.
  - `make db-rollback`: Rolls back the last applied migration.
  - `make seed`: Seeds the database with initial test data.
  - `make seed-truncate`: Truncates existing data and re-seeds the database.

## Database Schema Overview

The centralized wallet system contains the following tables, each serving a specific purpose in tracking users, wallets, and transactions. Below is an overview of the database schema, with a detailed description of each table and its columns.

### **Users Table**

- **id**: An auto-incrementing unique identifier for each user.
- **email**: The user’s email address, which must be unique across all users.
- **password**: A securely hashed password for authentication purposes.
- **created_at**: The timestamp when the user was created.
- **updated_at**: The timestamp when the user's information was last updated.

**Description**:
This table holds all user information, including their email and hashed password. Each user has a unique ID, which is referenced in the wallets table.

---

### **Wallets Table**

- **id**: An auto-incrementing unique identifier for each wallet.
- **user_id**: A foreign key that links the wallet to a specific user from the `users` table.
- **wallet_number**: A unique identifier for each wallet, often used in transactions.
- **balance**: The current balance in the wallet, stored as a numeric value to two decimal places.
- **created_at**: The timestamp when the wallet was created.
- **updated_at**: The timestamp when the wallet's balance or details were last updated.

**Description**:
This table keeps track of all wallets in the system. Each user can have only one wallet. The wallet number is unique and is used for identifying wallets during transactions. The balance is updated when users perform deposit, withdraw, or transfer operations.

---

### **Transactions Table**

- **id**: An auto-incrementing unique identifier for each transaction.
- **from_wallet_number**: The wallet number from which the money is transferred or withdrawn. This can be null in case of a deposit.
- **to_wallet_number**: The wallet number to which money is transferred or deposited. This can be null in case of a withdrawal.
- **transaction_type**: The type of transaction, which can be either `deposit`, `withdraw`, or `transfer`.
- **amount**: The amount of money involved in the transaction, stored as a numeric value to two decimal places.
- **created_at**: The timestamp when the transaction was completed.

**Description**:
This table records all transactions within the wallet system. It supports three types of transactions:

1. **Deposit**: Money is added to a wallet.
2. **Withdraw**: Money is taken from a wallet.
3. **Transfer**: Money is moved from one wallet to another.

Each transaction has its own unique identifier and stores relevant details such as the amount, type, and involved wallets.

---

## Database Relationships

- Each user has a **one-to-one relationship** with a wallet, meaning every user can only have one wallet associated with them.
- The **Transactions table** uses the wallet numbers from the `wallets` table to track both the sender (from_wallet_number) and the receiver (to_wallet_number) for `transfer` transactions, and it uses either of the wallet numbers for `deposit` and `withdraw` operations.

By structuring the database in this way, the system ensures that all financial transactions are logged and tracked accurately. The relationships between users, wallets, and transactions are maintained through foreign keys, providing a robust framework for managing centralized wallets.


## Testing

### Running Tests

To run the tests:

```bash
make test
```

### Descriptions

Due to time constraints, the overall test coverage of the project is not as extensive as desired. However, I have ensured that the core features of the application are well-tested. The following sections explain the testing approach:

#### Unit Tests

Unit tests have been written to cover the essential components of the application. The focus is on testing core logic and edge cases using mock implementations. The following features have been covered in the unit tests:

- **Wallet Service & Handlers**: These tests ensure that the wallet operations (deposit, withdraw, transfer, balance checking, transaction history) are functioning correctly and handle edge cases.
- **Transaction Service**: Tests cover the transaction recording and history retrieval operations.
- **User Handlers & Service**: These tests validate the user registration, login, and logout processes, including edge cases like invalid inputs and failed authentication.
- **JWT Middleware**: Tests validate the JWT authentication process, checking for invalid tokens, expired tokens, and blacklisted tokens.
- **Wallet Middleware**: Tests cover wallet retrieval from Redis and the database, ensuring correct behavior in both cache hits and misses.

Unit tests mainly use mock objects to isolate and test individual components without external dependencies like databases or Redis.

#### Integration Tests

Integration tests have been set up to test the interaction between services and the underlying database (PostgreSQL) and Redis cache. These tests utilize `testcontainer-go` to spin up PostgreSQL and Redis containers for testing purposes.

The primary focus for integration tests is on:

- **Wallet Service**: Testing wallet operations in a real environment where data is persisted in PostgreSQL, ensuring that wallet balance updates and transaction records are consistent.
- **Transaction Service**: Validating that transaction records are correctly created, and the transaction history is retrieved accurately, including edge cases when interacting with the database.

Integration tests are vital for verifying that the system works correctly when integrating different layers (service, repository, database, Redis) and handling real-world edge cases that might not surface in unit testing.

### Test Summary

- **Unit Tests**: Focus on core logic and edge cases with mock implementations.
- **Integration Tests**: Focus on testing the integration with PostgreSQL and Redis via `testcontainer-go`, primarily covering wallet and transaction services.

While the current test coverage is not exhaustive, it covers the most critical components of the application, ensuring robustness in the core features.

## Redis Usage

The project uses Redis for several purposes, mainly to enhance performance and manage token invalidation. Here are the key areas where Redis is utilized:

1. **Blacklist Service for Authentication**:
   Redis is used to store blacklisted JWT tokens that have been invalidated upon user logout. This ensures that even if the token has not yet expired, it will be recognized as invalid if it’s been blacklisted. Redis stores the blacklisted token until it naturally expires, ensuring no long-term storage of these invalid tokens.

2. **Wallet Middleware**:
   Redis is leveraged to store or fetch the wallet number of a user. This is primarily used to boost performance when users need to check their transaction history. Rather than querying the database for the wallet number each time, the middleware first checks if the wallet number is cached in Redis. If found, it is fetched from the cache; otherwise, the database is queried, and the result is stored in Redis for future requests. This reduces the load on the database for frequent transaction-related queries.

3. **Transaction History Caching**:
   Transaction history queries, especially those that require joining multiple tables, can be resource-intensive. Redis is used to cache the results of these queries by generating a unique key based on the user ID, wallet number, page number, and order. This allows subsequent requests for the same data to be served quickly from Redis, reducing the load on the database.

   - **Cache Invalidation**: When operations that modify transaction history (such as deposit, transfer, or withdrawal) are performed, the cache is invalidated (removed) to ensure that the data remains accurate. This ensures that users always receive the latest transaction data after these operations.

### Redis and Performance

By using Redis as a caching layer for frequent or resource-intensive operations (like fetching wallet numbers or transaction histories), the application minimizes database access, improving both response times and the overall system's scalability.

Redis’ fast in-memory storage and eviction policies make it a great fit for tasks requiring quick access to frequently requested data, with mechanisms in place to ensure that stale or outdated data is promptly invalidated when necessary.


## Interview Related

### Satisfying Functional and Non-Functional Requirements

#### Functional Requirements

1. **User Wallet Operations**:
    - **Deposit**: Users can easily deposit money into their wallet through a well-structured and secure API.
    - **Withdraw**: Users can withdraw funds from their wallet as long as they have sufficient balance. Error handling is in place to prevent overdrafts.
    - **Transfer**: The API allows users to transfer funds to another user's wallet, ensuring that both sender and recipient exist and have valid wallets.
    - **Balance Retrieval**: Users can check their wallet balance through a dedicated API that fetches up-to-date information.
    - **Transaction History**: The API provides users with a paginated view of their transaction history, allowing them to track all deposits, withdrawals, and transfers.

2. **User Authentication**:
    - Users must register and log in to access wallet-related operations.
    - JWT-based authentication ensures that users have secure access to the system, with tokens invalidated on logout to prevent misuse.

3. **Wallet Creation**:
    - Users can create their wallet through a simple API call after registration.
    - Seed data for user accounts is also available for testing and demonstration purposes.

#### Non-Functional Requirements

1. **Performance**:
    - **Caching with Redis**: Wallet numbers and transaction history are cached in Redis to improve performance by reducing load on the database for frequently accessed data. This speeds up response times for users checking their transaction history or balance.
    - **Database Optimization**: The PostgreSQL database schema has been designed with normalization and appropriate indexes to ensure efficient querying of user and transaction data.
    - **Graceful Shutdown**: The system is equipped with a graceful shutdown mechanism to ensure that all in-progress requests are properly handled before the server shuts down, enhancing system reliability.

2. **Scalability**:
    - The project follows the separation of concerns pattern with well-defined layers (handler, service, repository), allowing easy scaling by adding new features or improving existing ones without impacting other parts of the system.
    - Using Redis for caching ensures that the system can handle a growing user base without heavily relying on the database for every transaction query.

3. **Security**:
    - **JWT Authentication**: The use of JWT for authentication, with a token expiration time of 72 hours, ensures that users are securely authenticated while preventing stale tokens from being used.
    - **Token Blacklisting**: Logged-out tokens are stored in Redis and blacklisted to ensure they cannot be used again, even if they have not yet expired. This helps prevent unauthorized access after logout.
    - **Error Handling and Validation**: All API endpoints have thorough input validation and error handling to prevent bad or malicious data from entering the system. Any unauthorized access or misuse is logged for auditing and debugging.

4. **Error Handling and Logging**:
    - Custom error messages are defined and logged at various points in the application. This ensures that developers can track issues and understand where they originate.
    - Errors in the service or repository layers are handled gracefully, and detailed logs are written to assist developers in debugging and maintenance.

5. **Test Coverage**:
    - Unit tests cover critical areas like wallet operations, transaction handling, and user authentication, ensuring that core features are reliable.
    - Integration tests, using Testcontainers, validate the system’s interaction with real instances of PostgreSQL and Redis to ensure that the code works as expected in production-like environments.
    - While time constraints limited total coverage, the most important and complex operations have been covered to ensure robustness.

6. **Simplicity and Maintainability**:
    - The project follows a clear and simple architecture that separates business logic, data handling, and presentation layers, making it easier to maintain and extend.
    - **Dependency Injection** is used in services and repositories, allowing easier testing and decoupling between layers.
    - The **README** provides clear instructions for setting up, running, and testing the application, ensuring that developers and reviewers can quickly understand and work with the code.

### Explaining Any Decisions You Made

1. **Choice of JWT for Authentication**:
    - I chose **JSON Web Tokens (JWT)** for user authentication due to its stateless nature and scalability. JWT is widely adopted for securing APIs and allows users to authenticate without requiring server-side session storage, which is ideal for scalability.
    - JWTs are set to expire after 72 hours to ensure security. This limits the lifespan of a potentially compromised token. A token blacklist mechanism is implemented using Redis, allowing tokens to be invalidated upon logout.
    - While JWT refresh tokens are a common practice for renewing tokens without forcing re-logins, they were not implemented in this version due to time constraints. This is a potential area for future improvement.

2. **Redis for Performance and Security**:
    - Redis was used for two primary purposes: caching and token blacklisting. Caching helps reduce the database load and enhances performance by storing frequently queried data like wallet numbers and transaction history. This is especially useful when dealing with expensive queries involving database joins.
    - Redis also serves as the backend for blacklisting JWT tokens after logout. This ensures that tokens are invalidated before they expire and can no longer be used for authentication. Redis’s fast access and TTL (time-to-live) support make it ideal for this use case.

3. **Repository-Service Pattern for Clean Architecture**:
    - The project follows the **repository-service pattern**, ensuring a clear separation between business logic (service layer) and data access logic (repository layer).
    - The repository layer interacts directly with the database, handling queries and transactions. The service layer encapsulates business logic and coordinates interactions between repositories and other services, like the transaction service for recording transactions after wallet operations.
    - This approach improves testability, maintainability, and decoupling of responsibilities. Dependency injection is used throughout the project to inject dependencies into services and handlers.

4. **Middleware for Error Handling and Logging**:
    - A centralized error-handling middleware was implemented to standardize the way errors are managed throughout the application. This middleware catches any unhandled errors and responds with a consistent error structure, ensuring uniform error messages.
    - The middleware also logs critical internal errors for further debugging, providing valuable information for diagnosing issues. This approach ensures that errors are managed in a clean, concise manner without repeating error handling code in every handler.

5. **Transaction History Design**:
    - **DB Transaction for Consistency**: To ensure transaction records and wallet balances remain consistent, I employed **database transactions** during deposit, withdrawal, and transfer operations. This ensures that either all operations succeed (wallet updates and transaction records) or all operations are reverted in case of failure. This design guarantees the atomicity of operations, making it easier to maintain the consistency of both wallet records and transaction history.
    - **From/To Wallet Number Design**: I chose to use `from_wallet_number` and `to_wallet_number` fields instead of a single wallet number or `user_id`. This design was selected for several reasons:
        1. **Clarity in Transactions**: Using both `from_wallet_number` and `to_wallet_number` makes it explicitly clear where funds are coming from and where they are going. This is especially useful in transfer operations between users.
        2. **Flexibility**: By using wallet numbers instead of user IDs, we allow for greater flexibility in the future. For example, this structure could easily support additional features like sub-accounts or multi-wallet users without needing to alter the transaction schema.
        3. **Security and Abstraction**: By using wallet numbers instead of user IDs, the system abstracts away direct user identification in transactions, potentially increasing security and allowing wallet numbers to serve as the primary reference for financial operations, which is a typical pattern in many financial systems.

6. **Wallet Number Generation**:
    - In this system, every user is assigned a **unique wallet number** upon wallet creation. The wallet number functions similarly to a **bank account number**, uniquely identifying each wallet in the system.
    - For the purpose of this project, a **simple algorithm** was used to generate wallet numbers by combining the user ID, a timestamp, and a random string. This was done to ensure that each wallet number is unique without introducing complex logic.

7. **Simple Token-Based Authentication**:
    - Given the time constraints, I opted for simplicity by implementing token-based authentication without refresh tokens. Users must log in again after their JWT expires in 72 hours. This reduces the complexity of token lifecycle management.
    - The implementation is future-proofed by incorporating a Redis-based blacklist system for logout, allowing tokens to be invalidated before their expiration. Introducing refresh tokens in future iterations would improve the user experience by avoiding frequent logins.

8. **Testing Strategy**:
    - I focused on writing unit tests for core functionalities (wallet services, transaction services, user handlers, etc.), ensuring that business logic and edge cases are covered with mock objects.
    - Integration tests were written using `testcontainers-go` to ensure the seamless integration between services and databases (Redis, PostgreSQL). This verifies the system's behavior in a real-world environment with external dependencies.
    - While full coverage was not achieved due to time constraints, I prioritized testing the most critical paths and edge cases in the application, ensuring that key areas are well-tested.

9. **Security Considerations**:
    - In addition to JWT authentication, user passwords are hashed and stored securely to prevent leaks in case of a data breach.
    - Sensitive operations (e.g., wallet transfers, balance checks) are protected by the JWT authentication mechanism, ensuring that only authorized users can perform actions on their own wallets. The use of Redis for token blacklisting ensures that compromised tokens can be revoked immediately upon logout.

### Features Not Included in the Submission

1. **Refresh Tokens for Authentication**:
    - I chose to implement **JWT-based authentication** without the use of **refresh tokens**. Typically, refresh tokens are used to prolong user sessions without requiring frequent re-authentication, but due to time constraints, I focused on implementing basic token expiration and token invalidation using Redis.
    - In a real-world scenario, adding refresh tokens would improve the user experience by allowing users to automatically refresh their tokens without requiring login after expiration.

2. **Advanced Wallet Number Generation**:
    - The current implementation of **wallet number generation** uses a basic combination of user ID, timestamp, and random string. While this is sufficient for the project’s scope, a more robust wallet number generation mechanism would be needed in production. For example, using a globally unique identifier (GUID) or a centralized sequence generator to ensure uniqueness and handle collision detection.

3. **Comprehensive Data Validation and Input Sanitization**:
    - Although some basic data validation is performed (e.g., checking for valid email format and minimum password length), the project does not include **comprehensive input validation** or **input sanitization** for all possible cases. Adding more robust validation would ensure that the system is more resilient against invalid or malicious inputs.

4. **End-to-End Testing:**
   I did not include full end-to-end tests such as route testing. The main reason for this omission is the time constraint. However, I have covered unit tests for handlers and integration tests for services, which sufficiently test the core functionality.

5. **Test Code Structure:**
   Not every test case is written using clean code principles or follows the best structure. This can be improved in future iterations, but given the time limit, I focused on covering the main scenarios rather than code refinement in tests.

6. **Simplified Database Design:**
   I chose to keep the database schema simple, focusing only on the necessary tables like `transactions` and `wallets`. More complex relational designs could be added in the future, but for now, this design satisfies the core requirements.

7. **Transaction History Design (Code Location):**
   - The transaction history logic is currently located in `wallet_handlers.go`. Ideally, this code should be separated into its own transaction handler and service. However, doing so would have introduced circular dependencies between services. While this could have been resolved by introducing a shared service layer, it would have complicated the code. Due to time constraints, the decision was made to leave it in `wallet_handlers.go` for now.

8. **Authentication Security Check:**
   - The current implementation of JWT authentication middleware fetches the `user_id` from the JWT but does not verify whether the user actually exists in the system. This could pose a potential security risk. A more secure and optimal approach would involve adding middleware or extending the JWT to check for the existence of the user in the database, ensuring the user is valid before allowing access to protected routes.


### Areas for Future Improvement

1. **Code Maintainability**:
    - **Database Code and Repository Layer**: The current repository layer requires writing raw SQL queries, which can lead to duplication and verbosity, making the code harder to maintain. Introducing an ORM (like GORM in Go) or a query builder library would simplify database interactions, reduce the likelihood of repetitive code, and make the codebase more manageable. An ORM would also allow easier migrations, relationship handling, and querying without needing to write complex SQL queries manually for each operation.
    - **Service and Repository Code Organization**: While the service layer decouples business logic from repository code, there is room for improvement in how the code is organized and maintained. Refactoring parts of the service logic that involve complex interactions with multiple repositories would improve clarity and modularity.

2. **Event Sourcing / Event-Driven Architecture for Transaction Handling**:
    - The current approach uses **database transactions** to ensure consistency when recording transactions (deposit, withdrawal, transfer). While this ensures strong consistency, it can introduce some latency due to the transactional nature of database operations.
    - An alternative approach would be to implement **event sourcing** or an **event-driven architecture**, where an event is generated and published whenever a transaction operation is initiated. Subscribers can then listen to these events and perform the necessary actions, such as recording the transaction, updating the wallet balance, etc.
    - This approach could improve the system’s scalability, reduce latency, and ensure eventual consistency. However, it introduces additional complexity and would require careful handling to ensure message delivery guarantees (e.g., using Kafka or a similar messaging system).

3. **Testing Coverage and Depth**:
    - **Expanded Test Coverage**: While critical parts of the project (wallet, transaction, authentication) have good test coverage, additional unit and integration tests would be beneficial to cover more edge cases and improve overall robustness. This includes testing error scenarios more thoroughly, particularly with complex transaction workflows and multi-service interactions.
    - **End-to-End (E2E) Testing**: Introducing E2E testing for API endpoints, simulating real-world scenarios, would enhance confidence that the system works as expected across the full stack (from request to database).

4. **Performance Optimization**:
    - **Database Query Optimization**: Some database queries, especially those involving joins (e.g., retrieving transaction history), may become less efficient as the data grows. Optimizing SQL queries and adding appropriate indexes will be crucial for performance at scale. Additionally, implementing query caching strategies or leveraging database-specific features (like PostgreSQL materialized views) could further enhance performance for frequently accessed data.
    - **Rate Limiting**: Implementing rate limiting to prevent abuse of the API endpoints would help secure the system from DDoS attacks or excessive load from malicious users.

5. **Security Enhancements**:
    - **Implementing Refresh Tokens**: Currently, the system only uses access tokens, which expire after 72 hours. Introducing refresh tokens would improve the security and user experience by allowing users to stay logged in for longer periods without re-entering their credentials.
    - **Two-Factor Authentication (2FA)**: For higher security, especially in handling monetary transactions, integrating 2FA into the authentication process would add an additional layer of security for user accounts.

6. **Code Cleanliness**:
    - **Refactoring for Simplicity**: While the code structure is functional, there are areas where the code can be cleaned up for better readability and simplicity. Reducing the number of nested conditions, improving variable naming, and breaking down larger functions into smaller, reusable components would improve the overall maintainability of the codebase.
    - **DRY Principle (Don't Repeat Yourself)**: Some code across the services, especially related to database queries, could benefit from refactoring to follow the DRY principle. Common logic can be extracted into helper functions or shared service methods to avoid duplication and improve maintainability.
    - **Configuration Management**: The current configuration (e.g., database, Redis, JWT secret) is handled through environment variables, but this could be enhanced using a configuration management library that supports different environments (development, production) with more structured validation.

7. **Future Feature Expansions**:
    - **Scalability Considerations**: As the user base grows, further architectural improvements, such as breaking out services into microservices, would be necessary. This would allow independent scaling of components like wallet management, transaction handling, and authentication.
    - **Audit Logs**: Adding audit logs for key events like deposits, withdrawals, and transfers would be beneficial for both security and troubleshooting purposes.
    - **Improved Error Logging**: While error logging is currently implemented, improving the granularity of logs (e.g., adding user context and request metadata) would make it easier to debug issues in production.
8. **Reusing Wallet Middleware for Other Operations**:
    - Currently, the **wallet middleware** caches and fetches the wallet number to boost performance when checking transaction history. However, other wallet-related operations (such as deposit, withdraw, and transfer) can also leverage this middleware to reduce database queries and improve code reusability.
    - By caching wallet information more widely, overall system performance could be enhanced, especially for frequent operations. This approach would also make the code more DRY (Don’t Repeat Yourself) by centralizing the retrieval and validation of wallet numbers in the middleware.
9. **Advanced Wallet Number Generation**:
    - The current wallet number generation mechanism is basic and may not be suitable for large-scale production environments. It uses a combination of the user ID, timestamp, and random string, but this can be replaced with a more robust solution such as using **GUIDs** or implementing a **centralized sequence generator** for globally unique and collision-free wallet numbers.



## Conclusion

1. **Enjoyment of the Process**:
   - I thoroughly enjoyed the coding process and building this project from scratch in Golang. Although it was my first time constructing a full project in Go on my own, the experience has been incredibly rewarding. I learned a great deal about Go’s ecosystem, best practices, and how to design and structure a scalable project.

2. **Time Spent**:
   - A significant portion of the time was spent on research and studying best practices, especially around Golang conventions, testing strategies, and error handling. Writing the test cases took considerable time, but it also helped me uncover optimization opportunities, which led to refactoring and enhancing the code. I had to revisit and improve parts of the project after gaining insights from writing tests, making the code more maintainable and robust.

3. **Balancing Simplicity and Robustness**:
   - Throughout the development, I aimed to keep the codebase simple yet robust. However, as with any project, there is always the concern of over-design or overthinking certain aspects. I consciously tried to avoid over-complicating things and focused on delivering a well-designed, efficient solution. I am open to feedback and excited about the possibility of discussing improvements further.

4. **Commitment to Quality**:
   - Even with the time constraints, I made a strong effort to ensure the project adhered to best practices, including modular design, clean architecture, and comprehensive testing. While there are areas for future improvement, I believe the project showcases my ability to learn, adapt, and apply best practices to deliver a reliable solution.

5. **Looking Forward**:
   - I hope this submission demonstrates not only my technical skills but also my commitment to producing high-quality, maintainable code. I look forward to discussing this project further, receiving feedback, and potentially exploring more complex features or refinements.

Thank you for the opportunity, and I hope you find this project valuable!
