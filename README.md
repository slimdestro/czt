# cczTest | Golang
Reference: https://drive.google.com/file/d/12uM3gkuGfqF161KdRzWOF7LEuy_5aIIy/view?usp=sharing

## Project Structure

* **backend**: Pure REST API on port 8081. Handles Google OAuth2 and MySQL logic.
* **frontend**: Web server on port 8080. Handles UI rendering and session cookies.

---

## Setup and Run

### 1. Environment Config

Create a `.env` file in the backend folder:

```env
DB_DSN=user:password@tcp(127.0.0.1:3306)/dbName
GOOGLE_CLIENT_ID=__id.apps.googleusercontent.com
GOOGLE_CLIENT_SECRET=
GOOGLE_REDIRECT_URL=http://localhost:8081/api/auth/google/callback
FRONTEND_URL=http://localhost:8080

```

### 2. Database Migration

First create the database in MySQL then run this from the frontend folder to create tables:

```bash
go run cmd/migration/main.go

```

### 3. Running the App

**Start Backend:**

```bash
cd backend
go run main.go

```

**Start Frontend:**

```bash
cd frontend
go run main.go

```

---

## Architecture Note

Project uses a Client-Side Session to handle the port gap between 8080 and 8081:

* **Frontend:** Gatekeeper that manages the session_email cookie and controls UI access.
* **Backend:** Vault that gives data only when the X-User-Email header is sent.
* **log/slog:** logs like this makes sense to machines: {"time":"2026-01-01T14:00:00Z","level":"ERROR","msg":"Login failed","user":"user@example.com","ip":"1.1.1.1"}. In distributed systems, we always use otel or some other sdk for logs/trace/metric exports.

* **The Bridge:** After Google login the backend redirects back to frontend using URL Query Params like ?email=... to set the session.

### Industry Standard (JWT)

In real production apps we use JWT instead of simple cookies. Unlike email cookies which anyone can edit in the browser console a JWT is cryptographically signed by the backend. The frontend sends this token in an Authorization header so the backend can verify identity securely.