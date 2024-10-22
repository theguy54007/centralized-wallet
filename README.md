
# Centralized Wallet API

# Table of Contents

- [Overview](#overview)
- [Installation and Setup](#installation-and-setup)
- [API Workflow](#api-workflow)
- [Project Structure](#project-structure)
  - [Project Boostrap](#project-bootstrap)
  - [Structure](#structure)
  - [Domain Structure](#domain-structure)
  - [Authentication Mechanism](#authentication-mechanism)
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

The project is built using **Go (version 1.22)** and leverages the **Gin web framework** for its simplicity, performance, and ease of creating RESTful APIs. Gin provides built-in middleware support for routing, error handling, and JSON handling, making it a great fit for building efficient and scalable APIs.

This project follows best practices for error handling, logging, and test coverage. The wallet system keeps track of users and wallets, logs all transactions securely, and uses Redis for caching and token management to enhance performance and security.

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

      ```env
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

  5. Start Postgres and Redis with Docker (make sure your local has installed Docker)

      ```bash
      make docker-run
      ```

  6. After Docker posgtres is ready, run migration

      ```bash
      make db-migrate
      ```

  7. Generate User Seed data

      ```bash
      make seed
      ```

  8. Run the API server:

      ```bash
      make run
      ```

## API Workflow

### 1. User Registration and Login

- Before accessing wallet-related operations, users need to register and login to get a valid JWT token, which is required for authentication on all wallet-related API requests.
- There are two ways to set up the user data:
  1. **Manual Registration**:
  - Use the `POST /register` endpoint to create a new user.
  - After registration, use the `POST /login` endpoint to get the JWT token.
  2. **Seed Data**:
  - Alternatively, you can generate seed data by running the command:

    ```bash
    make seed
    ```

  - This will create pre-defined user accounts and wallets for testing.

  - If you want to reseed with a clean database, use the command:

    ```bash
    make seed-truncate
    ```

- **Important**: Login to get the JWT token by using the `POST /login` endpoint and provide the email and password. The generated token will be needed in the `Authorization` header of every wallet-related API request.
- For the seed data, here are the default credentials:
  - **User 1**:
    - Email: <jack@example.com>
    - Password: password1
  - **User 2**:
    - Email: <david@example.com>
    - Password: password2

### 2. JWT Token Authentication

- Once logged in, the server will return a JWT token, which must be included in the `Authorization` header in the format `Bearer <JWT token>` for all wallet-related operations (deposit, withdraw, transfer, balance check, etc.).
- For example:

```
Authorization: Bearer <your-jwt-token-here>
```

### 3. Create Wallet

- After logging in and obtaining the JWT token, users must create a wallet before performing any wallet-related actions.
- Use the `POST /wallets/create` endpoint to create a new wallet for the authenticated user.
- The wallet creation request must include the JWT token in the `Authorization` header, and the system will generate a unique wallet number for the user.

### 4. Wallet-Related API Endpoints

 After obtaining the JWT token, you can interact with all wallet-related API endpoints by including the token in the Authorization header.

### 5. Logout

  Use the POST /logout endpoint to invalidate the token and log out the user. After logging out, the token will be blacklisted and no longer valid for future requests.

### API Workflow Overview

1. **Register or Seed Data**:
    - First, create a new user using the `POST /register` endpoint, or run the `make seed` command to generate seed data.

2. **Login**:
    - Use the `POST /login` endpoint to get the JWT token, which will authenticate all wallet-related requests.

3. **Create a Wallet**:
    - Use the token to manually create a wallet using the `/wallets/create` API.
    - This is required before making any deposit, withdrawal, or transfer requests.

4. **Authenticated Requests**:
    - Use the token to authenticate requests to wallet-related endpoints:
      - `POST /wallets/deposit`: Deposit money into your wallet.
      - `POST /wallets/withdraw`: Withdraw money from your wallet.
      - `POST /wallets/transfer`: Transfer money to another user.
      - `GET /wallets/balance`: Check your wallet balance.
      - `GET /wallets/transactions`: View your transaction history.

5. **Logout**:
    - Use the `POST /logout` endpoint to invalidate the token and log out the user. After logging out, the token will be blacklisted and no longer valid for future requests.


## Project Structure

### Project Bootstrap

This project was bootstrapped using **go-blueprint**, which provides a well-structured Go project starter template. The template helped set up the architecture, environment configuration, testing, and database handling. More information about go-blueprint can be found [here](https://github.com/Melkeydev/go-blueprint).

### Structure

The project follows a clean, modular architecture with separation of concerns across different layers. Here's an overview of the directory structure:

```bash
├── cmd
│   └── api
│   │   └── main.go       # Application entry point, sets up the server
│   ├── migration
│   │   └── main.go       # Commandline for migrating DB
│   └── seed
│       └── main.go       # Command line for generating user data
├── internal
│   ├── auth              # Authentication middleware and JWT handling
│   ├── database          # Database connection and setup
│   ├── logging           # Logging handling logic
│   ├── models            # Managing each DB table model structure and response struct
│   ├── seed              # Managing Seed file for seed generator and integrating test
│   ├── redis             # Redis connection and operations
│   ├── server            # Server setup and routes registration
│   ├── transaction       # Transaction domain (service, repo)
│   ├── user              # User domain (handler, service, repo)
│   ├── wallet            # Wallet domain (handler, service, repo)
│   └── utils             # Utility functions, API response helpers,Custom error handling logic and middlewares
├── tests                 # Unit and integration test utilities, mock services and repositories
│   ├── integrations      # Integrations test files for wallet and transaction
│   ├── mocks             # several mocking files like mock_wallet_service, mock_wallet_repo, mock_transaction_repo, etc
│   └── testutils         # test use helper, test db and test redis related code
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

    - Error: `400 Bad Request`

    ```json
    {
      "status": "error",
      "message": "Invalid email format"
    }
    ```

    - Error: `400 Bad Request`

    ```json
    {
      "status": "error",
      "message": "Password must be at least 6 characters"
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

- **All required token API error**:
  - Error: `401 Unauthorized`

  ```json
  {
    "status": "error",
    "message": "Invalid authorization format"
  }
  ```

  - Error: `401 Unauthorized`

  ```json
  {
    "status": "error",
    "message": "Authorization token is required"
  }
  ```

  - Error: `401 Unauthorized`

    ```json
    {
      "status": "error",
      "message": "Invalid token"
    }
    ```

- **All Internal Server Error**:
  - Error: `500 Internal Server Error`

  ```json
  {
    "status": "error",
    "message": "Internal Server Error"
  }
  ```

  - Error: `500 Internal Server Error`

    ```json
    {
      "status": "error",
      "message": "Database operation failed"
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
    - **Redis Caching**: Wallet numbers and transaction history are cached in Redis to reduce database load and improve response times, especially for frequently accessed data like balance checks and transaction history.
    - **Database Optimization**: The PostgreSQL schema uses normalization and indexing to ensure efficient querying and reduce latency.

2. **Scalability**:
    - The project uses well-defined layers (handler, service, repository), making it easier to scale and add new features without affecting other parts of the system.
    - Redis caching ensures the system can handle a growing user base without overloading the database for frequent queries.

3. **Security**:
    - **JWT Authentication**: Secure authentication is implemented using JWT with a 72-hour token expiration to ensure only valid tokens are used.
    - **Token Blacklisting**: Logged-out tokens are blacklisted in Redis, preventing their reuse and enhancing security.
    - **Input Validation and Error Handling**: All API endpoints validate inputs and handle errors securely to prevent unauthorized access or malicious data entry.

4. **Error Handling and Logging**:
    - Errors are logged centrally, helping developers debug and track issues efficiently.
    - Service and repository layer errors are handled gracefully, with detailed logs written for troubleshooting.

5. **Test Coverage**:
    - **Unit Tests**: Critical features like wallet operations and authentication are covered to ensure reliability.
    - **Integration Tests**: Real instances of PostgreSQL and Redis are used in integration tests to simulate production environments, ensuring system stability.
    - The most important parts of the application are thoroughly tested, although not every part is covered due to time constraints.

6. **Simplicity and Maintainability**:
    - The project is designed with a clear separation of concerns, making it easy to maintain and extend.
    - **Dependency Injection** is used to decouple services and repositories, simplifying testing and enhancing flexibility.

### Explaining Any Decisions You Made

1. **JWT for Authentication**:
   - JWT was chosen for its stateless nature, scalability, and simplicity. It ensures security with token expiration set to 72 hours. Redis was used to implement token blacklisting for invalidating tokens upon logout. A refresh token mechanism was not included due to time constraints but could be added later.

2. **Redis for Performance and Security**:
   - Redis was used for caching frequently accessed data like wallet numbers and transaction history, reducing database load. It also supports token blacklisting for securing user sessions after logout, leveraging Redis’s fast access and TTL features.

3. **Dependency Injection**:
   - The project leverages **dependency injection** to manage dependencies between services and handlers. This approach allows for better testability, as services can easily be swapped out for mocks during testing. It also ensures that dependencies are centralized, making it easier to manage and extend the application.


4. **Repository-Service Pattern**:
   - The project follows the repository-service pattern, separating data access logic (repository) from business logic (service). Services coordinate between repositories and other services, like recording transactions after wallet operations. This pattern enhances testability, maintainability, and decoupling.

5. **Error Handling and Logging**:
   - A centralized error-handling mechanism was implemented via middleware. It standardizes error responses across the application and logs critical internal errors for debugging purposes.

6. **Transaction History Design**:
   - **DB Transaction**: All wallet operations (deposit, withdraw, transfer) are wrapped in database transactions to ensure data consistency. If any step fails, the entire operation is rolled back.
   - **From/To Wallet Number**: The transaction design uses both `from_wallet_number` and `to_wallet_number` for clarity, security, and flexibility. This allows the system to easily support more complex financial operations like multi-wallet users.

7. **Wallet Number Generation**:
   - Wallet numbers are generated uniquely upon wallet creation, similar to bank account numbers. A simple algorithm combining user ID, timestamp, and a random string was used for this project. More advanced methods could be implemented for production use.

8. **Simple Authentication**:
   - Token-based authentication was implemented for simplicity, without refresh tokens. Users must re-login after 72 hours. Redis-based token blacklisting ensures compromised tokens can be invalidated before they expire.

9. **Testing Strategy**:
   - Unit tests were prioritized for key functionalities like wallet services and handlers. Integration tests were performed using `testcontainers-go` to verify interactions with Redis and PostgreSQL. Full coverage wasn't achieved due to time constraints, but core features are well-tested.

10. **Security Considerations**:
   - Passwords are securely hashed, and sensitive operations like transfers and balance checks are protected by JWT authentication. Redis helps manage token blacklisting, ensuring tokens can be revoked upon logout.

### Features Not Included in the Submission

1. **Refresh Tokens for Authentication**:
    - I implemented **JWT-based authentication** but did not include **refresh tokens** due to time constraints. In production, refresh tokens would enhance the user experience by allowing session extension without frequent re-logins.

2. **Advanced Wallet Number Generation**:
    - The wallet number generation is currently basic (user ID, timestamp, random string). For a production system, a more robust approach like GUID or centralized sequence generation would ensure uniqueness and prevent collisions.

3. **Comprehensive Data Validation and Input Sanitization**:
    - Basic validation is in place, but the project lacks comprehensive **input validation** and **sanitization**. This could be improved to ensure the system is more resilient to invalid or malicious data inputs.

4. **End-to-End Testing**:
    - Full end-to-end tests (e.g., route tests) were not included due to time constraints. However, the core functionality is covered with unit tests for handlers and integration tests for services.

5. **Test Code Structure**:
    - While the core scenarios are tested, not every test follows clean code principles. Given more time, I would refine the test structure to ensure better maintainability.

6. **Simplified Database Design**:
    - The database schema is simplified to include only the essential tables like `transactions` and `wallets`. A more complex relational design could be added in the future as needed.

7. **Transaction History Design (Code Location)**:
    - Transaction history logic is in `wallet_handlers.go`, although it ideally belongs in its own handler and service. Due to potential circular dependencies and time constraints, this was left in the wallet handler. A shared service layer could address this in future iterations.

8. **Authentication Security Check**:
    - The current JWT middleware does not verify the existence of the user in the database. A more secure implementation would add a user existence check to prevent unauthorized access, improving the overall security of the system.

### Areas for Future Improvement

1. **Code Maintainability**:
    - **Database Code and Repository Layer**: The current repository layer requires writing raw SQL queries, leading to duplication and verbosity. Introducing an ORM (e.g., GORM) or a query builder could simplify database operations and improve maintainability.
    - **Service and Repository Code Organization**: The service layer can be further refactored to improve clarity and modularity, especially where services interact with multiple repositories.

2. **Event Sourcing / Event-Driven Architecture for Transaction Handling**:
    - Currently, **database transactions** ensure consistency during wallet operations, but an **event-driven architecture** could reduce latency and enhance scalability. For example, emitting events for each operation (deposit, withdrawal, transfer) and processing them asynchronously would improve overall performance.

3. **Testing Coverage and Depth**:
    - **Expanded Test Coverage**: While core functionalities are tested, additional unit tests and integration tests are needed for edge cases and multi-service interactions.
    - **End-to-End (E2E) Testing**: Adding E2E tests for API endpoints would enhance confidence that the system works across the full stack (request to database).

4. **Performance Optimization**:
    - **Database Query Optimization**: Optimizing queries and adding indexes, particularly for transaction history retrieval (involving joins), will ensure the system scales well. Caching strategies could further enhance performance.
    - **Rate Limiting**: Implementing rate limiting would secure the system against abuse and malicious activities, protecting it from DDoS attacks or high load.

5. **Security Enhancements**:
    - **Refresh Tokens**: Adding refresh tokens would allow users to stay logged in longer, improving security and user experience by enabling token renewal without re-login.
    - **Two-Factor Authentication (2FA)**: Adding 2FA, especially for monetary transactions, would increase the security of user accounts.

6. **Code Cleanliness**:
    - **Refactoring for Simplicity**: Breaking down larger functions into smaller, reusable components and improving readability would enhance code maintainability.
    - **DRY Principle**: Refactoring common logic (e.g., database queries) into helper functions or shared service methods would reduce duplication and improve maintainability.

7. **Future Feature Expansions**:
    - **Scalability Considerations**: As the system grows, separating services into microservices would enable independent scaling of wallet management, transactions, and authentication.
    - **Audit Logs**: Adding audit logs for deposits, withdrawals, and transfers would improve security and troubleshooting.
    - **Improved Error Logging**: Enhancing logs with user context and request metadata would make debugging easier in production.

8. **Reusing Wallet Middleware for Other Operations**:
    - The **wallet middleware** currently caches wallet numbers for transaction history. Expanding its use to other wallet operations (deposit, withdraw, transfer) could improve performance and code reusability.

9. **Advanced Wallet Number Generation**:
    - The current wallet number generation mechanism is basic and may not scale. A more robust solution, such as using **GUIDs** or a **centralized sequence generator**, would ensure globally unique and collision-free wallet numbers.

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
