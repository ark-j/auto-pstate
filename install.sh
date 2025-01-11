#!/bin/bash

set -euo pipefail

# build the targets
go mod download
go build -o auto-pstate -ldflags="-s -w" ./cmd/client/main.go
go build -o pstate-daemon -ldflags="-s -w" ./cmd/server/main.go

# copy targets and service files
sudo cp auto-pstate /usr/bin/ -v
sudo cp pstate-daemon /usr/bin/ -v
sudo cp auto-pstate.service /etc/systemd/system/ -v

# start and enable the services
sudo systemctl enable --now auto-pstate

# create auto-psate group
sudo groupadd auto-pstate
# bind to root user for necessary permissions
sudo chown root:auto-pstate /run/auto-epp/epp.sock
# modify read and write permissions to be able to used by group and root
sudo chmod 660 /run/auto-epp/epp.sock

# clean
rm -rf auto-pstate pstate-daemon
