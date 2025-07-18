@echo off
REM Script to test the pricing endpoint with curl on Windows

REM URL of the endpoint
set ENDPOINT=http://localhost:8080/pricing

REM JSON payload for the request
set PAYLOAD={"loan_amount": 10000.0, "customer_id": "CUST12345"}

REM Execute POST request with curl
curl -X POST "%ENDPOINT%" -H "Content-Type: application/json" -d "%PAYLOAD%" --silent

REM Check request status
if %ERRORLEVEL% equ 0 (
    echo.
    echo Requisição enviada com sucesso!
) else (
    echo.
    echo Erro ao enviar a requisição.
)