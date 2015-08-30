I need a small bit of logic to run every 5 minutes to make sure that my Cloudflare records are up to date.

The initial implementation was in Powershell, which I prefer.
But, I need this to run every 5 minutes on my server running Linux, hence the Golang version.

There were plans to add a shell/bash version, but the golang version works, the systemd service is running and I hate shell scripting, so this is probably "done".

# cloudflare-dyndns

This is a tool (golang app + systemd unit files) that will run a dynamic DNS client on your machine in order to keep Cloudflare DNS records up to date and pointing at the network that this tool runs from.

## Building

You will need the `gb` tool for building go code: http://getgb.io

`make` will restore the vendored packages using the gb-vendor plugin, and then it will build the `cloudflare-dyndns` binary.

That's all.

## Install the Service

First, update `cloudflare-dyndns.service` to point to the location of the built binary. It currently references a path that is probably only present on my machines.

Second, copy `cloudflare-dyndns.config.example` to `cloudflare-dyndns.config` and make the appropiate changes to the file, inserting your CloudFlare email and api-key.

Third, edi the `cloudflare-dyndns.timer` unit file so that the app will run as often as you need. It's set to run every 5 minutes by default.

Finally, execute `make install-systemd` to install the config file and the systemd unit files.

## Uninstall the Service

If you'd like to undo the installation steps, you may run `make uninstall-systemd` to remove the config files, disable the service and remove the service unit files.
