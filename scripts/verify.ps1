$ErrorActionPreference = 'Stop'
$repository = Split-Path -Parent $PSScriptRoot
$previousToolchain = $env:GOTOOLCHAIN
$previousGOOS = $env:GOOS
function Check-Native {
    param([string]$Command, [string[]]$Arguments)
    & $Command @Arguments
    if ($LASTEXITCODE -ne 0) { throw "$Command failed with exit code $LASTEXITCODE" }
}
Push-Location $repository
try {
    $env:GOTOOLCHAIN = 'go1.24.0'
    Push-Location frontend
    try {
        Check-Native npx @('svelte-check', '--tsconfig', './tsconfig.json')
        Check-Native npm @('run', 'build')
    } finally { Pop-Location }
    Check-Native go @('vet', './...')
    Check-Native go @('test', './...')
    if ($env:OS -eq 'Windows_NT') {
        Check-Native go @('test', '-tags', 'integration', './...')
        Check-Native wails @('build')
    } else {
        $env:GOOS = 'windows'
        Check-Native go @('vet', './...')
        Write-Host 'Integration tests and exe build were not run (requires Windows).'
    }
} finally {
    $env:GOTOOLCHAIN = $previousToolchain
    $env:GOOS = $previousGOOS
    Pop-Location
}
