@echo off

set GOARCH=386

go build -ldflags "-H windowsgui" -o ../build/winmix.exe ../
