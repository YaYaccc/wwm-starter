@echo off
setlocal

REM ============================================================================
REM yysls Build Script
REM This script installs the 'rsrc' tool (if missing), generates the .syso resource
REM file containing the icon, and builds the Go application.
REM ============================================================================

REM 1. Setup Environment
echo [Setup] Checking Go environment...
for /f "tokens=*" %%i in ('go env GOPATH') do set GOPATH=%%i
if "%GOPATH%"=="" set GOPATH=%USERPROFILE%\go
set PATH=%PATH%;%GOPATH%\bin

REM 2. Check/Install rsrc tool
echo [Step 1/3] Checking for 'rsrc' tool...
where rsrc >nul 2>nul
if %errorlevel% neq 0 (
    echo 'rsrc' not found. Installing github.com/akavel/rsrc...
    go install github.com/akavel/rsrc@latest
    if %errorlevel% neq 0 (
        echo [Error] Failed to install 'rsrc'. Please check your internet connection.
        pause
        exit /b 1
    )
)

REM 3. Generate Resource File (.syso)
echo [Step 2/3] Embedding icon from resource\favicon.ico...
if not exist "resource\favicon.ico" (
    echo [Error] resource\favicon.ico not found!
    pause
    exit /b 1
)

rsrc -arch amd64 -ico resource\favicon.ico -o wwm-starter.syso
if %errorlevel% neq 0 (
    echo [Error] Failed to generate resource file.
    pause
    exit /b 1
)

REM 4. Build Application
echo [Step 3/3] Building wwm-starter.exe...
go build -ldflags="-s -w" -o wwm-starter.exe .
if %errorlevel% neq 0 (
    echo [Error] Build failed.
    pause
    exit /b 1
)

echo.
echo ========================================
echo  Build Successful!
echo  Output: %CD%\wwm-starter.exe
echo ========================================
echo.

REM Clean up the generated .syso file
if exist wwm-starter.syso (
    echo [Cleanup] Removing wwm-starter.syso...
    del wwm-starter.syso
)

endlocal
