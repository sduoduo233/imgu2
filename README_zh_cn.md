# imgu2
使用Golang编写的图片分享平台（图床）。

# 截图

<details>
  <summary>显示截图</summary>

![image preview page](https://github.com/sduoduo233/imgu2/blob/master/screenshots/1.png?raw=true)

![my uploads page](https://github.com/sduoduo233/imgu2/blob/master/screenshots/2.png?raw=true)

![user list](https://github.com/sduoduo233/imgu2/blob/master/screenshots/3.png?raw=true)

![image list](https://github.com/sduoduo233/imgu2/blob/master/screenshots/4.png?raw=true)

![login page](https://github.com/sduoduo233/imgu2/blob/master/screenshots/5.png?raw=true)

</details>

# Features

- 轻量级设计
- Google & GitHub OAuth 登陆
- 图片重新编码
- SQLite 数据库
- 多种存储后端, 包括 S3-compatible, FTP, 和本地存储

# 如何安装

1. 在 Release 中下载最新的二进制文件

2. 把二进制文件上传到 Linux VPS

3. 在相同目录创建一个 `.env` 文件，填入以下内容：

```
IMGU2_SMTP_USERNAME=mailer@example.com
IMGU2_SMTP_PASSWORD=example_password
IMGU2_SMTP_HOST=example.com
IMGU2_SMTP_PORT=25
IMGU2_SMTP_SENDER=mailer@example.com
IMGU2_JWT_SECRET=example_secret_string
```

`IMGU2_JWT_SECRET` 应该是一个比较难猜到的长字符串, 在 Linux 中可以用 `openssl rand -hex 8` 生成。

`IMGU2_SMTP_*` 是SMTP设置，用以发送验证邮件。

4. 启动程序 `./imgu2-linux-amd64`.

5. 配置 NGINX:

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