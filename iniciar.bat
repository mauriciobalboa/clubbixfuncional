@echo off
title Clubbix Server
echo ================================================
echo   CLUBBIX — Iniciando servidor...
echo ================================================
echo.

set "GOPATH=C:\Program Files\Go\bin"
set "PATH=%GOPATH%;%PATH%"

cd /d "%~dp0backend"

if not exist "clubbix.exe" (
    echo Compilando o servidor pela primeira vez...
    go build -o clubbix.exe .
    if errorlevel 1 (
        echo ERRO: falha ao compilar. Verifique se o Go esta instalado.
        pause
        exit /b 1
    )
    echo Compilacao concluida!
    echo.
)

echo Servidor rodando em: http://localhost:8080
echo.
echo Pressione Ctrl+C para parar.
echo.
clubbix.exe
pause
