# 2FA System

A secure and scalable Two-Factor Authentication (2FA) microservice written in Go. This service provides endpoints to send and verify one-time codes via email, with advanced security features and monitoring capabilities.

## 🚀 Features

- **Send 2FA codes via email** using SMTP
- **Verify codes** with secure validation
- **Rate limiting** (per email and per IP)
- **Brute force protection** with configurable thresholds
- **Real-time monitoring panel** with beautiful UI
- **Multiple storage backends** (in-memory and Redis)
- **RESTful API** with clear endpoints
- **Configurable** via environment variables
- **Production-ready** with proper error handling

## 🔒 Security Features

- **Per-Email Rate Limiting**: 1 request per minute per email
- **Per-IP Rate Limiting**: 1 request per minute per IP address
- **Brute Force Protection**: Blocks after 5 failed attempts for 10 minutes
- **Secure Code Generation**: 6-digit random codes
- **Automatic Code Expiration**: Codes expire after use or timeout

## 📊 Monitor Panel

Access the real-time monitoring panel at `/monitor` to view:
- **Per-Email Rate Limits**: Currently rate-limited emails
- **Per-IP Rate Limits**: Currently rate-limited IP addresses
- **Failed Attempts**: Users with failed verification attempts
- **Blocked Users**: Users currently blocked due to brute force

The monitor panel features:
- **Real-time updates** (auto-refreshes every 5 seconds)
- **Beautiful, responsive UI** with modern styling
- **JSON API endpoint** at `/monitor-data` for programmatic access
- **Visual status indicators** for blocked/active users

## 🗄️ Storage Options

### In-Memory Storage (Default)
- Fast and simple for development
- Data lost on server restart
- Good for testing and small deployments

### Redis Storage (Production)
- Persistent across server restarts
- Scalable for multiple instances
- Automatic TTL for data expiration
- Configure with environment variables

## 📋 API Endpoints

### `POST /send-code`

Send a 2FA code to a user's email.

**Request Body:**
```json
{
  "email": "user@example.com"
}
```

**Response:**
```json
{
  "message": "Código enviado com sucesso"
}
```

**Rate Limits:**
- 1 request per minute per email
- 1 request per minute per IP

---

### `POST /verify-code`

Verify a 2FA code for a user.

**Request Body:**
```json
{
  "email": "user@example.com",
  "code": "123456"
}
```

**Response (success):**
```json
{
  "success": true,
  "message": "Código verificado com sucesso"
}
```

**Response (failure):**
```json
{
  "success": false,
  "message": "Código inválido ou expirado"
}
```

**Security:**
- Blocks after 5 failed attempts for 10 minutes
- Same rate limits as send-code endpoint

---

### `GET /monitor`

Real-time monitoring panel with beautiful UI.

---

### `GET /monitor-data`

JSON API for monitoring data.

**Response:**
```json
{
  "emailRateLimit": {
    "user@example.com": {
      "blocked": true,
      "blockedUntil": "2024-01-01T12:00:00Z"
    }
  },
  "ipRateLimit": {
    "192.168.1.1": {
      "blocked": true,
      "blockedUntil": "2024-01-01T12:00:00Z"
    }
  },
  "failedAttempts": {
    "user@example.com": {
      "attempts": 3
    }
  },
  "blockedUntil": {
    "user@example.com": {
      "blocked": true,
      "blockedUntil": "2024-01-01T12:00:00Z"
    }
  }
}
```

## 🛠️ Getting Started

### Prerequisites

- Go 1.21+
- SMTP server (e.g., Gmail, Mailgun, etc.)
- Redis (optional, for production)

### Installation

1. **Clone the repository:**
   ```sh
   git clone https://github.com/joaofaugusto/2fa.git
   cd 2fa-system
   ```

2. **Install dependencies:**
   ```sh
   go mod tidy
   ```

3. **Set environment variables:**

   Create a `.env` file:
   ```env
   # SMTP Configuration
   SMTP_HOST=smtp.example.com
   SMTP_PORT=587
   SMTP_USER=your_smtp_user
   SMTP_PASSWORD=your_smtp_password
   FROM_EMAIL=your@email.com
   APP_ENV=development
   
   # Storage Configuration (optional)
   STORAGE_TYPE=memory  # or "redis"
   REDIS_ADDR=localhost:6379  # only if using Redis
   REDIS_PASSWORD=  # only if using Redis
   ```

4. **Build and run:**
   ```sh
   go build -o main .
   ./main
   ```

   The server will start on `http://localhost:8080`.

## 🔧 Configuration

### Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `SMTP_HOST` | SMTP server hostname | Required |
| `SMTP_PORT` | SMTP server port | Required |
| `SMTP_USER` | SMTP username | Required |
| `SMTP_PASSWORD` | SMTP password | Required |
| `FROM_EMAIL` | Sender email address | Required |
| `APP_ENV` | Application environment | `development` |
| `STORAGE_TYPE` | Storage backend (`memory` or `redis`) | `memory` |
| `REDIS_ADDR` | Redis server address | `localhost:6379` |
| `REDIS_PASSWORD` | Redis password | Empty |

### Storage Configuration

#### In-Memory Storage
```env
STORAGE_TYPE=memory
```

#### Redis Storage
```env
STORAGE_TYPE=redis
REDIS_ADDR=localhost:6379
REDIS_PASSWORD=your_redis_password
```

## 📁 Project Structure

```
.
├── config/          # Configuration loading
├── handlers/        # HTTP handlers with security features
├── models/          # Request/response models
├── services/        # Business logic (code generation, email sending)
├── storage/         # Storage interfaces and implementations
│   ├── interface.go # Storage interface
│   ├── memory.go    # In-memory storage
│   ├── redis.go     # Redis storage
│   └── factory.go   # Storage factory
├── templates/       # HTML templates
│   └── monitor.html # Monitor panel template
├── main.go          # Entry point
└── README.md        # This file
```

## 🚀 Production Deployment

### Using Redis (Recommended)

1. **Install and start Redis:**
   ```sh
   # Ubuntu/Debian
   sudo apt install redis-server
   sudo systemctl start redis-server
   
   # macOS
   brew install redis
   brew services start redis
   ```

2. **Set environment variables:**
   ```env
   STORAGE_TYPE=redis
   REDIS_ADDR=localhost:6379
   ```

3. **Run the application:**
   ```sh
   go build -o main .
   ./main
   ```

### Docker Deployment

Create a `Dockerfile`:
```dockerfile
FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY . .
RUN go mod download
RUN go build -o main .

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/main .
EXPOSE 8080
CMD ["./main"]
```

## 🔍 Monitoring and Debugging

### Monitor Panel
- Access at `http://localhost:8080/monitor`
- Real-time updates every 5 seconds
- Beautiful, responsive UI

### Logs
The application logs important events:
- Rate limit violations
- Brute force attempts
- Email sending errors
- Storage connection issues

### Health Checks
- Monitor endpoint returns 200 OK when healthy
- Check `/monitor-data` for detailed system status

## 🔒 Security Considerations

### Rate Limiting
- **Per-email**: Prevents email abuse
- **Per-IP**: Prevents IP-based attacks
- **Configurable**: Easy to adjust limits

### Brute Force Protection
- **5 failed attempts** = 10-minute block
- **Automatic reset** on successful verification
- **Persistent across restarts** (with Redis)

### Code Security
- **6-digit codes** with high entropy
- **Single-use**: Codes expire after verification
- **10-minute TTL**: Automatic expiration

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests if applicable
5. Submit a pull request

## 📄 License

This project is licensed under the terms of the [LICENSE](LICENSE) file.

## ⚠️ Disclaimer

**For production use:**
- Use Redis storage for persistence
- Configure proper SMTP settings
- Set up monitoring and alerting
- Consider using HTTPS in production
- Implement proper logging and metrics
- Add authentication to the monitor panel if needed

---

**Happy coding! 🔐** 