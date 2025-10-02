#!/bin/bash
set -e

# 參數處理
ROOT_PASSWORD="${1:-0123456789}"
echo "change root password"
echo "root:${ROOT_PASSWORD}" | sudo chpasswd

# 禁止修改密碼
echo "disable password change"
sudo chage -M -1 $(whoami)
sudo chage -E -1 $(whoami)
sudo chage -M -1 root
sudo chage -E -1 root

# 備份並移除密碼修改工具
echo "remove passwd"
mv /usr/bin/passwd /usr/bin/passwd.bak
tee /usr/bin/passwd > /dev/null << 'EOF'
#!/bin/bash
echo "already disabled by admin"
exit 1
EOF
chmod +x /usr/bin/passwd

echo "waiting for package manager to be available"
wait_for_dnf() {
    while sudo fuser /var/lib/rpm/.rpm.lock >/dev/null 2>&1 || \
          sudo fuser /var/cache/dnf/metadata_lock.pid >/dev/null 2>&1; do
        sleep 1
    done
}
wait_for_dnf

# 更新系統
echo "upgrading packages"
dnf update -y

# 安裝 EPEL 套件庫
echo "install EPEL repository"
dnf install -y epel-release

# 安裝基本套件
echo "installing basic packages"
dnf install -y wget curl tar zip unzip tree make nano vim git screen tmux htop iotop bc qemu-guest-agent

# 啟用 QEMU Guest Agent
echo "enable qemu-guest-agent"
systemctl enable qemu-guest-agent

# 設定時區
echo "setting timezone to Asia/Taipei"
timedatectl set-timezone Asia/Taipei

# 設定系統參數
echo "setting sysctl"
echo "net.ipv4.icmp_echo_ignore_all = 1" | sudo tee -a /etc/sysctl.conf > /dev/null
echo "fs.inotify.max_user_watches=524288" | sudo tee -a /etc/sysctl.conf > /dev/null
sudo sysctl -p || true

# 更新 GRUB 設定
echo "updating GRUB configuration"
sed -i -e "s/GRUB_CMDLINE_LINUX=\"/GRUB_CMDLINE_LINUX=\"ipv6.disable=0 /g" /etc/default/grub
grub2-mkconfig -o /boot/grub2/grub.cfg

# 建立 swap 檔案
echo "creating swap file"
# 檢查是否已存在 swap 檔案
if [ -f /swapfile ]; then
    sudo swapoff /swapfile 2>/dev/null || true
    sudo rm -f /swapfile
fi

# 建立新的 swap 檔案
sudo fallocate -l 2G /swapfile
sudo chmod 600 /swapfile
sudo mkswap /swapfile
sudo swapon /swapfile

echo '/swapfile none swap sw 0 0' | sudo tee -a /etc/fstab

# 下載並設定系統資訊腳本
echo "setting up script sysinfo"
curl -f -o /usr/local/bin/sysinfo https://gist.githubusercontent.com/pardnchiu/561ef0581911eac7aed33c898a1a2b21/raw/f0e2524de2328448a781863f7800d6f8bdfbd2f4/sysinfo
sudo chmod +x /usr/local/bin/sysinfo
sudo cp -f /usr/local/bin/sysinfo /etc/profile.d/ssh-motd.sh
sudo chmod +x /etc/profile.d/ssh-motd.sh

# 系統清理
echo "cleaning up"
sudo dnf clean all
sudo dnf autoremove -y
sudo rm -rf /tmp/*
sudo rm -rf /var/tmp/*
sudo truncate -s 0 /var/log/*log


echo '' > .bash_history && history -c