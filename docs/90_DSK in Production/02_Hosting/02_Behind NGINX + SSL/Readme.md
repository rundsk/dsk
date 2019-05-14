# With NGINX as a reverse-proxy and SSL

The following example assumes that you've installed the `dsk` binary to
`/bin/dsk`, are keeping the [DDT](/The-Design-Definitions-Tree) in `/var/ddt`
and that your operating system uses systemd as its init system.

For SSL support we'll put DSK behind NGINX. The webserver will do the
termination for us, then forward all requests to DSK. DSK will be listening on
the loopback interface on port 8080.

```ini
[Unit]
Description=Design System Kit

[Service]
ExecStart=/bin/dsk -port 8080 /var/ddt
User=www-data
Group=www-data
WorkingDirectory=/var/ddt

[Install]
WantedBy=default.target
```

```nginx
server {
	listen 443 ssl http2;

	server_name example.com;
	root /var/ddt;

	ssl_certificate /etc/ssl/certs/example.com.crt;
	ssl_certificate_key /etc/ssl/private/example.com.key;

	location / {
		proxy_set_header X-Real-IP $remote_addr;
		proxy_set_header X-Forwarded-For $remote_addr;
		proxy_set_header Host $host;
		proxy_pass http://127.0.0.1:8080;
	}	
}
```

