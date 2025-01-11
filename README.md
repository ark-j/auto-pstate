# auto-pstate
auto-pstate is simple script for switching amd epp pstate automatically. you can install using one line
`curl -sSL https://github.com/ark-j/auto-pstate/releases/download/0.0.2/install | bash`

# enable epp
- first you need to enable amd-epp by setting kernel parameter `amd_pstate=active`

# Building it locally
- scripts need golang to initially compile binary
- then you can clone this repo
- and run `install.sh`

# Post install
- Add your user to auto-pstate group `sudo usermod -aG auto-pstate $USER`
