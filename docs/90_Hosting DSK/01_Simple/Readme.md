# Simple

As long as you don't want any SSL encryption (you probably do), this is the
quickest way to get started. It keeps DSK running on your server and answers
requests directly.

Please replace `192.168.1.1` with the public IP address of
your machine. After [installing and starting the service
unit](https://www.digitalocean.com/community/tutorials/how-to-use-systemctl-to-manage-systemd-services-and-units), the web interface should be available.


```ini
[Unit]
Description=Design System Kit

[Service]
ExecStart=/bin/dsk -host 192.168.1.1 -port 80 /var/ds
WorkingDirectory=/var/ds

[Install]
WantedBy=default.target
```

