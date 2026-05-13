# Golang Order Manager API

A production-inspired REST API built with Go, focused on authentication, session management, and scalable architecture.

## 🚀 Features

- JWT authentication
- Refresh token rotation
- Session management
- Soft delete
- Layered architecture
- PostgreSQL integration
- Rate limiting
- Secure password hashing
- Structured error handling

## 🛠️ Tech Stack

- Go
- Echo
- PostgreSQL
- JWT
- Docker
- golang-migrate

## 📁 Project Structure

/internal
  /handlers
  /services
  /repository
  /security
  /middleware
  /responses

## 🔐 Authentication Flow

- Access Token (JWT)
- Refresh Token Rotation
- Session revocation
- Logout all sessions

## 📌 API Endpoints

### Auth

POST /auth/register
POST /auth/login
POST /auth/refresh
POST /auth/logout
POST /auth/logout-all

### Users

GET /users/me
PATCH /users/me
DELETE /users/me

## ⚙️ Running Locally

```bash
make run