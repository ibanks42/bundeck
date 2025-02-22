@echo off

REM Kill the bundeck-dbg process
taskkill /f /im bundeck-dbg.exe

REM Navigate to the web directory
cd web

REM Run the build command
bun run build

REM Navigate back to the parent directory
cd ..

REM Build the Go application
go build -o ./bundeck-dbg.exe .

.\bundeck-dbg.exe
