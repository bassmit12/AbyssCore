#!/bin/sh
# Keycloak init script for AbyssCore
# Creates realm, client, roles, and test user

KEYCLOAK_URL="${KEYCLOAK_URL:-http://localhost:8080}"
REALM="${REALM:-abysscore}"
CLIENT_ID="${CLIENT_ID:-abysscore-frontend}"
TEST_USER="${TEST_USER:-testplayer}"
TEST_PASSWORD="${TEST_PASSWORD:-testpass}"

echo "Waiting for Keycloak..."
until curl -sf "${KEYCLOAK_URL}/realms/master" > /dev/null; do
  sleep 2
done

echo "Getting admin token..."
TOKEN=$(curl -sf -X POST "${KEYCLOAK_URL}/realms/master/protocol/openid-connect/token" \
  -H "Content-Type: application/x-www-form-urlencoded" \
  -d "client_id=admin-cli" \
  -d "username=${KEYCLOAK_ADMIN}" \
  -d "password=${KEYCLOAK_ADMIN_PASSWORD}" \
  -d "grant_type=password" | grep -o '"access_token":"[^"]*' | cut -d'"' -f4)

if [ -z "$TOKEN" ]; then
  echo "ERROR: Failed to get admin token"
  exit 1
fi

echo "Creating realm: ${REALM}..."
curl -sf -X POST "${KEYCLOAK_URL}/admin/realms" \
  -H "Authorization: Bearer ${TOKEN}" \
  -H "Content-Type: application/json" \
  -d "{
    \"realm\": \"${REALM}\",
    \"enabled\": true,
    \"displayName\": \"AbyssCore\"
  }" || echo "Realm may already exist"

echo "Creating client: ${CLIENT_ID}..."
curl -sf -X POST "${KEYCLOAK_URL}/admin/realms/${REALM}/clients" \
  -H "Authorization: Bearer ${TOKEN}" \
  -H "Content-Type: application/json" \
  -d "{
    \"clientId\": \"${CLIENT_ID}\",
    \"publicClient\": true,
    \"redirectUris\": [\"http://localhost:3000/*\"],
    \"webOrigins\": [\"http://localhost:3000\"],
    \"standardFlowEnabled\": true,
    \"attributes\": {\"pkce.code.challenge.method\": \"S256\"}
  }" || echo "Client may already exist"

echo "Creating roles..."
for ROLE in player admin; do
  curl -sf -X POST "${KEYCLOAK_URL}/admin/realms/${REALM}/roles" \
    -H "Authorization: Bearer ${TOKEN}" \
    -H "Content-Type: application/json" \
    -d "{\"name\": \"${ROLE}\"}" || echo "Role ${ROLE} may already exist"
done

echo "Creating test user: ${TEST_USER}..."
curl -sf -X POST "${KEYCLOAK_URL}/admin/realms/${REALM}/users" \
  -H "Authorization: Bearer ${TOKEN}" \
  -H "Content-Type: application/json" \
  -d "{
    \"username\": \"${TEST_USER}\",
    \"email\": \"${TEST_USER}@abysscore.local\",
    \"enabled\": true,
    \"credentials\": [{
      \"type\": \"password\",
      \"value\": \"${TEST_PASSWORD}\",
      \"temporary\": false
    }]
  }" || echo "User may already exist"

echo "Keycloak init complete."
