# docker-volume-rbd
A Docker volume driver for RBD

##### CoreOS
If you are a CoreOS user (like me) you must provide a way to run the `rbd` command.
I have my Ceph config in `/etc/ceph` so I can do this:
```
core@core-1 ~ $ cat /opt/bin/rbd
#!/bin/bash
docker run -i --rm \
--privileged \
--pid host \
--name rbd \
--net host \
--volume /dev:/dev \
--volume /sys:/sys \
--volume /etc/ceph:/etc/ceph \
ceph/base rbd $@
```
