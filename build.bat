@echo off

go build -ldflags="-H=windowsgui" ./cmd/preditor
go install -ldflags="-H=windowsgui" ./cmd/preditor
