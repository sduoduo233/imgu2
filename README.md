# imgu2
使用Golang编写的图片分享平台（图床）。

[English](https://github.com/sduoduo233/imgu2/blob/master/README_en_us.md)

帮助我们翻译:

[![Crowdin](https://badges.crowdin.net/imgu2/localized.svg)](https://crowdin.com/project/imgu2)


# 截图

<details>
  <summary>显示截图</summary>

![image preview page](https://github.com/sduoduo233/imgu2/blob/master/screenshots/1.png?raw=true)

![my uploads page](https://github.com/sduoduo233/imgu2/blob/master/screenshots/2.png?raw=true)

![user list](https://github.com/sduoduo233/imgu2/blob/master/screenshots/3.png?raw=true)

![image list](https://github.com/sduoduo233/imgu2/blob/master/screenshots/4.png?raw=true)

![login page](https://github.com/sduoduo233/imgu2/blob/master/screenshots/5.png?raw=true)

</details>

# 功能

- 轻量级设计
- Google & GitHub OAuth 登陆
- 图片重新编码，支持 WebP, PNG, JPEG, GIF 和 AVIF
- SQLite 数据库
- 多种存储后端, 包括 S3-compatible, FTP, 和本地存储

# 配置开发环境

0. Clone 这个库

1. 安装 Golang

2. 安装必要的包

```bash
# Ubuntu
sudo apt install build-essential libglib2.0-dev libvips-dev libheif-dev libheif-plugin-* libheif1

# Arch Linux
sudo pacman -S libvips libheif pkg-config gcc openslide imagemagick poppler-glib

```

3. `go build`

4. 你可能需要设置 `IMGU2_DEBUG_LIBHEIF_PLUGIN_PATHS` 环境变量. 阅读 `libvips/libvips.go` 了解详细步骤.

# 如何安装

1. 下载 Docker 镜像 `docker pull sduoduo233/imgu2:latest`

2. 启动容器

```bash
docker run --detach -p 3000:3000 -e IMGU2_SMTP_USERNAME="mailer@example.com"  -e IMGU2_SMTP_PASSWORD="example_password" -e IMGU2_SMTP_HOST="example.com" -e IMGU2_SMTP_PORT=25 -e IMGU2_SMTP_SENDER="mailer@example.com" -e IMGU2_SMTP_AUTH_TLS="false" -e IMGU2_JWT_SECRET="example_secret_string" -v ./db:/app/sqlite -v ./uploads:/app/uploads sduoduo233/imgu2:latest
```

`IMGU2_JWT_SECRET` 应该是一个比较难猜到的长字符串, 在 Linux 中可以用 `openssl rand -hex 8` 生成。

`IMGU2_SMTP_*` 是SMTP设置，用以发送验证邮件。


或者你可以用 docker compose:

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
      IMGU2_SMTP_HOST: "example.com"
      IMGU2_SMTP_AUTH_TLS: "false"
      IMGU2_SMTP_PORT: "25"
      IMGU2_SMTP_SENDER: "mailer@example.com"
      IMGU2_JWT_SECRET: "example_secret_string"
    volumes:
      - ./db:/app/sqlite
      - ./uploads:/app/uploads
    restart: unless-stopped
```

3. 访问 `http://你的IP地址:3000`

4. 配置 NGINX:

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

6. 默认管理员账号是 `admin@example.com`, 密码是 `admin`。部署完成后请立刻改密码。