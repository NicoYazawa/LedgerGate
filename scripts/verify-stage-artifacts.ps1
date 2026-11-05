param(
    [Parameter(Mandatory = $true)]
    [string]$Stage
)
$ErrorActionPreference = "Stop"
$repositoryRoot = Split-Path $PSScriptRoot -Parent
$missing = @()
$plantUML = @(
    Get-ChildItem -LiteralPath (Join-Path $repositoryRoot "docs/diagrams/stages") -Filter "*-$Stage.puml" -File -ErrorAction SilentlyContinue
)
$plantUMLRender = @(
    Get-ChildItem -LiteralPath (Join-Path $repositoryRoot "docs/diagrams/stages") -Filter "*-$Stage.svg" -File -ErrorAction SilentlyContinue
)
if ($plantUML.Count -ne 1) { $missing += "docs/diagrams/stages/*-$Stage.puml" }
if ($plantUMLRender.Count -ne 1) { $missing += "docs/diagrams/stages/*-$Stage.svg" }
foreach ($path in @(
    "docs/codegraph/$Stage/graph.mmd",
    "docs/codegraph/$Stage/graph.locs.json",
    "docs/codegraph/$Stage/graph.html",
    "docs/reviews/$Stage/mermaidiff.md",
    "docs/reviews/$Stage/mermaidiff.html",
    "docs/reviews/$Stage/README.md"
)) {
    $fullPath = Join-Path $repositoryRoot $path
    if (-not (Test-Path -LiteralPath $fullPath -PathType Leaf) -or (Get-Item -LiteralPath $fullPath).Length -eq 0) {
        $missing += $path
    }
}
if ($missing.Count -gt 0) {
    throw "Missing stage artifacts for '$Stage':`n$($missing -join "`n")"
}
Write-Output "Stage artifacts verified for '$Stage'."
