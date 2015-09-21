I will be maintaining the rust version of this going forward: [cloudflare-dyndns-rust](https://github.com/colemickens/cloudflare-dyndns-rust), so please give it a shot and file any issues against that, if you'd like me to take a look.

---

# cloudflare-dyndns

This is a tool (golang app + systemd unit files) that will run a dynamic DNS client on your machine in order to keep Cloudflare DNS records up to date and pointing at the network that this tool runs from.

## Building

You will need the `gb` tool for building go code: http://getgb.io

`make` will restore the vendored packages using the gb-vendor plugin, and then it will build the `cloudflare-dyndns` binary.

That's all.

## Install the Service

First, update `cloudflare-dyndns.service` to point to the location of the built binary. It currently references a path that is probably only present on my machines.

Second, copy `cloudflare-dyndns.config.example` to `cloudflare-dyndns.config` and make the appropiate changes to the file, inserting your CloudFlare email and api-key.

Third, edit the `cloudflare-dyndns.timer` unit file so that the app will run as often as you need. It's set to run every 5 minutes by default.

Finally, execute `make install-systemd` to install the config file and the systemd unit files.

## Uninstall the Service

If you'd like to undo the installation steps, you may run `make uninstall-systemd` to remove the config files, disable the service and remove the service unit files.
