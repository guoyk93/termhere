# termhere

A simple reverse shell tunnel with pty support (window size, ctrl-c, ctrl-d, etc.)

## Usage

1. Get the binary `termhere` from [GitHub Releases](https://github.com/guoyk93/termhere/releases)

2. Start `termhere server` in **Machine 1**

   ```shell
   export TERMHERE_TOKEN=aaa
   ./termhere server
   ```

3. Start `termhere client` in **Machine 2**

   ```shell
   export TERMHERE_TOKEN=aaa
   ./termhere client -s 10.10.10.10:7777 /bin/bash
   ```

Now you can get a shell of **Machine 2** from the `termhere server` running in **Machine 1**

## TLS Support

`termhere` supports TLS with client authentication, you can use `--cert-file` and `--key-file` to enable TLS, and
use `--client-ca-file` to enable client authentication.

**TLS Simple**

```shell
# server
termhere server -l 'tcp+tls://:7777' --cert-file server.full-crt.pem --key-file server.key.pem
# client
termhere client -s "tcp+tls://127.0.0.1:7777" --ca-file rootca.crt.pem
```

**TLS with Client Auth**

```shell
# server
termhere server -l 'tcp+tls://:7777' --cert-file server.full-crt.pem --key-file server.key.pem --client-ca-file rootca.crt.pem
# client
termhere client -s "tcp+tls://127.0.0.1:7777" --ca-file rootca.crt.pem --cert-file client.full-crt.pem --key-file client.key.pem
```

Fore more information about **TLS Support**, please refer to [uniconn](https://github.com/guoyk93/uniconn)

## 捐赠

View https://guoyk.xyz/donation

## Credits

GUO YANKE, MIT License
