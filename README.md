# siren - image builder for systemd-machined

siren is a tiny and really fast (<3 overlayfs) tool for building OS containers for use with systemd-nspawn and machinectl.

## WIP
This is a rewrite of [siren-sh](https://github.com/LEW21/siren-sh) in Go. It still needs more testing and documentation.

## Install
```console
$ mkdir siren-build
$ GOPATH=`pwd`/siren-build go get github.com/LEW21/siren
```

Siren will be compiled as a single static binary called `siren`, saved in the `siren-build/bin/` directory. You can copy it wherever you want.

## Sirenfiles
Sirenfiles are text documents containing all the commands necessary for building an image. They are quite similar to Dockerfiles. You can find multiple ready to use Sirenfiles at [LEW21/sirenfiles](https://github.com/LEW21/sirenfiles).

## How it works
The only magical thing in siren is the way we compose image layers using overlayfs. For each built image, we create new var-lib-machines-imagename.mount and var-lib-machines-imagename.automount systemd units, and enable the second one. This way, systemd automatically mounts overlayfs at /var/lib/machines/imagename when somebody tries to access it, for example by starting a siren-built container. And this way, you don't need siren anywhere in the process of running containers, it's role is finished after the container is built.

## License
MIT license.

## Requirements
* Linux 4.0 (overlayfs with multiple read-only layers)
* systemd 220 (219 has NAT support for containers, required for internet access when running (but not building) containers)
* NO btrfs (Non-empty directory removal is bugged on overlayfs over btrfs. Use a good fs instead - eg. ext4.)
* sysctl net.ipv4.ip_forward=1 (I don't know if it's always required, but on my local PC containers can't access internet without it)
