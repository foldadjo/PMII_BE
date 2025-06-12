# PMII Backend API

This is the backend API for the PMII (Pergerakan Mahasiswa Islam Indonesia) application. It provides authentication and user management functionality.

## Features

- User registration and authentication
- Pengurus (administrator) management
- Password reset functionality
- JWT-based authentication
- MongoDB Atlas integrations

## Prerequisites

- Go 1.21 or later
- MongoDB Atlas account
- Vercel account (for deployment)

## Environment Variables

Create a `.env` file in the root directory with the following variables:

```
MONGODB_URI=your_mongodb_connection_string
JWT_SECRET=your_jwt_secret_key
PORT=3000
```

## API Endpoints

### Authentication

- `POST /api/auth/register` - Register a new user
- `POST /api/auth/login` - Login as a regular user
- `POST /api/auth/login-pengurus` - Login as a pengurus
- `POST /api/auth/forgot-password` - Request password reset
- `POST /api/auth/reset-password` - Reset password with token

### Pengurus Management

- `POST /api/create-pengurus` - Create a new pengurus entry (requires authentication)

## Database Schema

### Users Collection

```json
{
  "_id": "ObjectId",
  "email": "string (unique)",
  "password_hash": "string",
  "full_name": "string",
  "anggota_id": "string (nullable)",
  "role": "enum ['null', 'mapaba', 'pkd', 'pkl', 'pkn']",
  "code_kepengurusan": "string (nullable)",
  "active": "boolean",
  "created_at": "timestamp",
  "updated_at": "timestamp"
}
```

### Pengurus Collection

```json
{
  "_id": "ObjectId",
  "user_id": "ObjectId",
  "level": "enum ['PB', 'PKC', 'PC', 'PK', 'PR']",
  "wilayah": "string (nullable)",
  "cabang": "string (nullable)",
  "komisariat": "string (nullable)",
  "jabatan": "string",
  "aktif": "boolean",
  "mulai_jabatan": "date",
  "akhir_jabatan": "date",
  "created_at": "timestamp",
  "updated_at": "timestamp"
}
```

### Domains Collection

```json
{
  "_id": "ObjectId",
  "domain": "string (unique)",
  "is_active": "boolean",
  "created_at": "timestamp"
}
```

### Reset Password Tokens Collection

```json
{
  "_id": "ObjectId",
  "user_id": "ObjectId",
  "token": "string",
  "expires_at": "timestamp",
  "created_at": "timestamp"
}
```

## Security

- Passwords are hashed using bcrypt with cost factor 12
- JWT tokens are used for authentication
- Input validation is performed on all endpoints
- CORS is configured to allow specific origins
- Environment variables are used for sensitive data

## Development

1. Clone the repository
2. Install dependencies: `go mod download`
3. Create `.env` file with required variables
4. Run the server: `go run main.go`

## Deployment

The application is designed to be deployed on Vercel as serverless functions. Each API endpoint is implemented as a separate function in the `/api` directory.

## License

This project is proprietary and confidential. 