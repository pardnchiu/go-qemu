#cloud-config>user-data
users:
  - name: [USER]
    sudo: ALL=(ALL) NOPASSWD:ALL
    ssh_authorized_keys:
      - [SSH_KEYS]
    shell: /bin/bash

ssh_pwauth: true

chpasswd:
  list: |
    [USER]:[PASSWORD]
  expire: false

runcmd:
  # Clean up cloud-init instance data to allow re-use on next boot
  - [ sh, -c, '(sleep 3 && rm -rf /var/lib/cloud/instance /var/lib/cloud/instances/*) &' ]





#cloud-config>meta-data
instance-id: %s
local-hostname: %s





# Command to create the cloud-init ISO
mkisofs -output cloud-init.iso -volid cidata -joliet -rock user-data meta-data