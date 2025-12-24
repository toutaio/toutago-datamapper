# Credentials Management Example

This example demonstrates **secure credential management** with toutago-datamapper.

## What This Example Shows

- **Environment Variables**: Using `${VAR_NAME}` syntax
- **Default Values**: Providing fallbacks `${VAR_NAME:-default}`
- **Credential Files**: Separate secrets from configuration
- **Multi-Environment**: Development, staging, production
- **Security Best Practices**: What to commit, what to ignore

## Security Principles

### ✅ DO
- Store credentials in environment variables
- Use `.env` for local development
- Keep `credentials.yaml` out of git
- Commit `.env.example` as a template
- Use secrets managers in production

### ❌ DON'T
- Commit credentials to git
- Hardcode passwords in config
- Share credential files
- Use production creds in development
- Log or display credentials

## File Structure

```
credentials/
├── config.yaml              # Safe to commit - uses placeholders
├── .env.example            # Safe to commit - template only
├── .env                    # DO NOT COMMIT - local secrets
├── credentials.yaml.example # Safe to commit - template
└── credentials.yaml        # DO NOT COMMIT - actual secrets
```

## Configuration Syntax

### Basic Environment Variable
```yaml
sources:
  mydb:
    connection: "${DB_HOST}"  # Replaced with env var
```

### With Default Value
```yaml
sources:
  mydb:
    connection: "${DB_HOST:-localhost}"  # Uses localhost if not set
```

### Multiple Variables
```yaml
sources:
  postgres:
    adapter: postgres
    connection: "host=${DB_HOST} port=${DB_PORT} dbname=${DB_NAME} user=${DB_USER} password=${DB_PASSWORD}"
```

## Running the Example

### 1. Set Environment Variables

**Option A: Export directly**
```bash
export DB_PATH="./data"
export DB_FORMAT="json"
go run main.go
```

**Option B: Use .env file**
```bash
# Copy template
cp .env.example .env

# Edit .env with your values
# Then source it:
source .env
go run main.go
```

**Option C: Inline**
```bash
DB_PATH=./data DB_FORMAT=json go run main.go
```

## Expected Output

```
=== Credentials Management Example ===

Environment Setup:
  • Set environment variables before running
  • Create .env file for local development
  • Use credentials.yaml for secrets (DO NOT COMMIT!)

Checking environment variables...
   ✓ DB_PATH: SET
   ✓ DB_FORMAT: SET

1. Creating mapper with credential resolution...
   ✓ Mapper created successfully
   ✓ Credentials resolved from environment

2. Performing operations with secure credentials...
   ✓ Created account: admin

3. Fetching account...
   ✓ Retrieved: admin (admin@example.com)

=== Security Best Practices Demonstrated ===

✅ Configuration Separation:
   • config.yaml - Committed to git
   • .env - Local development only
   • credentials.yaml - NEVER commit!

✅ Environment Variables:
   • Production uses real env vars
   • Local uses .env file
   • CI/CD uses secrets manager

✅ Multiple Environments:
   • Development: .env.development
   • Staging: .env.staging
   • Production: System environment
```

## Real-World Examples

### MySQL/PostgreSQL
```yaml
# config.yaml (safe to commit)
sources:
  maindb:
    adapter: postgres
    connection: "host=${DB_HOST} port=${DB_PORT:-5432} dbname=${DB_NAME} user=${DB_USER} password=${DB_PASSWORD} sslmode=${DB_SSL_MODE:-require}"
```

```bash
# .env (DO NOT commit)
DB_HOST=prod-database.example.com
DB_PORT=5432
DB_NAME=myapp
DB_USER=app_user
DB_PASSWORD=super_secret_password
DB_SSL_MODE=require
```

### Redis
```yaml
# config.yaml
sources:
  cache:
    adapter: redis
    connection: "${REDIS_URL}"
    options:
      password: "${REDIS_PASSWORD}"
      db: "${REDIS_DB:-0}"
```

```bash
# .env
REDIS_URL=redis://cache.example.com:6379
REDIS_PASSWORD=redis_secret
REDIS_DB=0
```

### AWS
```yaml
# config.yaml
sources:
  s3storage:
    adapter: s3
    connection: "${AWS_REGION}"
    options:
      access_key: "${AWS_ACCESS_KEY_ID}"
      secret_key: "${AWS_SECRET_ACCESS_KEY}"
      bucket: "${S3_BUCKET}"
```

```bash
# .env
AWS_REGION=us-east-1
AWS_ACCESS_KEY_ID=AKIAIOSFODNN7EXAMPLE
AWS_SECRET_ACCESS_KEY=wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY
S3_BUCKET=my-app-bucket
```

## Multi-Environment Setup

### Development
```bash
# .env.development
DB_HOST=localhost
DB_NAME=myapp_dev
DB_USER=developer
DB_PASSWORD=dev_password
```

### Staging
```bash
# .env.staging
DB_HOST=staging-db.example.com
DB_NAME=myapp_staging
DB_USER=staging_user
DB_PASSWORD=staging_secret
```

### Production
```bash
# Use system environment variables
# Set in Kubernetes secrets, AWS Secrets Manager, etc.
```

## Docker Integration

### docker-compose.yml
```yaml
version: '3.8'
services:
  app:
    build: .
    environment:
      - DB_HOST=postgres
      - DB_PORT=5432
      - DB_NAME=myapp
      - DB_USER=appuser
      - DB_PASSWORD=${DB_PASSWORD}  # From .env file
    env_file:
      - .env
```

## Kubernetes Integration

### ConfigMap (non-sensitive)
```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: app-config
data:
  DB_HOST: "postgres-service"
  DB_PORT: "5432"
  DB_NAME: "myapp"
```

### Secret (sensitive)
```yaml
apiVersion: v1
kind: Secret
metadata:
  name: app-secrets
type: Opaque
data:
  DB_PASSWORD: <base64-encoded-password>
  REDIS_PASSWORD: <base64-encoded-password>
```

### Pod
```yaml
apiVersion: v1
kind: Pod
metadata:
  name: myapp
spec:
  containers:
  - name: app
    image: myapp:latest
    envFrom:
    - configMapRef:
        name: app-config
    - secretRef:
        name: app-secrets
```

## CI/CD Integration

### GitHub Actions
```yaml
name: Deploy
on: [push]
jobs:
  deploy:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v2
      - name: Run tests
        env:
          DB_HOST: localhost
          DB_USER: test
          DB_PASSWORD: ${{ secrets.DB_PASSWORD }}
        run: go test ./...
```

### GitLab CI
```yaml
deploy:
  script:
    - go build
  variables:
    DB_HOST: $DB_HOST
    DB_PASSWORD: $DB_PASSWORD
```

## Best Practices

### 1. Never Commit Secrets
```bash
# .gitignore
.env
.env.local
.env.*.local
credentials.yaml
secrets.yaml
*.key
*.pem
```

### 2. Use Strong Passwords
```bash
# Generate secure passwords
openssl rand -base64 32
```

### 3. Rotate Credentials
- Change passwords regularly
- Use different passwords per environment
- Revoke old credentials after rotation

### 4. Audit Access
- Log credential usage
- Monitor for unauthorized access
- Alert on failures

### 5. Use Secrets Managers
- AWS Secrets Manager
- HashiCorp Vault
- Azure Key Vault
- Google Secret Manager

## Troubleshooting

### Missing Environment Variable
```
Error: environment variable DB_PASSWORD is not set
```
**Solution**: Set the variable or provide a default

### Wrong Environment
```
Error: permission denied for database myapp_production
```
**Solution**: Check you're using the right .env file

### Leaked Credentials
**Solution**: 
1. Rotate credentials immediately
2. Remove from git history: `git filter-branch`
3. Use `git-secrets` to prevent future leaks

## Security Checklist

- [ ] Credentials not in config files
- [ ] .env in .gitignore
- [ ] Different passwords per environment
- [ ] Production uses secrets manager
- [ ] No credentials in logs
- [ ] Regular password rotation
- [ ] Limited credential access
- [ ] Audit logging enabled

## Next Steps

- Integrate with secrets manager (Vault, AWS Secrets Manager)
- Implement credential rotation
- Add audit logging
- Set up monitoring and alerts
- Use short-lived credentials
