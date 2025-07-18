#!/bin/bash
# Script to test the pricing endpoint with curl on Linux/macOS

# URL of the endpoint
ENDPOINT="http://localhost:8080/pricing"

# JSON payload for the request
PAYLOAD='{
    "loan_amount": 10000.0,
    "customer_id": "CUST12345"
}'

# Execute POST request with curl
curl -X POST "$ENDPOINT" \
     -H "Content-Type: application/json" \
     -d "$PAYLOAD" \
     --silent | jq .

# Check request status
if [ $? -eq 0 ]; then
    echo "Requisição enviada com sucesso!"
else
    echo "Erro ao enviar a requisição."
fi