param(
    [string]$InstallerVersion = "0.1.0.0",
    [switch]$SkipBundleBuild
)

$ErrorActionPreference = "Stop"

$scriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
$repoRoot = Split-Path -Parent (Split-Path -Parent $scriptDir)
$payloadDir = Join-Path $scriptDir "payload"
$bootstrapperPayloadDir = Join-Path $payloadDir "ba"
$distDir = Join-Path $scriptDir "dist"
# Burn expects the staged engine payload at: payload\active-stack.exe
$payloadExe = Join-Path $payloadDir "active-stack.exe"
$bootstrapperProject = Join-Path $scriptDir "ui\ActiveStack.Bootstrapper.Host\ActiveStack.Bootstrapper.Host.csproj"
$bootstrapperExe = Join-Path $bootstrapperPayloadDir "ActiveStack.Bootstrapper.Host.exe"
$mbaNativeDll = Join-Path $bootstrapperPayloadDir "mbanative.dll"
$projectFile = Join-Path $scriptDir "ActiveStack.Bundle.wixproj"

New-Item -ItemType Directory -Force -Path $payloadDir | Out-Null
New-Item -ItemType Directory -Force -Path $bootstrapperPayloadDir | Out-Null
New-Item -ItemType Directory -Force -Path $distDir | Out-Null

Push-Location $repoRoot
try {
    go build -trimpath -o $payloadExe .\cmd\active-stack
    if ($LASTEXITCODE -ne 0) {
        throw "Go payload build failed."
    }
}
finally {
    Pop-Location
}

Push-Location $scriptDir
try {
    dotnet publish $bootstrapperProject -c Release -r win-x64 --self-contained true `
        -p:PublishSingleFile=true `
        -p:IncludeNativeLibrariesForSelfExtract=true `
        -o $bootstrapperPayloadDir
    if ($LASTEXITCODE -ne 0) {
        throw "Bootstrapper application publish failed."
    }
}
finally {
    Pop-Location
}

$nugetRoot = Join-Path $env:USERPROFILE ".nuget\packages"
$mbaNativeSource = Join-Path $nugetRoot "wixtoolset.bootstrapperapplicationapi\6.0.2\runtimes\win-x64\native\mbanative.dll"
if (-not (Test-Path $mbaNativeSource)) {
    throw "Required WiX native BA support file not found at $mbaNativeSource"
}

Copy-Item -LiteralPath $mbaNativeSource -Destination $mbaNativeDll -Force

Copy-Item -LiteralPath $payloadExe -Destination (Join-Path $bootstrapperPayloadDir "active-stack.exe") -Force

if ($SkipBundleBuild) {
    Write-Host "Payload staged at $payloadExe"
    exit 0
}

$installedSdks = & dotnet --list-sdks 2>$null
if ($LASTEXITCODE -ne 0 -or -not $installedSdks -or -not ($installedSdks | Select-Object -First 1)) {
    throw "Building the Windows bundle requires a .NET SDK. Install the .NET 8 SDK and run the script again."
}

Push-Location $scriptDir
try {
    dotnet build $projectFile -c Release -p:InstallerVersion=$InstallerVersion
    if ($LASTEXITCODE -ne 0) {
        throw "WiX bundle build failed."
    }
}
finally {
    Pop-Location
}

$builtBundle = Get-ChildItem -Path (Join-Path $scriptDir "bin") -Recurse -Filter *.exe |
    Where-Object { $_.FullName -ne $payloadExe } |
    Sort-Object LastWriteTimeUtc -Descending |
    Select-Object -First 1

if (-not $builtBundle) {
    throw "Bundle output not found under $scriptDir\\bin"
}

$releaseName = "ActiveStack-Setup-$InstallerVersion.exe"
$releasePath = Join-Path $distDir $releaseName
Copy-Item -LiteralPath $builtBundle.FullName -Destination $releasePath -Force

Write-Host "Bundle ready at $releasePath"
