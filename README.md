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

## Credits

GUO YANKE, MIT License