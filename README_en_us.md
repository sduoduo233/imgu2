# imgu2

An image sharing platform powered by Golang

# Screenshots

<details>
  <summary>Show screenshots</summary>

![image preview page](https://github.com/sduoduo233/imgu2/blob/master/screenshots/1.png?raw=true)

![my uploads page](https://github.com/sduoduo233/imgu2/blob/master/screenshots/2.png?raw=true)

![user list](https://github.com/sduoduo233/imgu2/blob/master/screenshots/3.png?raw=true)

![image list](https://github.com/sduoduo233/imgu2/blob/master/screenshots/4.png?raw=true)

![login page](https://github.com/sduoduo233/imgu2/blob/master/screenshots/5.png?raw=true)

</details>

# Features

- Lightweight design
- OAuth login with Google & GitHub
- Image re-encoding, with support for WebP, PNG, JPEG, GIF and AVIF.
- SQLite database integration
- Multiple storage options supported, including S3-compatible, FTP, and local file systems

# Setup development environment

0. Clone this repo

1. Install Golang

2. Install necessary packages

```bash
# Ubuntu
sudo apt install build-essential libglib2.0-dev libvips-dev libheif-dev libheif-plugin-* libheif1

# Arch Linux
sudo pacman -S libvips libheif pkg-config gcc openslide imagemagick poppler-glib
```

3. `go build`

4. You might need to set the `IMGU2_DEBUG_LIBHEIF_PLUGIN_PATHS` environment variable. Read `libvips/libvips.go` for more information.

# How to install

1. Pull the docker image `docker pull sduoduo233/imgu2:latest`

2. Start container

```bash
docker run --detach -p 3000:3000 -e IMGU2_SMTP_USERNAME="mailer@example.com"  -e IMGU2_SMTP_PASSWORD="example_password" -e IMGU2_SMTP_HOST="example.com" -e IMGU2_SMTP_PORT=25 -e IMGU2_SMTP_SENDER="mailer@example.com" -e IMGU2_JWT_SECRET="example_secret_string" -v ./db:/app/sqlite -v ./uploads:/app/uploads sduoduo233/imgu2:latest
```

`IMGU2_JWT_SECRET` should be a hard-to-guess string, which can be generated using `openssl rand -hex 8` on Linux.

`IMGU2_SMTP_*` are the SMTP configurations required for sending verification emails.

Or you may use this docker compose file:

```yaml
version: '3.8'

services:
  imgu2:
    image: sduoduo233/imgu2:latest
    ports:
      - "3000:3000"
    environment:
      IMGU2_SMTP_USERNAME: "mailer@example.com"
      IMGU2_SMTP_PASSWORD: "example_password"
      IMGU2_SMTP_AUTH_TLS: "true"  # use TLS
      IMGU2_SMTP_HOST: "example.com"
      IMGU2_SMTP_PORT: "25"
      IMGU2_SMTP_SENDER: "mailer@example.com"
      IMGU2_JWT_SECRET: "example_secret_string"
    volumes:
      - ./db:/app/sqlite
      - ./uploads:/app/uploads
    restart: unless-stopped
```

3. Visit `http://YOUR_IP:3000`

4. Configure NGINX for Reverse Proxying and SSL. Add the following configuration to NGINX:

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

6. The default admin email is `admin@example.com`, and the password is `admin`. It is crucial to change these credentials as soon as you deploy the platform for security reasons.