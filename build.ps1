param (
    [switch]$Release = $false,
    [switch]$NoConsole = $false
)

# Set environment variables
$env:GOOS = "windows"
$env:GOARCH = "amd64"

# Build flags
$ldflags = ""
if ($NoConsole) {
    $ldflags = "-H windowsgui"
}

$buildFlags = @("-ldflags", $ldflags)

if ($Release) {
    Write-Host "Building release version..."
    if ($NoConsole) {
        $buildFlags = @("-ldflags", "-s -w -H windowsgui")
    } else {
        $buildFlags = @("-ldflags", "-s -w")
    }
}

# Clean old builds
if (Test-Path "plate.exe") {
    Remove-Item "plate.exe"
}

# Build the application
Write-Host "Building application..."
go build $buildFlags -o plate.exe

if ($LASTEXITCODE -eq 0) {
    Write-Host "Build successful! Output: plate.exe"

    # Try to sign the executable if signtool is available
    $signtool = Get-Command "signtool.exe" -ErrorAction SilentlyContinue
    if ($signtool) {
        Write-Host "Signing executable..."
        & signtool.exe sign /a /tr http://timestamp.digicert.com /td sha256 /fd sha256 plate.exe
    }
} else {
    Write-Host "Build failed with exit code $LASTEXITCODE"
    exit $LASTEXITCODE
}
