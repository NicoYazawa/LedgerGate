param(
    [int]$Pr = 0,
    [string]$Path = "artifacts/pr-impact/summary.md"
)

$ErrorActionPreference = "Stop"

function Write-Utf8NoBom {
    param([string]$FilePath, [string]$Text)
    $enc = New-Object System.Text.UTF8Encoding -ArgumentList $false
    [System.IO.File]::WriteAllText($FilePath, $Text, $enc)
}

if ($Pr -le 0) {
    Write-Output "No PR number; skipping impact comment."
    exit 0
}

if (-not (Test-Path $Path)) {
    Write-Output "No impact summary at $Path; skipping comment."
    exit 0
}

$body = Get-Content -Path $Path -Raw
if ($body -match '<!-- codegraph-impact:skip -->') {
    Write-Output "Impact gate skipped (no code files); skipping comment."
    exit 0
}

if ($body.Length -gt 60000) {
    $body = $body.Substring(0, 60000) + "`n`n_Comment truncated; download the `pr-impact` artifact for the full diagram._"
}

$tempPath = Join-Path ([System.IO.Path]::GetTempPath()) "ledgergate-impact-comment.md"
Write-Utf8NoBom $tempPath $body
$tempPathForGh = $tempPath -replace '\\', '/'

$repo = $env:GITHUB_REPOSITORY
if ([string]::IsNullOrWhiteSpace($repo)) {
    $repo = (gh repo view --json nameWithOwner --jq .nameWithOwner).Trim()
}

$marker = '<!-- codegraph-impact -->'
$ids = @(gh api "repos/$repo/issues/$Pr/comments" --paginate --jq '.[] | select(.body | contains("<!-- codegraph-impact -->")) | .id')
$ids = @($ids | Where-Object { -not [string]::IsNullOrWhiteSpace($_) })

if ($ids.Count -gt 0) {
    $id = $ids[0].Trim()
    gh api --method PATCH "repos/$repo/issues/comments/$id" -F "body=@$tempPathForGh" | Out-Null
    Write-Output "Updated impact comment $id on PR #$Pr."
} else {
    gh pr comment $Pr --repo $repo --body-file $tempPathForGh | Out-Null
    Write-Output "Created impact comment on PR #$Pr."
}
