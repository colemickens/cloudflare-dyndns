build:
	gb build

run: build
	./bin/cloudflare-dyndns \
		-key "" \
		-email "cole.mickens@gmail.com" \
		-records "*.mickens.xxx,mickens.xxx,*.mickens.io,mickens.io,recessionomics.us,www.recessionomics.us,*.mickens.me,mickens.me,*.mickens.tv,mickens.tv,*.mickens.us,mickens.us,cole.mickens.us"

install-systemd:
	sudo cp systemd/cloudflare-dyndns.service /etc/systemd/system/
	sudo cp systemd/cloudflare-dyndns.timer /etc/systemd/system/
	sudo systemctl enable cloudflare-dyndns.service
	sudo systemctl enable cloudflare-dyndns.timer
	sudo systemctl start cloudflare-dyndns.service
	sudo systemctl start cloudflare-dyndns.timer
