# Authentication Flow

This document describes every authentication flow in the API using sequence diagrams.

---

## 1. Register

```mermaid
sequenceDiagram
    autonumber
    actor Client
    participant API
    participant DB

    Client->>API: POST /auth/register<br/>{email, username, password}
    API->>API: Validate fields (email, username, password required)

    API->>DB: SELECT — email already in use?
    DB-->>API: exists / not exists

    alt email already in use
        API-->>Client: 400 email already in use
    end

    API->>DB: SELECT — username already in use?
    DB-->>API: exists / not exists

    alt username already in use
        API-->>Client: 400 username already in use
    end

    API->>API: bcrypt hash password (cost 14)
    API->>DB: INSERT user
    DB-->>API: created user

    API-->>Client: 204 No Content
```

---

## 2. Login

```mermaid
sequenceDiagram
    autonumber
    actor Client
    participant API
    participant DB

    Client->>API: POST /auth/login<br/>{email, password}
    API->>API: Validate fields

    API->>DB: SELECT user WHERE email = ?
    DB-->>API: user row (with hashed password)

    alt user not found
        API-->>Client: 404 user not found
    end

    API->>API: bcrypt.Compare(storedHash, password)

    alt wrong password
        API-->>Client: 401 invalid credentials
    end

    API->>API: Generate JWT (HS256, signed with SECRET_KEY)<br/>payload: {userID, exp: now + TOKEN_EXPIRATION_MINUTES}
    API->>API: Generate refresh token (crypto/rand 32 bytes → hex)
    API->>API: SHA-256 hash the refresh token
    API->>DB: INSERT refresh_tokens<br/>{user_id, hash, expires_at: now + REFRESH_TOKEN_EXPIRATION_MINUTES}
    DB-->>API: ok

    API-->>Client: 200 {access_token, refresh_token, user}

    Note over Client: Client stores both tokens.<br/>Raw refresh token is never stored in DB.
```

---

## 3. Authenticated Request (Middleware)

```mermaid
sequenceDiagram
    autonumber
    actor Client
    participant Middleware
    participant Handler

    Client->>Middleware: Any protected request<br/>Authorization: Bearer <access_token>

    alt no Authorization header
        Middleware-->>Client: 401 Unauthorized
    end

    Middleware->>Middleware: Strip "Bearer " prefix
    Middleware->>Middleware: jwt.Parse(token, SECRET_KEY)

    alt invalid signature / malformed / expired
        Middleware-->>Client: 401 Unauthorized
    end

    Middleware->>Middleware: Extract userID from claims
    Middleware->>Handler: context.Set("userID", uuid)
    Handler-->>Client: Handler response
```

---

## 4. Token Refresh (Rotation)

```mermaid
sequenceDiagram
    autonumber
    actor Client
    participant API
    participant DB

    Client->>API: POST /auth/refresh<br/>{refresh_token}
    API->>API: SHA-256 hash the incoming refresh token

    API->>DB: SELECT WHERE token = hash
    DB-->>API: {user_id, revoked_at, expires_at}

    alt token not found
        API-->>Client: 401 refresh token not found
    end

    alt revoked_at IS NOT NULL or now > expires_at
        API-->>Client: 401 refresh token expired
    end

    API->>DB: SELECT user WHERE id = user_id
    DB-->>API: user

    API->>API: Generate new JWT
    API->>DB: UPDATE refresh_tokens SET revoked_at = now()<br/>WHERE token = old_hash
    DB-->>API: ok

    API->>API: Generate new refresh token (crypto/rand)
    API->>API: SHA-256 hash the new token
    API->>DB: INSERT refresh_tokens<br/>{user_id, new_hash, new expires_at}
    DB-->>API: ok

    API-->>Client: 200 {access_token, refresh_token, user}

    Note over Client,DB: Old refresh token is revoked immediately.<br/>Each token can only be used once.
```

---

## 5. Logout (Single Session)

```mermaid
sequenceDiagram
    autonumber
    actor Client
    participant API
    participant DB

    Client->>API: POST /auth/logout<br/>Authorization: Bearer <access_token><br/>{refresh_token}

    Note over API: Auth middleware validates the JWT first

    API->>API: SHA-256 hash the refresh token

    API->>DB: UPDATE refresh_tokens SET revoked_at = now()<br/>WHERE token = hash

    alt token not found or already revoked
        API-->>Client: 401 Unauthorized
    end

    DB-->>API: ok
    API-->>Client: 200 Logout successful
```

---

## 6. Logout All (All Sessions)

```mermaid
sequenceDiagram
    autonumber
    actor Client
    participant Middleware
    participant API
    participant DB

    Client->>Middleware: POST /auth/logout-all<br/>Authorization: Bearer <access_token>
    Middleware->>Middleware: Validate JWT → extract userID
    Middleware->>API: context with userID

    API->>DB: UPDATE refresh_tokens SET revoked_at = now()<br/>WHERE user_id = ? AND revoked_at IS NULL
    DB-->>API: N rows updated

    API-->>Client: 200 Logout successful

    Note over Client,DB: All active sessions across all devices are terminated.<br/>Existing JWTs remain valid until they expire naturally.
```

---

## Token Lifecycle

```mermaid
stateDiagram-v2
    [*] --> Active : Login / Refresh

    Active --> Revoked : Logout (single)\nor used in Refresh
    Active --> Revoked : Logout-all
    Active --> Expired : expires_at reached

    Revoked --> [*]
    Expired --> [*]

    note right of Active
        Raw token held by client.
        SHA-256 hash stored in DB.
    end note
```
