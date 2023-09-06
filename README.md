# termhere

a simple reverse shell tunnel with pty support

## Usage

1. Get binary `termhere`
2. Run server in machine 1

    ```shell
    export TERMHERE_TOKEN=aaa
    termhere server
    ```

3. Run client in machine 2

    ```shell
    export TERMHERE_TOKEN=aaa
    termhere client -s 10.10.10.10:7777 /bin/bash
    ```
   
Now you can get a shell in machine 1

## Credits

GUO YANKE, MIT License