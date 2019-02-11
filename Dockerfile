# ------------------------------------------------------------------------
from golang:1.11.5-alpine3.8 as build

add vendor /go/src/init/vendor
add *.go /go/src/init/
workdir /go/src/init

run CGO_ENABLED=0 go build -o /layer/init .

# ------------------------------------------------------------------------
from alpine:3.8

env busybox_v=1.28.1-defconfig-multiarch \
    arch=x86_64

run apk add --update curl

workdir /layer

add build-layer /
run /build-layer

copy --from=build /layer/init /layer/init

entrypoint ["sh","-c","find |cpio -H newc -o |base64"]
