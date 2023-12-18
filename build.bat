@echo off

go build -ldflags="-H=windowsgui"
go install -ldflags="-H=windowsgui"
