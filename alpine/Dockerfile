# ------------------------------------------------------------------------
from alpine:3.15

add alpine-minirootfs-3.15.0-x86_64.tar.gz /layer/

workdir /layer

run apk update
run apk add -p . musl lvm2 lvm2-dmeventd udev cryptsetup e2fsprogs btrfs-progs
run rm -rf usr/share/apk var/cache/apk

entrypoint ["sh","-c","find |cpio -H newc -o |base64"]
