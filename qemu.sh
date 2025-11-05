# Qemu on mac
qemu-system-aarch64 \\
  -accel hvf \\
  -m 2048 \\
  -smp 2,sockets=1,cores=2,threads=1 \\
  -cpu host \\
  -M virt \\
  -bios /opt/homebrew/share/qemu/edk2-aarch64-code.fd \\
  -drive file=100-0.qcow2,format=qcow2,if=virtio \\
  -drive file=100-cloud-init.iso,format=raw,media=cdrom,readonly=on \\
  -netdev user,id=net0,hostfwd=tcp::22-:22 \\
  -device virtio-net-pci,netdev=net0,mac=[MAC] \\
  -device intel-hda \\
  -device hda-duplex \\
  -device qemu-xhci \\
  -rtc base=localtime,clock=host \\
  -vnc 127.0.0.1:0,password=on \\
  -monitor unix:100-monitor.sock,server,nowait \\
  -serial unix:vm-100-serial.sock,server,nowait \\
  -smbios type=1,uuid=[UUID] \\
  -device virtio-gpu-pci \\
  -display none \\
  -nographic

#cloud-config>user-data
users:
  - name: [HOSTNAME]
    sudo: ALL=(ALL) NOPASSWD:ALL
    ssh_authorized_keys:
      - [SSH_KEYS]
    shell: /bin/bash

ssh_pwauth: false

chpasswd:
  list: |
    [USER]:[PASSWORD]
  expire: false

runcmd:
  # Clean up cloud-init instance data to allow re-use on next boot
  - [ sh, -c, '(sleep 3 && rm -rf /var/lib/cloud/instance /var/lib/cloud/instances/*) &' ]

#cloud-config>meta-data
instance-id: [UUID]
local-hostname: [HOSTNAME]

# Command to create the cloud-init ISO
mkisofs \\
  -output cloud-init.iso \\
  -volid cidata \\
  -joliet \\
  -rock user-data meta-data