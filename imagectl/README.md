# imagectl - image manager for systemd-machined

imagectl is a image manager, built for managing machined's images (the /var/lib/machines directory).

## Features
* Seamless integration with systemd
 * machinectl start/stop and all other machine commands (but NOT image / image transfer commands)
 * systemctl start/stop/... systemd-nspawn@...service
 * Without any changes in machined and systemd-nspawn@.service.
* Layered images (using overlayfs)
 * Instant creation of new images using existing ones as a base
* Tagging

## Usage
```
Usage: imagectl COMMAND [arg...]
       imagectl [ -h | --help | -v | --version ]

Image manager for systemd-machined.

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
$ mkdir imagectl-build
$ GOPATH=`pwd`/imagectl-build go get github.com/LEW21/siren/imagectl/cmd/imagectl
```

Imagectl will be compiled as a single static binary called `imagectl`, saved in the `imagectl-build/bin/` directory. You can copy it wherever you want.

## License
MIT license.

## Requirements
* Linux 4.0
* NO btrfs on /var/lib/image-layers (Non-empty directory removal is bugged on overlayfs over btrfs. Use a good fs instead - eg. ext4.)
