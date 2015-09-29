# docker-volume-rbd
A Docker volume driver for RBD

##### Sample
Client side:
```
core@core-1 ~ $ docker run -it --volume-driver rbd -v foo:/foo alpine /bin/sh -c "echo -n 'hello ' > /foo/hw.txt"
core@core-1 ~ $ docker run -it --volume-driver rbd -v foo:/foo alpine /bin/sh -c "echo world >> /foo/hw.txt"
core@core-1 ~ $ docker run -it --volume-driver rbd -v foo:/foo alpine cat /foo/hw.txt
hello world
```
Server side:
```
core@core-1 ~ $ sudo ./docker-volume-rbd
2015/09/29 13:52:23 [Init] INFO volume root is /var/lib/docker/volumes/rbd
2015/09/29 13:52:23 [Init] INFO loading RBD kernel module...
2015/09/29 13:52:23 [Init] INFO listening on /var/run/docker/plugins/rbd.sock
2015/09/29 13:52:30 [Create] INFO image does not exists. Creating it now...
2015/09/29 13:52:33 [Mount] INFO locking image foo
2015/09/29 13:52:33 [Mount] INFO mapping image foo
2015/09/29 13:52:34 [Mount] INFO creating /var/lib/docker/volumes/rbd/rbd/foo
2015/09/29 13:52:34 [Mount] INFO mounting device /dev/rbd0
2015/09/29 13:52:34 [Unmount] INFO unmounting device /dev/rbd0
2015/09/29 13:52:34 [Unmount] INFO unmapping image foo
2015/09/29 13:52:34 [Unmount] INFO unlocking image foo
2015/09/29 13:52:40 [Mount] INFO locking image foo
2015/09/29 13:52:40 [Mount] INFO mapping image foo
2015/09/29 13:52:41 [Mount] INFO creating /var/lib/docker/volumes/rbd/rbd/foo
2015/09/29 13:52:41 [Mount] INFO mounting device /dev/rbd0
2015/09/29 13:52:41 [Unmount] INFO unmounting device /dev/rbd0
2015/09/29 13:52:41 [Unmount] INFO unmapping image foo
2015/09/29 13:52:42 [Unmount] INFO unlocking image foo
2015/09/29 13:52:48 [Mount] INFO locking image foo
2015/09/29 13:52:48 [Mount] INFO mapping image foo
2015/09/29 13:52:49 [Mount] INFO creating /var/lib/docker/volumes/rbd/rbd/foo
2015/09/29 13:52:49 [Mount] INFO mounting device /dev/rbd0
2015/09/29 13:52:49 [Unmount] INFO unmounting device /dev/rbd0
2015/09/29 13:52:49 [Unmount] INFO unmapping image foo
2015/09/29 13:52:49 [Unmount] INFO unlocking image foo
```

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
