# ------------------------------------------------------------------------
from alpine:3.15

add alpine/alpine-minirootfs-3.15.0-x86_64.tar.gz /layer/

run apk update
run apk add -p /layer lvm2-static

workdir /layer

run rm -rf usr/share/apk var/cache/apk

#add build-layer /
#run /build-layer

copy dist/init /layer/init

entrypoint ["sh","-c","find |cpio -H newc -o |base64"]
