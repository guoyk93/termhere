# termhere

一个带有 PTY 支持的简单的反向 Shell 工具（支持窗口尺寸，Ctrl+C，Ctrl+D 等）

## 使用方法

1. 从 [GitHub Releases](https://github.com/guoyk93/termhere/releases) 获取二进制文件 `termhere`

2. 在 **控制机** 上启动 `termhere server`，这会监听一个端口

   ```shell
   export TERMHERE_TOKEN=aaa
   ./termhere server
   ```

3. 在 **受控机** 上启动 `termhere client`，这会连接 **控制机** 的端口

   ```shell
   export TERMHERE_TOKEN=aaa
   ./termhere client -s 10.10.10.10:7777 /bin/bash
   ```

此时就可以在 **控制机** 的终端上，操作 **受控机** 了

## TLS 支持

`termhere` 支持 TLS，并且支持 TLS 的客户端验证模式

**单纯 TLS**

```shell
# server
termhere server -l 'tcp+tls://:7777' --cert-file server.full-crt.pem --key-file server.key.pem
# client
termhere client -s "tcp+tls://127.0.0.1:7777" --ca-file rootca.crt.pem
```

**TLS 带客户端验证**

```shell
# server
termhere server -l 'tcp+tls://:7777' --cert-file server.full-crt.pem --key-file server.key.pem --client-ca-file rootca.crt.pem
# client
termhere client -s "tcp+tls://127.0.0.1:7777" --ca-file rootca.crt.pem --cert-file client.full-crt.pem --key-file client.key.pem
```

更多参数可以查阅 [uniconn](https://github.com/guoyk93/uniconn) 文档

## 捐赠

查看 https://guoyk.net/donation

## 许可证

GUO YANKE, MIT License
