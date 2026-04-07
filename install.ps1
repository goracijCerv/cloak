#Stops script if there are any errors
$ErrorActionPreference ="Stop"

Write-Host "Starting the installation of Cloak..." -ForegroundColor Cyan
#Cheking  the system architecture
$SysArch = $env:PROCESSOR_ARCHITECTURE.ToLower()
$Arch = ""

if ($SysArch -eq "amd64") {
    $Arch = "amd64"
} elseif ($SysArch -eq "arm64") {
    $Arch = "arm64"
} else {
    Write-Host "Unsupported architecture: $SysArch" -ForegroundColor Red
    exit 1
}

Write-Host "Architecture = $Arch"
#Define dynamic rutes
$InstallDir = "$env:LOCALAPPDATA\cloak\bin"
$ExePath = "$InstallDir\cloak.exe"
$BinaryName = "cloak_windows_$Arch.exe"
$RepoUrl = "https://github.com/goracijCerv/cloak/releases/latest/download/$BinaryName"

#Create the folder if doesnt exist
if (-not (Test-Path -Path $InstallDir)){
    Write-Host "Creating the folder in $InstallDir..."
    New-Item -ItemType Directory -Force -Path $InstallDir | Out-Null
}

#Downloading the right exe
Write-Host "Downloading Cloak ($BinaryName) from GitHub..."
Invoke-WebRequest -Uri $RepoUrl -OutFile $ExePath

#Modify the path envariotment variable
Write-Host "Setting the PATH environment variable.."
$UserPath = [Environment]::GetEnvironmentVariable("Path", [System.EnvironmentVariableTarget]::User)

if ($UserPath -notlike "*$InstallDir*"){
    $NewPath = "$UserPath;$InstallDir"
    [Environment]::SetEnvironmentVariable("Path", $NewPath, [System.EnvironmentVariableTarget]::User)
    Write-Host "Directory successfully added to PATH!" -ForegroundColor Green
    $NeedsRestart = $true
} else {
    Write-Host "Directory was already in the PATH."
    $NeedsRestart = $false
}

Write-Host ""
Write-Host "========================================" -ForegroundColor Cyan
Write-Host "Cloak installed successfully!" -ForegroundColor Green
Write-Host "========================================" -ForegroundColor Cyan
Write-Host ""

if ($NeedsRestart){
    Write-Host "IMPORTANT: Since the PATH was updated, you must RESTART this PowerShell window or open a new one to be able to use the 'cloak' command." -ForegroundColor Yellow
} else {
    Write-Host "You can now run the command: cloak --help"
}