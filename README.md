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

Note that the `.mvn/jvm.config` file exists with this content:

```
-Xmx16G -Xms128m -Xss256k -Djava.awt.headless=true
```

Here is how I am running the test app:

```
$ mvn compile && mvn -e exec:java -Dexec.mainClass=com.rabbitmq.mqttTest.App -Dexec.args='3000 1 1'
Picked up _JAVA_OPTIONS: -Dawt.useSystemAAFontSettings=on -Dswing.aatext=true
[INFO] Scanning for projects...
[INFO] 
[INFO] ------------------------------------------------------------------------
[INFO] Building mqtt-test 1.0-SNAPSHOT
[INFO] ------------------------------------------------------------------------
[INFO] 
[INFO] --- maven-resources-plugin:2.6:resources (default-resources) @ mqtt-test ---
[INFO] Using 'UTF-8' encoding to copy filtered resources.
[INFO] skip non existing resourceDirectory /home/lbakken/issues/pt/153139703-evict-queues/mqtt-test/src/main/resources
[INFO] 
[INFO] --- maven-compiler-plugin:3.1:compile (default-compile) @ mqtt-test ---
[INFO] Nothing to compile - all classes are up to date
[INFO] ------------------------------------------------------------------------
[INFO] BUILD SUCCESS
[INFO] ------------------------------------------------------------------------
[INFO] Total time: 0.413 s
[INFO] Finished at: 2017-12-06T16:34:42-08:00
[INFO] Final Memory: 10M/155M
[INFO] ------------------------------------------------------------------------
Picked up _JAVA_OPTIONS: -Dawt.useSystemAAFontSettings=on -Dswing.aatext=true
[INFO] Error stacktraces are turned on.
[INFO] Scanning for projects...
[INFO] 
[INFO] ------------------------------------------------------------------------
[INFO] Building mqtt-test 1.0-SNAPSHOT
[INFO] ------------------------------------------------------------------------
[INFO] 
[INFO] --- exec-maven-plugin:1.6.0:java (default-cli) @ mqtt-test ---
[WARNING] 
java.lang.OutOfMemoryError: unable to create new native thread
    at java.lang.Thread.start0 (Native Method)
    at java.lang.Thread.start (Thread.java:717)
    at org.eclipse.paho.client.mqttv3.internal.ClientComms$ConnectBG.start (ClientComms.java:565)
    at org.eclipse.paho.client.mqttv3.internal.ClientComms.connect (ClientComms.java:220)
    at org.eclipse.paho.client.mqttv3.internal.ConnectActionListener.connect (ConnectActionListener.java:166)
    at org.eclipse.paho.client.mqttv3.MqttAsyncClient.connect (MqttAsyncClient.java:497)
    at org.eclipse.paho.client.mqttv3.MqttClient.connect (MqttClient.java:238)
    at com.rabbitmq.mqttTest.App$Worker.run (App.java:97)
    at java.lang.Thread.run (Thread.java:748)
```

Just before the above exception happens, here is the thread count:

```
$ fgrep Threads: /proc/"$(pgrep java)"/status
Threads:        9637
```

That count kept going up and up without limit.
