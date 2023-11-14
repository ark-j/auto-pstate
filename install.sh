#!/bin/bash
go mod download
go build -o auto-pstate -ldflags="-s -w" main.go
sudo cp auto-pstate /usr/bin/ -v
sudo cp auto-pstate.service /etc/systemd/system/ -v
sudo systemctl enable auto-pstate
sudo systemctl start auto-pstate