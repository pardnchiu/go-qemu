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
sudo mv /usr/bin/passwd /usr/bin/.passwd
sudo mv /usr/bin/chpasswd /usr/bin/.chpasswd 2>/dev/null || true
sudo tee /usr/bin/passwd > /dev/null << 'EOF'
#!/bin/bash
echo "already disabled by admin"
exit 1
EOF
sudo chmod +x /usr/bin/passwd
sudo tee /usr/bin/chpasswd > /dev/null << 'EOF'
#!/bin/bash
echo "already disabled by admin"
exit 1
EOF
sudo chmod +x /usr/bin/chpasswd

echo "waiting for package manager to be available"
wait_for_apt() {
  while sudo fuser /var/lib/dpkg/lock-frontend >/dev/null 2>&1 || \
        sudo fuser /var/lib/apt/lists/lock >/dev/null 2>&1 || \
        sudo fuser /var/cache/apt/archives/lock >/dev/null 2>&1; do
    sleep 1
  done
}
wait_for_apt

sudo systemctl stop unattended-upgrades 2>/dev/null || true
sudo systemctl disable unattended-upgrades 2>/dev/null || true
sudo killall apt apt-get dpkg 2>/dev/null || true

# 設定套件源
echo "editing sources.list"
cat << 'EOF' | sudo tee /etc/apt/sources.list > /dev/null
deb http://tw.archive.ubuntu.com/ubuntu/ focal main restricted universe multiverse
deb http://tw.archive.ubuntu.com/ubuntu/ focal-updates main restricted universe multiverse
deb http://tw.archive.ubuntu.com/ubuntu/ focal-backports main restricted universe multiverse
deb http://security.ubuntu.com/ubuntu/ focal-security main restricted universe multiverse
EOF

# 更新套件清單
echo "updateing packages"
sudo apt-get update

# 升級系統
echo "upgrading packages"
sudo apt-get upgrade -y

# 安裝基本套件
echo "installing basic packages"
sudo apt-get install language-pack-en ufw rsync wget curl tar zip unzip tree make nano vim git screen tmux sysbench stress-ng htop iotop bc fio qemu-guest-agent --fix-missing -y

# 啟用 QEMU Guest Agent
echo "enable qemu-guest-agent"
sudo systemctl enable qemu-guest-agent

# 設定語系
echo "setting locale to en_US.UTF-8"
sudo locale-gen en_US.UTF-8
sudo localectl set-locale LANG=en_US.UTF-8

# 設定時區
echo "setting timezone to Asia/Taipei"
sudo timedatectl set-timezone Asia/Taipei

# 設定系統參數
echo "setting sysctl"
echo "net.ipv4.icmp_echo_ignore_all = 1" | sudo tee -a /etc/sysctl.conf > /dev/null
echo "fs.inotify.max_user_watches=524288" | sudo tee -a /etc/sysctl.conf > /dev/null
sudo sysctl -p || true

# 更新 GRUB 設定
echo "updating GRUB configuration"
sudo sed -i -e "s/GRUB_CMDLINE_LINUX=\"/GRUB_CMDLINE_LINUX=\"ipv6.disable=0 /g" /etc/default/grub
sudo update-grub

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
sudo curl -f -o /usr/local/bin/sysinfo https://gist.githubusercontent.com/pardnchiu/561ef0581911eac7aed33c898a1a2b21/raw/f0e2524de2328448a781863f7800d6f8bdfbd2f4/sysinfo
sudo chmod +x /usr/local/bin/sysinfo
sudo cp -f /usr/local/bin/sysinfo /etc/profile.d/ssh-motd.sh
sudo chmod +x /etc/profile.d/ssh-motd.sh

# 有需要再自己去 sshd_config 裡面改
echo "#PasswordAuthentication no" | sudo tee /etc/ssh/sshd_config.d/60-cloudimg-settings.conf
sudo systemctl restart ssh

echo "HISTCONTROL=ignorespace" >> ~/.bashrc
source ~/.bashrc

echo "cleaning up"
sudo apt-get clean
sudo apt-get autoclean 
sudo apt-get autoremove -y
sudo rm -rf /tmp/*
sudo rm -rf /var/tmp/*
sudo truncate -s 0 /var/log/*log

echo '' > .bash_history && history -c