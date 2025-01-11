sudo systemctl disable --now auto-pstate.service

sudo rm -rf /run/auto-epp \
	/usr/bin/auto-pstate \
	/usr/bin/pstate-daemon \
	/etc/systemd/system/auto-pstate.service
