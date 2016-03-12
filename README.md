# docker-volume-rbd

[![Build Status](https://travis-ci.org/h0tbird/docker-volume-rbd.svg?branch=master)](https://travis-ci.org/h0tbird/docker-volume-rbd)

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
I have my Ceph config in `/etc/ceph` and `/var/lib/ceph` (on the host) so I can do this:

###### With `docker` (0.427s, concurrent, chicken/egg dilemma):
```
core@core-1 ~ $ cat /opt/bin/rbd
#!/bin/bash
docker run -i --rm \
--privileged \
--pid host \
--net host \
--volume /dev:/dev \
--volume /sys:/sys \
--volume /etc/ceph:/etc/ceph \
--volume /var/lib/ceph:/var/lib/ceph \
--entrypoint $(basename $0) \
h0tbird/ceph:v9.2.0-2 "$@"
```

###### With `rkt fly` (3s, concurrent, chicken/egg friendly):
```
core@core-1 ~ $ cat /opt/bin/ceph
#!/bin/bash
sudo rkt run \
--stage1-path=/usr/share/rkt/stage1-fly.aci \
--interactive \
--net=host \
--insecure-options=image \
--volume volume-etc-ceph,kind=host,source=/etc/ceph \
--volume volume-var-lib-ceph,kind=host,source=/var/lib/ceph docker://h0tbird/ceph:v9.2.0-2 \
--exec /usr/bin/$(basename $0) -- "$@"
```

###### With `systemd-nspawn` (0.248s, nonconcurrent, chicken/egg friendly):
```
core@core-1 ~ $ cat /opt/bin/rbd
#!/bin/bash

readonly CEPH_DOCKER_IMAGE=h0tbird/ceph
readonly CEPH_DOCKER_TAG=v9.2.0-2
readonly CEPH_USER=root

machinename=$(echo "${CEPH_DOCKER_IMAGE}-${CEPH_DOCKER_TAG}" | sed -r 's/[^a-zA-Z0-9_.-]/_/g')
machinepath="/var/lib/toolbox/${machinename}"
osrelease="${machinepath}/etc/os-release"

[ -f ${osrelease} ] || {
  sudo mkdir -p "${machinepath}"
  sudo chown ${USER}: "${machinepath}"
  docker pull "${CEPH_DOCKER_IMAGE}:${CEPH_DOCKER_TAG}"
  docker run --name=${machinename} "${CEPH_DOCKER_IMAGE}:${CEPH_DOCKER_TAG}" /bin/true
  docker export ${machinename} | sudo tar -x -C "${machinepath}" -f -
  docker rm ${machinename}
  sudo touch ${osrelease}
}

[ "$1" == 'dryrun' ] || {
  sudo systemd-nspawn \
  --quiet \
  --directory="${machinepath}" \
  --capability=all \
  --share-system \
  --bind=/dev:/dev \
  --bind=/etc/ceph:/etc/ceph \
  --bind=/var/lib/ceph:/var/lib/ceph \
  --user="${CEPH_USER}" \
  --setenv=CMD="$(basename $0)" \
  --setenv=ARG="$*" \
  /bin/bash -c '\
  mount -o remount,rw -t sysfs sysfs /sys; \
  $CMD $ARG'
}
```
