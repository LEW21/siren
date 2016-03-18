# siren - image builder for systemd-machined

siren is a tiny and really fast (<3 overlayfs) tool for building OS containers for use with systemd-nspawn and machinectl.

It contains a subproject called [imagectl](https://github.com/LEW21/siren/tree/master/imagectl) - used for managing images.

## Features
### imagectl
* Seamless integration with systemd
 * machinectl start/stop and all other machine commands (but NOT image / image transfer commands)
 * systemctl start/stop/... systemd-nspawn@...service
 * Without any changes in machined and systemd-nspawn@.service.
* Layered images (using overlayfs)
 * Instant creation of new images using existing ones as a base
* Tagging

### siren
* Dockerfile-like syntax for Sirenfiles.
* Automatic pulling and building of base images from git repositories.

## Sirenfiles
Sirenfiles are text documents containing all the commands necessary for building an image. They are quite similar to Dockerfiles.

```sirenfile
ID http.python 2016.03.09
FROM arch-2016.03.09 git+https://github.com/LEW21/sirenfiles.git#arch

RUN pacman -S --noconfirm python-pip python-crypto
RUN pip install gunicorn

RUN mkdir /app
COPY wsgi.py /app

ENABLE http.socket
ENABLE http.service
```

You can find multiple ready to use Sirenfiles at [LEW21/sirenfiles](https://github.com/LEW21/sirenfiles).

## Usage
```
Usage: siren COMMAND [arg...]
       siren [ -h | --help | -v | --version ]

Image builder for systemd-machined.

Siren Commands:
        build DIR_PATH [TAG]        Build an image from a Sirenfile
        pull URI [TAG]              Pull and build an image from a git repostory

Image Commands:
   new, create NAME [BASE_NAME]     Create a new image
        tag TAG NAME                Create an alias for the image
    ro, set-read-only NAME [BOOL]   Mark or unmark image read-only
        set-ready NAME [BOOL]       Assemble or disassemble layered image
    rm, remove NAME...              Remove an image
    ls, list                        Show available container and VM images
```
machinectl (list-images, read-only) and docker (images) image management command names are also supported.

## Installation
```console
$ mkdir siren-build
$ GOPATH=`pwd`/siren-build go get github.com/LEW21/siren
```

Siren will be compiled as a single static binary called `siren`, saved in the `siren-build/bin/` directory. You can copy it wherever you want.

## License
MIT license.

## Requirements
* Linux 4.0
* systemd 220
* NO btrfs on /var/lib/image-layers (Non-empty directory removal is bugged on overlayfs over btrfs. Use a good fs instead - eg. ext4.)
* sysctl net.ipv4.ip_forward=1 (I don't know if it's always required, but on my local PC containers can't access internet without it)

## Previous work
This is a rewrite of [siren-sh](https://github.com/LEW21/siren-sh).
