param(
    [string]$Base = "origin/main",
    [string]$Out = "artifacts/pr-impact",
    [int]$BlastRadius = 20,
    [switch]$Enforce,
    [switch]$AllowSignatures,
    [int]$MaxMermaidLines = 300
)

$ErrorActionPreference = "Stop"

$utf8 = New-Object System.Text.UTF8Encoding -ArgumentList $false
[Console]::OutputEncoding = $utf8
$OutputEncoding = $utf8

function Write-Utf8NoBom {
    param([string]$Path, [string]$Text)
    $enc = New-Object System.Text.UTF8Encoding -ArgumentList $false
    [System.IO.File]::WriteAllText($Path, $Text, $enc)
}

function Resolve-BaseRef {
    param([string]$Ref)
    $sha = git rev-parse --verify --quiet $Ref
    if ($LASTEXITCODE -eq 0 -and -not [string]::IsNullOrWhiteSpace($sha)) {
        return ($sha | Select-Object -First 1).Trim()
    }
    if ($Ref -like 'origin/*') {
        $branch = $Ref.Substring(7)
        git fetch origin $branch | Out-Null
        $sha = git rev-parse --verify --quiet $Ref
        if ($LASTEXITCODE -eq 0 -and -not [string]::IsNullOrWhiteSpace($sha)) {
            return ($sha | Select-Object -First 1).Trim()
        }
    }
    throw "Cannot resolve base ref '$Ref'. Fetch it first, for example: git fetch origin main"
}

function Clean-Logs {
    param([string]$Text)
    $lines = $Text -split "`n" | ForEach-Object { $_.TrimEnd("`r") }
    $kept = $lines | Where-Object { $_ -notmatch '^\[codegraph\]' }
    return (($kept -join "`n").Trim())
}

function Limit-Text {
    param([string]$Text, [int]$MaxLines)
    $lines = $Text -split "`n" | ForEach-Object { $_.TrimEnd("`r") }
    if ($lines.Count -le $MaxLines) { return $Text }
    $head = $lines | Select-Object -First $MaxLines
    return (($head -join "`n") + "`n... (truncated)")
}

New-Item -ItemType Directory -Force -Path $Out | Out-Null
$summaryPath = Join-Path $Out "summary.md"

try {
    $baseSha = Resolve-BaseRef $Base
} catch {
    $message = @(
        '<!-- codegraph-impact -->',
        '',
        '## Impact gate: ERROR',
        '',
        $_.Exception.Message
    ) -join "`n"
    Write-Utf8NoBom $summaryPath $message
    Write-Output $message
    if ($Enforce) { exit 1 } else { exit 0 }
}

$headSha = (git rev-parse --short HEAD).Trim()
$changed = @(git diff --name-only "$baseSha...HEAD")
$codeChanged = @($changed | Where-Object { $_ -match '\.(go|ts|tsx|js|jsx|py|rb|java|kt|rs|c|cc|cpp|h|hpp|cs)$' })

if ($codeChanged.Count -eq 0) {
    $message = @(
        '<!-- codegraph-impact -->',
        '<!-- codegraph-impact:skip -->',
        '',
        '## Impact gate',
        '',
        'No code files changed in this PR. The codegraph impact gate was skipped.',
        '',
        "- Base: ``$Base`` ($($baseSha.Substring(0, 7)))",
        "- Head: ``$headSha``",
        "- Changed files: $($changed.Count)",
        '',
        'Docs, SQL and OpenAPI changes are still covered by the risk tiers, the reviewer and the tests.'
    ) -join "`n"
    Write-Utf8NoBom $summaryPath $message
    Write-Output "No code files changed; impact gate skipped."
    exit 0
}

if (-not (Get-Command codegraph -ErrorAction SilentlyContinue)) {
    $message = @(
        '<!-- codegraph-impact -->',
        '',
        '## Impact gate: ERROR',
        '',
        '`codegraph` is not installed. Install `@optave/codegraph@3.17.0` and rerun.'
    ) -join "`n"
    Write-Utf8NoBom $summaryPath $message
    Write-Output $message
    if ($Enforce) { exit 1 } else { exit 0 }
}

$buildOut = & codegraph build . | Out-String
$buildExit = $LASTEXITCODE
Write-Utf8NoBom (Join-Path $Out "build.log") (Clean-Logs $buildOut)

$structureOut = & codegraph branch-compare $baseSha HEAD --format mermaid --no-tests | Out-String
$structureExit = $LASTEXITCODE
$structure = Clean-Logs $structureOut
Write-Utf8NoBom (Join-Path $Out "structure.mmd") $structure

$impactOut = & codegraph diff-impact $baseSha --format mermaid --no-tests --depth 3 -n 200 | Out-String
$impactExit = $LASTEXITCODE
$impact = Clean-Logs $impactOut
Write-Utf8NoBom (Join-Path $Out "impact.mmd") $impact

$checkArgs = @('check', $baseSha, '--cycles', '--boundaries', '--blast-radius', $BlastRadius, '-j')
if (-not $AllowSignatures) {
    $checkArgs += '--signatures'
}
$checkOut = & codegraph @checkArgs | Out-String
$checkExit = $LASTEXITCODE
$checkText = Clean-Logs $checkOut
Write-Utf8NoBom (Join-Path $Out "check.json") $checkText

$rulesOut = & codegraph check | Out-String
$rulesExit = $LASTEXITCODE
$rulesText = Clean-Logs $rulesOut
Write-Utf8NoBom (Join-Path $Out "rules.txt") $rulesText

$predicates = @()
try {
    $checkJson = $checkText | ConvertFrom-Json
    $predicates = @($checkJson.predicates)
} catch {
    $predicates = @()
}

$gateFailed = ($buildExit -ne 0) -or ($checkExit -ne 0)
$rulesFailed = ($rulesExit -ne 0)
$artifactFailed = ($structureExit -ne 0) -or ($impactExit -ne 0)

if ($gateFailed) {
    if ($Enforce) { $status = "FAILED" } else { $status = "WARN (advisory)" }
} elseif ($rulesFailed) {
    $status = "WARN (rules advisory)"
} elseif ($artifactFailed) {
    $status = "WARN (artifact)"
} else {
    $status = "PASS"
}

$lines = @()
$lines += '<!-- codegraph-impact -->'
$lines += ''
$lines += "## Impact gate: $status"
$lines += ''
$lines += "| Field | Value |"
$lines += "|---|---|"
$lines += "| Base | ``$Base`` ($($baseSha.Substring(0, 7))) |"
$lines += "| Head | ``$headSha`` |"
$lines += "| Changed files | $($changed.Count) |"
$lines += "| Code files changed | $($codeChanged.Count) |"
$lines += "| Blast-radius threshold | $BlastRadius |"
$lines += "| Signatures allowed | $([bool]$AllowSignatures) |"
$lines += "| Enforced | $([bool]$Enforce) |"
$lines += ''
$lines += '### Structure diff'
$lines += ''
if ([string]::IsNullOrWhiteSpace($structure)) {
    $lines += '_No structural changes reported._'
} else {
    $lines += '```mermaid'
    $lines += (Limit-Text $structure $MaxMermaidLines)
    $lines += '```'
}
$lines += ''
$lines += '### Impact'
$lines += ''
if ([string]::IsNullOrWhiteSpace($impact)) {
    $lines += '_No impact nodes reported._'
} else {
    $lines += '```mermaid'
    $lines += (Limit-Text $impact $MaxMermaidLines)
    $lines += '```'
}
$lines += ''
$lines += '### Gate predicates'
$lines += ''
if ($predicates.Count -eq 0) {
    $lines += '_Predicate results unavailable; check `check.json`._'
} else {
    $lines += '| Predicate | Result |'
    $lines += '|---|---|'
    foreach ($p in $predicates) {
        $result = 'FAIL'
        if ($p.passed) { $result = 'PASS' }
        $lines += "| $($p.name) | $result |"
    }
}
$lines += ''
$lines += '<details><summary>Manifesto rules</summary>'
$lines += ''
if ([string]::IsNullOrWhiteSpace($rulesText)) {
    $lines += '_No rules output._'
} else {
    $lines += '```text'
    $lines += (Limit-Text $rulesText 80)
    $lines += '```'
}
$lines += ''
$lines += '</details>'
$lines += ''
$lines += '### Notes'
$lines += ''
$lines += '- codegraph sees code dependencies, not money invariants, SQL semantics, OpenAPI compatibility or runtime behavior.'
$lines += '- A green gate does not replace tests, the risk tiers, or the independent review.'
$lines += '- Manifesto rules are advisory until the thresholds are tuned for this repo; predicate failures are the hard gate.'
$lines += '- `mermaidiff` is a reviewer aid, not part of this gate.'
$lines += '- Set the repository variable `CODEGRAPH_ENFORCE=true` to turn advisory failures into a required check.'

Write-Utf8NoBom $summaryPath ($lines -join "`n")

Write-Output "Impact gate: $status"
Write-Output "Summary: $summaryPath"
Write-Output "Artifacts: $Out"

if ($gateFailed -and $Enforce) { exit 1 }
exit 0
