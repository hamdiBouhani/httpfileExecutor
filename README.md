# httpfileExecutor
## First time only — sets up module and pulls deps
make init

## Run with default placeholder credentials
make run

## Run with your real credentials inline
make run \
  OAUTH_CLIENT_ID="real-id" \
  OAUTH_CLIENT_SECRET="real-secret" \
  OAUTH_TOKEN_URL="https://auth.example.com/oauth/token" \
  API_TOKEN="my-bearer-token" \
  HTTP_FILE="api_requests.http"

## Build a standalone binary
make build

## Build and immediately execute
make build-run API_TOKEN="abc123"

## Cleanup
make clean