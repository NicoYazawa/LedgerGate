param(
    [string]$Base = "origin/main",
    [string]$Out = "artifacts/pr-impact",
    [string]$Stage = "",
    [string]$CodeGraphRoot = "docs/codegraph",
    [string]$CodeGraphScripts = "",
    [int]$BlastRadius = 20,
    [int]$MaxDepth = 3,
    [switch]$AllowSignatures,
    [switch]$Enforce
)

$ErrorActionPreference = "Stop"
$repo = Split-Path $PSScriptRoot -Parent
$arguments = @(
    "run", "./cmd/pr-impact",
    "-repo", $repo,
    "-base", $Base,
    "-out", $Out,
    "-stage", $Stage,
    "-codegraph-root", $CodeGraphRoot,
    "-blast-radius", $BlastRadius,
    "-max-depth", $MaxDepth
)
if ($CodeGraphScripts) {
    $arguments += @("-codegraph-scripts", $CodeGraphScripts)
}
if ($AllowSignatures) {
    $arguments += "-allow-signatures"
}
if ($Enforce) {
    $arguments += "-enforce"
}

Push-Location $repo
try {
    & go @arguments
    exit $LASTEXITCODE
} finally {
    Pop-Location
}
