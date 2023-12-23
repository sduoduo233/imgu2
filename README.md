# imgu2
An image sharing platform powered by Golang

# Screenshots

<details>
  <summary>Show screenshots</summary>

![image preview page](https://raw.githubusercontent.com/sduoduo233/imgu2/master/screenshots/1.png)

![my uploads page](https://raw.githubusercontent.com/sduoduo233/imgu2/master/screenshots/2.png)

![user list](https://raw.githubusercontent.com/sduoduo233/imgu2/master/screenshots/3.png)

![image list](https://raw.githubusercontent.com/sduoduo233/imgu2/master/screenshots/4.png)

![login page](https://raw.githubusercontent.com/sduoduo233/imgu2/master/screenshots/5.png)

</details>

# Features

- Lightweight design
- OAuth login with Google & GitHub
- Image re-encoding
- SQLite database integration
- Multiple storage options supported, including S3-compatible, FTP, and local file systems

# How to install

1. Get the latest binary from the GitHub releases.

2. Transfer the downloaded file to your Linux web server.

3. In the same directory as the executable file, create a `.env` file with the following content:

```
IMGU2_SMTP_USERNAME=mailer@example.com
IMGU2_SMTP_PASSWORD=example_password
IMGU2_SMTP_HOST=example.xyz
IMGU2_SMTP_PORT=25
IMGU2_SMTP_SENDER=mailer@example.com
IMGU2_JWT_SECRET=example_secret_string
```

`IMGU2_JWT_SECRET` should be a hard-to-guess string, which can be generated using `openssl rand -hex 8` on Linux.

`IMGU2_SMTP_*` are the SMTP configurations required for sending verification emails.

4. Run the executable using `./imgu2-linux-amd64`.

5. Configure NGINX for Reverse Proxying and SSL. Add the following configuration to NGINX:

```nginx
server {
  listen 443 ssl;
  ssl_certificate /path/to/ssl/certificate;
  ssl_certificate_key /path/to/ssl/certificate/key;

  server_name example.com;
  root /var/www/html;
  index index.php index.html;

  client_max_body_size 32M;

  location / {
    proxy_pass http://127.0.0.1:3000;
  }
}
```
