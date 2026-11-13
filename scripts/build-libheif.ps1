<#
.SYNOPSIS
    Builds libheif.dll (with the libde265 HEVC decoder inside) for EMLy.

.DESCRIPTION
    Downloads pinned libde265 and libheif source releases and builds, with
    the MinGW-w64 gcc on PATH:
      - libde265 as a static library, linked into
      - libheif as a self-contained shared library: libheif.dll, with
        libgcc/libstdc++/winpthread linked statically, so it depends on
        system DLLs only.

    EMLy loads libheif.dll at runtime from its own folder (see
    backend/utils/heif): if the DLL is missing EMLy still starts and falls
    back to the in-browser libheif-js decoder. Keeping libheif (LGPL-3.0) in
    a separate, replaceable DLL is what lets EMLy comply with the LGPL
    without publishing EMLy's own code.

    Outputs:
      third_party/libheif/bin/libheif.dll   the DLL to ship (installer, CI)
      third_party/libheif/licenses/         LGPL texts + NOTICE.txt to ship
      build/bin/libheif.dll                 copy for `wails dev` / local runs

    Only the HEVC decoder is enabled: EMLy only needs to *display* HEIC
    photos. Every encoder, AVIF codec and plugin loading is disabled.

    Requirements on PATH: gcc/g++ (MinGW-w64), cmake >= 3.16, ninja, tar,
    objdump. e.g. `scoop install mingw cmake ninja`

.EXAMPLE
    .\scripts\build-libheif.ps1
    .\scripts\build-libheif.ps1 -Clean
#>
[CmdletBinding()]
param(
    [string]$LibheifVersion = "1.23.7",
    [string]$Libde265Version = "1.1.3",
    [switch]$Clean
)

$ErrorActionPreference = "Stop"

$root = Split-Path -Parent $PSScriptRoot
$thirdParty = Join-Path $root "third_party"
$prefix = Join-Path $thirdParty "libheif"
$work = Join-Path $thirdParty ".build"
$depsPrefix = Join-Path $work "deps"   # static libde265, not shipped

foreach ($tool in "gcc", "g++", "cmake", "ninja", "tar", "objdump") {
    if (-not (Get-Command $tool -ErrorAction SilentlyContinue)) {
        throw "Required tool '$tool' not found on PATH (try: scoop install mingw cmake ninja)"
    }
}

if ($Clean) {
    foreach ($dir in $work, $prefix) {
        if (Test-Path $dir) { Remove-Item -Recurse -Force $dir -Confirm:$false }
    }
}
New-Item -ItemType Directory -Force $work, $prefix, $depsPrefix | Out-Null

function Get-Source([string]$name, [string]$url) {
    $archive = Join-Path $work "$name.tar.gz"
    $srcDir = Join-Path $work $name
    if (-not (Test-Path $srcDir)) {
        Write-Host "==> Downloading $url"
        Invoke-WebRequest -Uri $url -OutFile $archive
        tar -xzf $archive -C $work
        if ($LASTEXITCODE -ne 0) { throw "tar failed for $archive" }
    }
    return $srcDir
}

function Invoke-CMakeBuild([string]$name, [string]$srcDir, [string]$installPrefix, [string[]]$options) {
    $buildDir = Join-Path $work "$name-build"
    Write-Host "==> Configuring $name"
    cmake -S $srcDir -B $buildDir -G Ninja `
        -DCMAKE_BUILD_TYPE=Release `
        -DCMAKE_C_COMPILER=gcc `
        -DCMAKE_CXX_COMPILER=g++ `
        -DCMAKE_INSTALL_PREFIX="$installPrefix" `
        -DCMAKE_PREFIX_PATH="$depsPrefix" `
        @options
    if ($LASTEXITCODE -ne 0) { throw "cmake configure failed for $name" }

    Write-Host "==> Building $name"
    cmake --build $buildDir --parallel
    if ($LASTEXITCODE -ne 0) { throw "cmake build failed for $name" }

    cmake --install $buildDir --strip
    if ($LASTEXITCODE -ne 0) { throw "cmake install failed for $name" }
}

# --- libde265 (HEVC decoder), static, linked into libheif.dll ----------------
$de265Src = Get-Source "libde265-$Libde265Version" `
    "https://github.com/strukturag/libde265/releases/download/v$Libde265Version/libde265-$Libde265Version.tar.gz"
Invoke-CMakeBuild "libde265-static" $de265Src $depsPrefix @(
    "-DBUILD_SHARED_LIBS=OFF",
    "-DENABLE_SDL=OFF",
    "-DENABLE_DECODER=OFF",   # dec265 sample app
    "-DENABLE_ENCODER=OFF"
)

# --- libheif, shared --------------------------------------------------------
$heifSrc = Get-Source "libheif-$LibheifVersion" `
    "https://github.com/strukturag/libheif/releases/download/v$LibheifVersion/libheif-$LibheifVersion.tar.gz"
Invoke-CMakeBuild "libheif-shared" $heifSrc $prefix @(
    "-DBUILD_SHARED_LIBS=ON",
    "-DCMAKE_C_FLAGS=-DLIBDE265_STATIC_BUILD",
    "-DCMAKE_CXX_FLAGS=-DLIBDE265_STATIC_BUILD",
    # Self-contained DLL: no libgcc_s/libstdc++/libwinpthread DLLs to ship.
    "-DCMAKE_SHARED_LINKER_FLAGS=-static",
    "-DWITH_LIBDE265=ON",
    "-DWITH_LIBDE265_PLUGIN=OFF",
    "-DENABLE_PLUGIN_LOADING=OFF",
    "-DWITH_X265=OFF",
    "-DWITH_X264=OFF",
    "-DWITH_AOM_DECODER=OFF",
    "-DWITH_AOM_ENCODER=OFF",
    "-DWITH_DAV1D=OFF",
    "-DWITH_RAV1E=OFF",
    "-DWITH_SvtEnc=OFF",
    "-DWITH_KVAZAAR=OFF",
    "-DWITH_UVG266=OFF",
    "-DWITH_VVDEC=OFF",
    "-DWITH_VVENC=OFF",
    "-DWITH_FFMPEG_DECODER=OFF",
    "-DWITH_JPEG_DECODER=OFF",
    "-DWITH_JPEG_ENCODER=OFF",
    "-DWITH_OpenJPEG_DECODER=OFF",
    "-DWITH_OpenJPEG_ENCODER=OFF",
    "-DWITH_OPENJPH_ENCODER=OFF",
    "-DWITH_OpenH264_DECODER=OFF",
    "-DWITH_OpenH264_ENCODER=OFF",
    "-DWITH_LIBSHARPYUV=OFF",
    "-DWITH_HEADER_COMPRESSION=OFF",
    "-DWITH_UNCOMPRESSED_CODEC=OFF",
    "-DWITH_GDK_PIXBUF=OFF",
    "-DWITH_EXAMPLES=OFF",
    "-DBUILD_TESTING=OFF",
    "-DENABLE_EXPERIMENTAL_FEATURES=OFF"
)

$dll = Join-Path $prefix "bin\libheif.dll"
if (-not (Test-Path $dll)) { throw "libheif.dll not produced at $dll" }

# --- verify: the DLL must import system DLLs only ---------------------------
$imports = objdump -p $dll | Select-String "DLL Name:" | ForEach-Object { ($_ -split ":\s*")[1].Trim() }
$foreign = $imports | Where-Object { $_ -notmatch '^(KERNEL32|USER32|ADVAPI32|msvcrt|ucrtbase|api-ms-win-.*)\.dll$' }
if ($foreign) {
    throw "libheif.dll depends on non-system DLLs: $($foreign -join ', ')"
}

# --- licenses (LGPL-3.0: ship texts + where to get the sources) -------------
$licenses = Join-Path $prefix "licenses"
New-Item -ItemType Directory -Force $licenses | Out-Null
Copy-Item (Join-Path $heifSrc "COPYING") (Join-Path $licenses "libheif-COPYING.txt") -Force
Copy-Item (Join-Path $de265Src "COPYING") (Join-Path $licenses "libde265-COPYING.txt") -Force
@"
EMLy uses the following libraries, distributed in libheif.dll:

  libheif $LibheifVersion   https://github.com/strukturag/libheif   LGPL-3.0
  libde265 $Libde265Version   https://github.com/strukturag/libde265  LGPL-3.0

Their full license texts are in this folder. libheif.dll was built
unmodified from the release sources above (libde265 linked statically into
it) with EMLy's scripts/build-libheif.ps1. You may replace libheif.dll with
your own build of libheif exposing the same C API; EMLy loads it at runtime
from its installation folder.
"@ | Set-Content -Encoding UTF8 (Join-Path $licenses "NOTICE.txt")

# --- copy next to the dev/build executables ---------------------------------
$binDir = Join-Path $root "build\bin"
New-Item -ItemType Directory -Force $binDir | Out-Null
Copy-Item $dll $binDir -Force

Write-Host ""
Write-Host "==> libheif $LibheifVersion + libde265 $Libde265Version"
Write-Host "    DLL:      $dll ($([math]::Round((Get-Item $dll).Length / 1MB, 1)) MB)"
Write-Host "    imports:  $($imports -join ', ')"
Write-Host "    copied to $binDir"
