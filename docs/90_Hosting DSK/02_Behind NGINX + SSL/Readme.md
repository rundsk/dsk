# With NGINX as a reverse-proxy and SSL

For SSL support we'll put DSK behind NGINX. The webserver will do the
termination for us, then forward all requests to DSK. DSK will be listening on
the loopback interface on port 8080.

```ini
# ...

[Service]
ExecStart=/bin/dsk -port 8080 /var/ds
User=www-data
Group=www-data
WorkingDirectory=/var/ds

# ...
```

```nginx
server {
	listen 443 ssl http2;

	server_name example.com;
	root /var/ds;

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

