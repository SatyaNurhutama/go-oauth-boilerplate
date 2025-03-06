# Go OAuth Boilerplate

A **Go** (Golang) boilerplate for implementing **OAuth 2.0** and **OpenID Connect** with **Google Login**, **Redis**, and **PostgreSQL**. This project provides a robust foundation for building secure authentication systems with features like **login**, **registration**, **token refresh**, and **logout**.

---

## Features

- **OAuth 2.0 and OpenID Connect**:
  - Login with Google (OAuth 2.0 + OpenID Connect).
  - Secure token generation and validation.

- **Authentication**:
  - Normal login with email and password.
  - User registration with email, password, and name.
  - Token-based authentication using **JWT** (JSON Web Tokens).

- **Token Management**:
  - Access tokens for short-term authentication.
  - Refresh tokens for long-term session management.
  - Token blacklisting for secure logout.

- **Database**:
  - **PostgreSQL** for persistent storage of user data.
  - **Redis** for caching refresh tokens and managing token blacklisting.

- **Framework**:
  - Built with **Gin**, a high-performance HTTP web framework for Go.

- **Security**:
  - Password hashing using **bcrypt**.
  - Secure token storage and validation.

---

## Tech Stack

- **Backend**: Go (Golang)
- **Framework**: Gin
- **Database**: PostgreSQL
- **Cache**: Redis
- **Authentication**: JWT, OAuth 2.0, OpenID Connect
- **Password Hashing**: bcrypt

---

## Prerequisites

Before running the project, ensure you have the following installed:

1. **Go** (version 1.20 or higher)
2. **PostgreSQL** (version 13 or higher)
3. **Redis** (version 6 or higher)
4. **Google OAuth Credentials**:
   - Create a project in the [Google Cloud Console](https://console.cloud.google.com/).
   - Enable the **Google OAuth 2.0 API**.
   - Create OAuth credentials (Client ID and Client Secret).
   - Set the authorized redirect URI (e.g., `http://localhost:8080/api/auth/login/google/callback`).

---

## Setup

### 1. Clone the Repository

```bash
git clone https://github.com/satyanurhutama/go-oauth-boilerplate.git
cd go-oauth-boilerplate
```

### 2. Set Up Environment Variable
```bash
# Server
PORT=8080

# Database (PostgreSQL)
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=postgres
DB_NAME=auth_project

# Redis
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASSWORD=

# JWT
JWT_SECRET=your_jwt_secret
JWT_EXPIRATION=24h

# Google OAuth
GOOGLE_CLIENT_ID=your_google_client_id
GOOGLE_CLIENT_SECRET=your_google_client_secret
GOOGLE_REDIRECT_URL=http://localhost:8080/api/auth/login/google/callback
```

### 3. Create users table
```bash
CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    email VARCHAR(255) UNIQUE NOT NULL,
    password VARCHAR(255) NOT NULL,
    name VARCHAR(255),
    provider VARCHAR(50) NOT NULL, -- e.g., "google", "email"
    provider_id VARCHAR(255) -- Unique ID from the provider (e.g., Google ID)
);
```

### 4. Start the Server
```bash
go run cmd/server/main.go
```

## Project Structure
```bash
go-oauth-boilerplate/
├── cmd/
│   └── server/              # Main entry point for the application
├── internal/
│   ├── auth/                # Authentication-related logic
│   │   ├── handler/         # HTTP handlers (Gin)
│   │   ├── repository/      # Database and Redis interactions
│   │   ├── usecase/         # Business logic
│   │   └── entity/          # Domain models
│   │   ├── dto/             # DTO
│   ├── config/              # Configuration management
│   ├── middleware/          # Custom middleware (e.g., auth middleware)
│   └── utils/               # Utility functions (e.g., JWT, hashing)
├── pkg/                     # Shared packages (e.g., Redis, PostgreSQL clients)
├── .env                     # Environment variables
├── go.mod                   # Go module file
└── README.md                # Project documentation
```
