param(
    [string[]]$Source,
    [ValidateSet("svg", "png")]
    [string]$Format = "svg"
)

$ErrorActionPreference = "Stop"

function Find-Java {
    $java = Get-Command java.exe, java -ErrorAction SilentlyContinue | Select-Object -First 1
    if ($java) {
        return $java.Source
    }
    if ($env:JAVA_HOME) {
        $candidate = Join-Path $env:JAVA_HOME "bin/java.exe"
        if (Test-Path -LiteralPath $candidate) {
            return $candidate
        }
    }
    foreach ($root in @("HKLM:\SOFTWARE\JavaSoft\JDK", "HKLM:\SOFTWARE\JavaSoft\JRE")) {
        if (-not (Test-Path $root)) {
            continue
        }
        foreach ($key in Get-ChildItem $root) {
            $javaHome = (Get-ItemProperty $key.PSPath).JavaHome
            $candidate = Join-Path $javaHome "bin/java.exe"
            if (Test-Path -LiteralPath $candidate) {
                return $candidate
            }
        }
    }
    throw "Java runtime not found. Install a JRE or set JAVA_HOME before rendering PlantUML."
}

$java = Find-Java
$env:JAVA_HOME = Split-Path (Split-Path $java -Parent) -Parent
$env:PATH = "$(Split-Path $java -Parent);$env:PATH"

$command = $null
$commandPrefix = @()
$plantuml = Get-Command plantuml -ErrorAction SilentlyContinue | Select-Object -First 1
if ($plantuml) {
    $command = $plantuml.Source
} else {
    $npx = Get-Command npx -ErrorAction SilentlyContinue | Select-Object -First 1
    if ($npx) {
        & $npx.Source --no-install plantuml-cli --version *> $null
        if ($LASTEXITCODE -eq 0) {
            $command = $npx.Source
            $commandPrefix = @("--no-install", "plantuml-cli")
        }
    }
}
if (-not $command) {
    throw "PlantUML CLI not found. Install plantuml or run: npx install plantuml-cli"
}

$repositoryRoot = Split-Path $PSScriptRoot -Parent
if (-not $Source) {
    $Source = @(
        Get-ChildItem -LiteralPath (Join-Path $repositoryRoot "docs/diagrams") -Filter "*.puml" -File -Recurse |
            ForEach-Object FullName
    )
}
if ($Source.Count -eq 0) {
    throw "No PlantUML sources found."
}

foreach ($item in $Source) {
    $file = if ([System.IO.Path]::IsPathRooted($item)) {
        $item
    } else {
        Join-Path (Get-Location) $item
    }
    $file = [System.IO.Path]::GetFullPath($file)
    if (-not (Test-Path -LiteralPath $file)) {
        throw "PlantUML source not found: $file"
    }

    $arguments = @($commandPrefix) + @("--check-before-run", "--no-error-image", "-t$Format", $file)
    & $command @arguments
    if ($LASTEXITCODE -ne 0) {
        throw "PlantUML failed for $file"
    }

    $output = [System.IO.Path]::ChangeExtension($file, $Format)
    if (-not (Test-Path -LiteralPath $output)) {
        throw "PlantUML did not create $output"
    }
    Write-Output "Rendered $output"
}
