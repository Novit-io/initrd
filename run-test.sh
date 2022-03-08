#! /bin/sh

disk1=tmp/test-vda.qcow2
disk2=tmp/test-vdb.qcow2

if ! [ -e $disk1 ]; then
    qemu-img create -f qcow2 $disk1 10G
fi
if ! [ -e $disk2 ]; then
    qemu-img create -f qcow2 $disk2 10G
fi

exec qemu-system-x86_64 -pidfile qemu.pid -kernel test-kernel -initrd test-initrd.cpio \
    -smp 2 -m 2048 \
    -netdev bridge,br=novit,id=eth0 -device virtio-net-pci,netdev=eth0 \
    -drive file=$disk1,if=virtio \
    -drive file=$disk2,if=virtio \
    -nographic -serial mon:stdio -append 'console=ttyS0'
#    -display curses
