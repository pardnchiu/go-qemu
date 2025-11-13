# Go QEMU
> [!NOTE]
> Removed macOS support. Focus on Debian/Ubuntu OS for feature implementation.

## Package Dependencies
- qemu-system
- genisoimage/mkisofs

## Setup Requirements for Sudo Actions
```bash
sudo usermod -aG kvm [user]
```

## Network Setup (網路設定)

### Bridge Configuration Commands
```bash
sudo mkdir -p /etc/qemu
echo "allow all" | sudo tee /etc/qemu/bridge.conf
sudo chmod u+s /usr/lib/qemu/qemu-bridge-helper
sudo apt install bridge-utils
sudo brctl addbr vmbr0
sudo brctl addif vmbr0 eth0
sudo ip link set vmbr0 up
sudo ip addr flush dev eth0
sudo ip addr add [host_ip]/24 dev vmbr0
sudo ip route add default via [router_ip] dev vmbr0
```

### Network Interface Configuration
Add to `/etc/network/interfaces`:
```
auto eth0
iface eth0 inet manual

auto vmbr0
iface vmbr0 inet static
  address [host_ip]
  netmask 255.255.255.0
  gateway [router_ip]
  dns-nameservers 8.8.8.8
  bridge_ports eth0
  bridge_stp off
  bridge_fd 0
  bridge_maxwait 0
```

### Default username
- Debian `debian` 
- Ubuntu `ubuntu`
- CentOS-Stream `centos`
- RockyLinux `rocky`
- AlmaLinux `alma`