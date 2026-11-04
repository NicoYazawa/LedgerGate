param(
    [Parameter(Mandatory = $true)]
    [string]$Name,

    [int]$Offset = 0,

    [string]$Path = ".env.local"
)

$ErrorActionPreference = "Stop"

$ports = [ordered]@{
    GATEWAY_PORT        = 8080
    CONTROL_PORT        = 8081
    LEDGER_PORT         = 8082
    BILLING_PORT        = 8083
    PAYMENT_PORT        = 8084
    RECONCILE_PORT      = 8085
    POSTGRES_PORT       = 5432
    REDIS_PORT          = 6379
    NATS_PORT           = 4222
    NATS_MONITOR_PORT   = 8222
    JAEGER_UI_PORT      = 16686
    OTEL_GRPC_PORT      = 4317
    OTEL_HTTP_PORT      = 4318
    PROMETHEUS_PORT     = 9090
    GRAFANA_PORT        = 3000
}

$lines = @("COMPOSE_PROJECT_NAME=$Name")

foreach ($key in $ports.Keys) {
    $value = $ports[$key] + $Offset
    if ($value -lt 1 -or $value -gt 65535) {
        throw "Port out of range for $key`: $value"
    }
    $lines += "$key=$value"
}

$lines | Set-Content -Path $Path -Encoding utf8

Write-Output "Wrote $Path"
Write-Output "  COMPOSE_PROJECT_NAME=$Name"
Write-Output "  host port offset=$Offset"
