### `ulimit` values

```
$ ulimit -a
core file size          (blocks, -c) unlimited
data seg size           (kbytes, -d) unlimited
scheduling priority             (-e) 0
file size               (blocks, -f) unlimited
pending signals                 (-i) 127574
max locked memory       (kbytes, -l) 16384
max memory size         (kbytes, -m) unlimited
open files                      (-n) 8388608
pipe size            (512 bytes, -p) 8
POSIX message queues     (bytes, -q) 819200
real-time priority              (-r) 0
stack size              (kbytes, -s) 8192
cpu time               (seconds, -t) unlimited
max user processes              (-u) 8388608
virtual memory          (kbytes, -v) unlimited
file locks                      (-x) unlimited
```

### `sysctl` values

```
fs.nr_open = 41943040
kernel.pid_max = 524288
kernel.threads-max = 8388608
```

### Running Test App

```
go run consumer.go
```

The app will connect in batches of `2048` and then will wait forever. While it is waiting, kill it using `CTRL-C`. You will slowly see queues being deleted in the management UI. If you re-start the app, you should notice that the first batch takes much, much longer than before.
