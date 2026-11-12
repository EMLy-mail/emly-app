<#
.SYNOPSIS
    Captures two heap profiles from a running EMLy debug build and opens the diff.

.DESCRIPTION
    1. Downloads a heap profile from the live pprof endpoint (after a GC).
    2. Waits -Seconds seconds: use the app in the meantime (open/close mails,
       attachments, PDFs...).
    3. Downloads a second heap profile.
    4. Opens `go tool pprof -diff_base` in the browser, showing only the
       memory that grew between the two snapshots.

    Requires EMLy built with -tags debug (endpoint on 127.0.0.1:6060).

.EXAMPLE
    .\scripts\heap-diff.ps1
    .\scripts\heap-diff.ps1 -Seconds 60 -SampleIndex alloc_space
#>
param(
    # Seconds to wait between the two snapshots
    [int]$Seconds = 20,
    # Address of the live pprof endpoint
    [string]$Endpoint = "http://127.0.0.1:6060",
    # Port for the pprof web UI
    [int]$UiPort = 8081,
    # inuse_space (memory still held) or alloc_space (memory allocated in between)
    [ValidateSet("inuse_space", "inuse_objects", "alloc_space", "alloc_objects")]
    [string]$SampleIndex = "inuse_space",
    # Folder where the .prof files are saved
    [string]$OutDir = (Join-Path $PSScriptRoot "..\profiles")
)

$ErrorActionPreference = "Stop"

$heapUrl = "$Endpoint/debug/pprof/heap?gc=1"

# Check the endpoint is reachable before waiting
try {
    Invoke-WebRequest -Uri "$Endpoint/debug/pprof/" -UseBasicParsing -TimeoutSec 5 | Out-Null
} catch {
    Write-Error "pprof endpoint not reachable at $Endpoint. Is EMLy running as a debug build (-tags debug)?"
    exit 1
}

New-Item -ItemType Directory -Force -Path $OutDir | Out-Null
$OutDir = (Resolve-Path $OutDir).Path
$stamp = Get-Date -Format "yyyyMMdd-HHmmss"
$heap1 = Join-Path $OutDir "heap-$stamp-1.prof"
$heap2 = Join-Path $OutDir "heap-$stamp-2.prof"

Write-Host "Snapshot 1 -> $heap1"
Invoke-WebRequest -Uri $heapUrl -OutFile $heap1 -UseBasicParsing

Write-Host "Use the app now. Second snapshot in $Seconds seconds..."
for ($i = $Seconds; $i -gt 0; $i--) {
    Write-Progress -Activity "Waiting before snapshot 2" -Status "$i s left" `
        -PercentComplete ((($Seconds - $i) / $Seconds) * 100)
    Start-Sleep -Seconds 1
}
Write-Progress -Activity "Waiting before snapshot 2" -Completed

Write-Host "Snapshot 2 -> $heap2"
Invoke-WebRequest -Uri $heapUrl -OutFile $heap2 -UseBasicParsing

Write-Host "Opening diff ($SampleIndex) at http://localhost:$UiPort (Ctrl+C to stop)"
go tool pprof "-http=:$UiPort" "-sample_index=$SampleIndex" "-diff_base=$heap1" $heap2
