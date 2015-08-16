I need a small bit of logic to run every 5 minutes to make sure that my Cloudflare records are up to date.

The initial implementation was in Powershell, which I prefer.
But, I need this to run every 5 minutes on my server running Linux, hence the Golang version.

There were plans to add a shell/bash version, but the golang version works, the systemd service is running and I hate shell scripting, so this is probably "done".

