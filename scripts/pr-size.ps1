param(
    [string]$Base = "origin/main",
    [switch]$Enforce,
    [int]$Top = 10
)

$ErrorActionPreference = "Stop"

function Resolve-Base {
    param([string]$Ref)
    $sha = git rev-parse --verify --quiet $Ref
    if ($LASTEXITCODE -ne 0 -or [string]::IsNullOrWhiteSpace($sha)) {
        return $null
    }
    return $sha.Trim()
}

function Get-Tier {
    param([string]$Path)

    if ($Path -match '^(internal/(ledger|billing|payment)/|internal/platform/(auth|crypt)/|migrations/(ledger|payment|control)/|api/openapi/(ledger|payment|control)\.yaml)') {
        return [pscustomobject]@{ Name = "money/auth/data"; Soft = 300; Hard = 450 }
    }
    if ($Path -match '^(internal/|migrations/|api/openapi/|cmd/)') {
        return [pscustomobject]@{ Name = "core"; Soft = 400; Hard = 600 }
    }
    if ($Path -match '^(go\.mod|go\.sum|Makefile|tools\.go|\.golangci\.yml|\.env\.example|docker-compose\.yml|deploy/)') {
        return [pscustomobject]@{ Name = "core"; Soft = 400; Hard = 600 }
    }
    if ($Path -match '^web/') {
        return [pscustomobject]@{ Name = "ui"; Soft = 800; Hard = 1200 }
    }
    if ($Path -match '^(docs/|scripts/|\.github/|README\.md$|AGENTS\.md$|\.scratch/|\.gitignore$)') {
        return [pscustomobject]@{ Name = "docs/tooling"; Soft = 800; Hard = 1200 }
    }
    return [pscustomobject]@{ Name = "default"; Soft = 500; Hard = 750 }
}

function Test-Excluded {
    param([string]$Path)
    if ($Path -match '^(vendor|node_modules|bin|dist)/') { return $true }
    if ($Path -match '(^|/)api/generated/') { return $true }
    if ($Path -match '\.gen\.go$') { return $true }
    if ($Path -match '^coverage') { return $true }
    if ($Path -in @('go.sum', 'package-lock.json', 'pnpm-lock.yaml', 'yarn.lock')) { return $true }
    return $false
}

function Test-TestFile {
    param([string]$Path)
    if ($Path -match '^internal/testkit/') { return $true }
    if ($Path -match '(^|/)(test|testdata)/') { return $true }
    if ($Path -match '(_test\.go|\.test\.(ts|tsx)|\.spec\.(ts|tsx))$') { return $true }
    return $false
}

$resolvedBase = Resolve-Base $Base
if (-not $resolvedBase -and $Base -eq 'origin/main') {
    $resolvedBase = Resolve-Base 'main'
}
if (-not $resolvedBase) {
    throw "Cannot resolve base '$Base' or fallback 'main'."
}

$mergeBase = git merge-base $resolvedBase HEAD
if ($LASTEXITCODE -ne 0 -or [string]::IsNullOrWhiteSpace($mergeBase)) {
    throw "Cannot find merge-base between '$Base' and HEAD."
}
$mergeBase = $mergeBase.Trim()

$numstat = git diff --numstat --no-renames $mergeBase HEAD
$rows = @()
$testChurn = 0
$excludedChurn = 0

foreach ($line in $numstat) {
    if ([string]::IsNullOrWhiteSpace($line)) { continue }
    $parts = $line -split "`t"
    if ($parts.Count -lt 3) { continue }

    $added = 0
    $deleted = 0
    if ($parts[0] -ne '-') { $added = [int]$parts[0] }
    if ($parts[1] -ne '-') { $deleted = [int]$parts[1] }
    $churn = $added + $deleted
    $path = $parts[2] -replace '\\', '/'

    if (Test-Excluded $path) {
        $excludedChurn += $churn
        continue
    }
    if (Test-TestFile $path) {
        $testChurn += $churn
        continue
    }

    $tier = Get-Tier $path
    $rows += [pscustomobject]@{
        File    = $path
        Added   = $added
        Deleted = $deleted
        Churn   = $churn
        Tier    = $tier.Name
        Soft    = $tier.Soft
        Hard    = $tier.Hard
    }
}

$productionChurn = 0
foreach ($row in $rows) { $productionChurn += $row.Churn }

$tiers = @($rows | Group-Object Tier | ForEach-Object {
    $first = $_.Group[0]
    [pscustomobject]@{ Name = $_.Name; Soft = $first.Soft; Hard = $first.Hard; Count = $_.Count }
} | Sort-Object Soft)

$activeTier = $null
if ($tiers.Count -gt 0) { $activeTier = $tiers[0] }
$soft = 500
$hard = 750
if ($activeTier) { $soft = $activeTier.Soft; $hard = $activeTier.Hard }

$status = "PASS"
$exitCode = 0
if ($productionChurn -gt $hard) {
    $status = "FAIL (hard limit)"
    $exitCode = 1
} elseif ($productionChurn -gt $soft) {
    if ($Enforce) {
        $status = "FAIL (soft limit, -Enforce)"
        $exitCode = 1
    } else {
        $status = "WARN (soft limit)"
    }
}

Write-Output "Base:            $Base ($($mergeBase.Substring(0, 7)))"
if ($activeTier) {
    Write-Output "Active tier:     $($activeTier.Name) (soft $soft / hard $hard)"
} else {
    Write-Output "Active tier:     none - no production files changed"
}
Write-Output "Production:      $productionChurn lines"
Write-Output "Tests:           $testChurn lines (excluded)"
Write-Output "Generated/vendor/lockfiles: $excludedChurn lines (excluded)"
Write-Output "Status:          $status"
Write-Output ""

if ($rows.Count -gt 0) {
    Write-Output "Top files by churn:"
    $rows | Sort-Object Churn -Descending | Select-Object -First $Top |
        Format-Table File, Added, Deleted, Churn, Tier -AutoSize
}

if ($rows.Count -eq 0) {
    Write-Output "No hand-written production changes."
}

exit $exitCode
