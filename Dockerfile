# ------------------------------------------------------------------------
from mcluseau/golang-builder:1.15.5 as build

# ------------------------------------------------------------------------
from alpine:3.12

env busybox_v=1.28.1-defconfig-multiarch \
    arch=x86_64

run apk add --update curl

workdir /layer

add build-layer /
run /build-layer

copy --from=build /go/bin/initrd /layer/init

entrypoint ["sh","-c","find |cpio -H newc -o |base64"]
