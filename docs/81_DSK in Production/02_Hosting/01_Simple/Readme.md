# Simple

The following example assumes that you've installed the `dsk` binary to
`/bin/dsk`, are keeping the [DDT](/The-Design-Definitions-Tree) in `/var/ddt`.
and that your operating system uses systemd as its init system.

As long as you don't want any SSL encryption (you probably do), this is the
quickest way to get started. It keeps DSK running on your server and answers
requests directly.

```ini
[Unit]
Description=Design System Kit

[Service]
ExecStart=/bin/dsk -host 192.168.1.1 -port 80 /var/ddt
WorkingDirectory=/var/ddt

[Install]
WantedBy=default.target
```

Please replace `192.168.1.1` with the public IP address of
your machine. After [installing and starting the service unit](https://www.digitalocean.com/community/tutorials/how-to-use-systemctl-to-manage-systemd-services-and-units),
the web interface should be available.
