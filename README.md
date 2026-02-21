# ThreadCraft Backend

Go + PostgreSQL REST API for the ThreadCraft Flutter app.

## Stack
- **Go 1.22** — single binary, fast cold starts
- **chi** — lightweight router
- **pgx v5** — PostgreSQL driver
- **golang-migrate** — SQL migrations
- **AWS SDK v2** — Cloudflare R2 presigned uploads (S3-compatible)
- **golang-jwt** — JWT auth

## Local Development

### Prerequisites
- Go 1.22+
- PostgreSQL 15+
- A Cloudflare R2 bucket (free tier covers this app easily)

### Setup

```bash
# 1. Clone and install deps
git clone <your-repo>
cd threadcraft-backend
go mod tidy

# 2. Create local database
createdb threadcraft

# 3. Configure environment
cp .env.example .env
# Edit .env with your values

# 4. Run
go run ./cmd/api
```

Server starts at `http://localhost:8080`.
Migrations run automatically on startup.

### Health check
```
GET /health → {"status":"ok"}
```

## API Routes

```
POST   /v1/auth/register
POST   /v1/auth/login
POST   /v1/auth/refresh
DELETE /v1/auth/logout          (auth required)

GET    /v1/users/me             (auth required)
PATCH  /v1/users/me             (auth required)

GET    /v1/projects             (auth required)
POST   /v1/projects             (auth required)
GET    /v1/projects/:id         (auth required)
PATCH  /v1/projects/:id         (auth required)
DELETE /v1/projects/:id         (auth required)
GET    /v1/projects/:id/export  (auth required) ?format=txt|json
GET    /v1/projects/:id/progress (auth required)
PUT    /v1/projects/:id/progress (auth required)

POST   /v1/uploads/presign      (auth required)
```

## Deploy to Railway

1. Push to GitHub
2. New project → Deploy from GitHub repo
3. Add PostgreSQL plugin (Railway provides `DATABASE_URL` automatically)
4. Set environment variables in Railway dashboard:
   ```
   JWT_SECRET=<openssl rand -hex 32>
   R2_ACCOUNT_ID=...
   R2_ACCESS_KEY_ID=...
   R2_SECRET_ACCESS_KEY=...
   R2_BUCKET_NAME=threadcraft-images
   R2_PUBLIC_URL=https://pub-xxx.r2.dev
   ```
5. Railway detects the Dockerfile and builds automatically

## Cloudflare R2 Setup

1. Cloudflare dashboard → R2 → Create bucket → `threadcraft-images`
2. Enable public access on the bucket → copy the public URL → set `R2_PUBLIC_URL`
3. R2 → Manage R2 API tokens → Create token with Object Read & Write
4. Copy Account ID, Access Key ID, Secret Access Key → set env vars

## Connect Flutter app

In `lib/core/api/api_client.dart`, update the base URL:
```dart
const _kBaseUrl = String.fromEnvironment('API_BASE_URL',
    defaultValue: 'https://your-railway-app.up.railway.app/v1');
```

Or pass it at build time:
```bash
flutter build ios --dart-define=API_BASE_URL=https://your-app.up.railway.app/v1
```
