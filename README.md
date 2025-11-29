# Remote Lock

A simple HTTP server that lets me remotely lock my PC via Tailscale (e.g. from my Home Assistant instance). Mostly coded by ChatGPT.

Designed for use with Dank Material Shell only, but you can of course adjust the command it runs to make it compatible with Hyprland or probably anything else:

```go
cmd := exec.Command("dms", "ipc", "call", "lock", "lock")
```

Also designed to only allow requests from a provided tailscale hostname (`ALLOWED_CLIENT` env var).

## Development

```bash
cp .env.example .env
# edit .env
go run .
```

## Building

```bash
go build -o dist/
```

And then copy it somewhere I guess:

```bash
cp dist/remote-lock ~/.local/bin/remote-lock
```

## Deployment

First, create a `remote-lock.env` file somewhere (based on `.env.example`), and **ensure it's not world-readable** (e.g. `chmod 600`) (secret tokens be present!).

Create `~/.config/systemd/user/remote-lock.service`:

```ini
[Unit]
Description=Remote Lock Service
After=tailscaled.service
Wants=tailscaled.service

[Service]
Type=simple
ExecStart=# /path/to/your/remote-lock/binary
Restart=on-failure
User=# your user name?
EnvironmentFile=# /path/to/your/remote-lock.env

[Install]
WantedBy=multi-user.target
```

Then:

```bash
sudo systemctl daemon-reload
systemctl --user start remote-lock.service
# Check status (optional)
systemctl --user status remote-lock.service
# Check logs (optional)
journalctl --user -u remote-lock.service -f
# Enable on boot, if all looks good
systemctl --user enable remote-lock.service
```

## Home Assistant action

Integrate it into Home Assistant by adding something like this to your `configuration.yaml`:

```yaml
rest_command:
  pc_lock:
    url: "http://100.78.139.128:51335/lock" # remote-pc-lock-service
    method: POST
    headers:
      X-Token: !secret pc_lock_token
    timeout: 3
    verify_ssl: false
```

Create the `pc_lock_token` secret (in `secrets.yaml`) and adjust the hostname and port to match your setup. Good luck!
