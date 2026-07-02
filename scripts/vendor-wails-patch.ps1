# Regenerates vendor/ from go.mod/go.sum and (re)applies the local Wails v2
# patches under patches/*.patch. Run this after any change to the pinned
# wailsapp/wails/v2 version, or any time vendor/ is missing/stale.
#
# Usage: powershell -File scripts/vendor-wails-patch.ps1

$ErrorActionPreference = "Stop"
$root = Split-Path -Parent $PSScriptRoot
Set-Location $root

Write-Host "Vendoring dependencies..."
go mod vendor
if ($LASTEXITCODE -ne 0) { throw "go mod vendor failed" }

Get-ChildItem -Path "patches" -Filter "*.patch" | Sort-Object Name | ForEach-Object {
    $patch = $_.FullName
    Write-Host "Checking $($_.Name)..."

    & git apply --check $patch 2>$null
    if ($LASTEXITCODE -eq 0) {
        Write-Host "Applying $($_.Name)..."
        & git apply $patch
        if ($LASTEXITCODE -ne 0) { throw "Failed to apply $($_.Name)" }
    } else {
        # Already applied (or vendor content changed upstream) - verify it's
        # actually in place rather than silently ignoring a real conflict.
        & git apply --reverse --check $patch 2>$null
        if ($LASTEXITCODE -eq 0) {
            Write-Host "$($_.Name) already applied, skipping."
        } else {
            throw "$($_.Name) does not apply and is not already applied - Wails source likely changed upstream. Re-diff the patch by hand."
        }
    }
}

Write-Host "Done."
