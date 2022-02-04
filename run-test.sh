#! /bin/sh

exec qemu-system-x86_64 -pidfile qemu.pid -kernel test-kernel -initrd test-initrd.cpio \
    -smp 2 -m 2048 \
    -nographic -serial mon:stdio -append 'console=ttyS0'
#    -display curses
