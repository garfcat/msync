# msync

msync is a command-line program to sync files and directories by multi-process .

# install

1. install [rclone](https://github.com/rclone/rclone).
2. git clone msync and build it.

# using

```shell
Usage of msync:
  -d, --dst string   dest directory
  -s, --src string   source directory
  -t, --thread int   the num of max thread (default 1)
```

```shell
 msync -s SRC  -d  DST  -t Parallel
```