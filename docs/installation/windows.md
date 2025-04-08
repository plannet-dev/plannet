# Installing Plannet on Windows

## Quick Install (Recommended)

### For x64 Systems

1. Download the latest release:
```powershell
# Using PowerShell
Invoke-WebRequest -Uri "https://github.com/plannet-dev/plannet/releases/latest/download/plannet-windows-amd64.exe" -OutFile "plannet.exe"
```

2. Move to a directory in your PATH:
```powershell
# Create a directory for Plannet (if it doesn't exist)
New-Item -ItemType Directory -Force -Path "$env:USERPROFILE\bin"

# Move Plannet to the directory
Move-Item -Force plannet.exe "$env:USERPROFILE\bin\plannet.exe"
```

3. Add to PATH:
```powershell
# Add to user PATH
$userPath = [Environment]::GetEnvironmentVariable("Path", "User")
if ($userPath -notlike "*$env:USERPROFILE\bin*") {
    [Environment]::SetEnvironmentVariable("Path", "$userPath;$env:USERPROFILE\bin", "User")
}
```

4. Verify installation (in a new PowerShell window):
```powershell
plannet --version
```

### For ARM64 Systems

1. Download the latest release:
```powershell
# Using PowerShell
Invoke-WebRequest -Uri "https://github.com/plannet-dev/plannet/releases/latest/download/plannet-windows-arm64.exe" -OutFile "plannet.exe"
```

2. Move to a directory in your PATH:
```powershell
# Create a directory for Plannet (if it doesn't exist)
New-Item -ItemType Directory -Force -Path "$env:USERPROFILE\bin"

# Move Plannet to the directory
Move-Item -Force plannet.exe "$env:USERPROFILE\bin\plannet.exe"
```

3. Add to PATH:
```powershell
# Add to user PATH
$userPath = [Environment]::GetEnvironmentVariable("Path", "User")
if ($userPath -notlike "*$env:USERPROFILE\bin*") {
    [Environment]::SetEnvironmentVariable("Path", "$userPath;$env:USERPROFILE\bin", "User")
}
```

4. Verify installation (in a new PowerShell window):
```powershell
plannet --version
```

## Manual Installation

If you prefer to install Plannet manually:

1. Download the appropriate executable from the [releases page](https://github.com/plannet-dev/plannet/releases)
2. Move the executable to a directory of your choice
3. Add that directory to your PATH:
   - Open System Properties (Win + Pause/Break)
   - Click "Advanced system settings"
   - Click "Environment Variables"
   - Under "User variables", find and select "Path"
   - Click "Edit"
   - Click "New"
   - Add the directory containing plannet.exe
   - Click "OK" on all windows
4. Open a new PowerShell window and verify the installation:
```powershell
plannet --version
``` 