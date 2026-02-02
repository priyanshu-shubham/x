# x CLI installer for Windows PowerShell

$ErrorActionPreference = "Stop"

$Repo = "priyanshu-shubham/x"
$BinaryName = "x.exe"

function Get-LatestVersion {
    $response = Invoke-RestMethod -Uri "https://api.github.com/repos/$Repo/releases/latest"
    return $response.tag_name
}

function Get-Architecture {
    if ([Environment]::Is64BitOperatingSystem) {
        if ($env:PROCESSOR_ARCHITECTURE -eq "ARM64") {
            return "arm64"
        }
        return "amd64"
    }
    throw "32-bit systems are not supported"
}

function Install-X {
    Write-Host "Detecting system..." -ForegroundColor Green

    $arch = Get-Architecture
    Write-Host "Architecture: $arch" -ForegroundColor Green

    $version = Get-LatestVersion
    Write-Host "Latest version: $version" -ForegroundColor Green

    $filename = "x_windows_$arch.zip"
    $downloadUrl = "https://github.com/$Repo/releases/download/$version/$filename"

    Write-Host "Downloading from $downloadUrl" -ForegroundColor Green

    # Create temp directory
    $tempDir = Join-Path $env:TEMP "x-install"
    if (Test-Path $tempDir) {
        Remove-Item -Recurse -Force $tempDir
    }
    New-Item -ItemType Directory -Path $tempDir | Out-Null

    # Download
    $zipPath = Join-Path $tempDir $filename
    Invoke-WebRequest -Uri $downloadUrl -OutFile $zipPath

    # Extract
    Expand-Archive -Path $zipPath -DestinationPath $tempDir

    # Install to user's local bin
    $installDir = Join-Path $env:LOCALAPPDATA "x"
    if (-not (Test-Path $installDir)) {
        New-Item -ItemType Directory -Path $installDir | Out-Null
    }

    $binaryPath = Join-Path $tempDir $BinaryName
    $destPath = Join-Path $installDir $BinaryName
    Copy-Item -Path $binaryPath -Destination $destPath -Force

    # Clean up
    Remove-Item -Recurse -Force $tempDir

    Write-Host "Installed to $destPath" -ForegroundColor Green

    # Check if install dir is in PATH
    $userPath = [Environment]::GetEnvironmentVariable("Path", "User")
    if ($userPath -notlike "*$installDir*") {
        Write-Host "Adding $installDir to PATH..." -ForegroundColor Yellow
        [Environment]::SetEnvironmentVariable("Path", "$userPath;$installDir", "User")
        Write-Host "Added to PATH. Restart your terminal to use 'x' command." -ForegroundColor Yellow
    }

    Write-Host "Done! Run 'x configure' to set up authentication." -ForegroundColor Green
}

Install-X
