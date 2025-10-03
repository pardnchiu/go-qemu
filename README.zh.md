# Go Qemu - Proxmox VM API

> Go Qemu 是基於 Go 語言開發的 Proxmox VE 虛擬機管理 API 服務，**自動下載官方映像檔並完成端到端部署**。支援 Debian、Ubuntu、RockyLinux 的**一鍵式自動化部署**。


[![pkg](https://pkg.go.dev/badge/github.com/pardnchiu/go-qemu.svg)](https://pkg.go.dev/github.com/pardnchiu/go-qemu)
[![version](https://img.shields.io/github/v/tag/pardnchiu/go-qemu?label=release)](https://github.com/pardnchiu/go-qemu/releases)
[![license](https://img.shields.io/github/license/pardnchiu/go-qemu)](LICENSE)<br>
[![readme](https://img.shields.io/badge/readme-EN-white)](README.md)
[![readme](https://img.shields.io/badge/readme-ZH-white)](README.zh.md)

## 核心特色

### 完整虛擬機生命週期管理
- 支援多種 Linux 發行版本（**官方映像檔自動下載與部署**）
- 即時 SSE 串流安裝進度回饋
- 智能 IP 地址分配與管理
- 完整的 SSH 金鑰配置

### 多節點叢集支援
- 支援 Proxmox VE 叢集環境
- 遠端節點 SSH 操作支援
- 多節點統一請求 API

### 零配置部署
- 自動從官方源下載最新的雲端映像檔，無需手動準備映像檔
  - Debian: 從 `cloud.debian.org` 下載
  - RockyLinux: 從 `dl.rockylinux.org` 下載
  - Ubuntu: 從 `cloud-images.ubuntu.com` 下載  

## 如何使用

### 健康檢查
```
GET /api/health
```

### 虛擬機管理

#### 創建虛擬機
> 支援 OS
> - Debian: 11, 12, 13
> - RockyLinux: 8, 9, 10
> - Ubuntu: 20.04, 22.04, 24.04  
```
POST /api/vm/install
```

```json
{
  "id": 101,                                          // 可選，VM ID，不指定則自動分配
  "name": "test-vm",                                  // VM 名稱
  "node": "pve1",                                     // 可選，指定節點名稱
  "os": "ubuntu",                                     // 必填，支援: debian, ubuntu, rockylinux
  "version": "22.04",                                 // 必填，作業系統版本
  "cpu": 2,                                           // vCPU 核心數，預設 2
  "ram": 2048,                                        // RAM 大小 (MB)，預設 2048
  "disk": "20G",                                      // 硬碟大小，預設 16G
  "passwd": "password123",                            // SSH 密碼，預設 "passwd"
  "pubkey": "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5..."     // 可選，SSH 公鑰
}
```

**SSE 回應**
```javascript
data: {"step":"preparation > checking VMID","status":"processing","message":"[*] using specified VMID: 101"}
data: {"step":"VM creation > creating VM","status":"success","message":"[+] VM created successfully (2.45s)"}
```

#### 啟動虛擬機
```
POST /api/vm/{id}/start
```

**SSE 回應**
```javascript
data: {"step":"starting VM","status":"success","message":"[+] VM starting (1.23s)"}
data: {"step":"waiting for SSH","status":"success","message":"[+] VM is ready (15.67s)"}
```

#### 重啟虛擬機
```
POST /api/vm/{id}/reboot
```

**SSE 回應**
```javascript
data: {"step":"rebooting VM","status":"success","message":"[+] VM rebooting (1.23s)"}
data: {"step":"waiting for SSH","status":"success","message":"[+] VM is ready (15.67s)"}
```

#### 虛擬機狀態
```
GET /api/vm/{id}/status
```

#### 關閉虛擬機
```
POST /api/vm/{id}/shutdown
```

#### 停止虛擬機（強制關機）
```
POST /api/vm/{id}/stop
```

#### 移除虛擬機
```
POST /api/vm/{id}/destroy
```

#### 調整虛擬機 CPU
```
POST /api/vm/{id}/set/cpu
```

```json
{
  "cpu": 4
}
```

#### 調整虛擬機記憶體
```
POST /api/vm/{id}/set/memory
```

```json
{
  "memory": 4096
}
```

#### 擴充虛擬機硬碟
```
POST /api/vm/{id}/set/disk
```

```json
{
  "disk": "10G"
}
```

#### 遷移虛擬機節點
```
POST /api/vm/{id}/set/node
```

```json
{
  "node": "pve2"
}
```

**SSE 回應**
```javascript
data: {"step":"migrating","status":"processing","message":"..."}
data: {"step":"migrating","status":"success","message":"[+] migration completed"}
```

## 環境配置

### 必要環境變數
```bash
PORT=8080

# 主要節點名稱
MAIN_NODE=PVE1

# 網路閘道
GATEWAY=10.0.0.1

# 允許存取的 IP（逗號分隔，0.0.0.0 允許所有）
ALLOW_IPS=0.0.0.0

# 節點 IP 配置
NODE_PVE1=10.0.0.2
NODE_PVE2=10.0.0.3
NODE_PVE3=10.0.0.4

# IP 分配範圍與指定儲存空間
ASSIGN_IP_START=100
ASSIGN_IP_END=254
ASSIGN_STORAGE=storage

# VM 資源限制
VM_MAX_CPU=8
VM_MAX_DISK=32
VM_MAX_RAM=16384
VM_BALLOON_MIN=2048

# VM root 密碼
# 預設為 0123456789（無密碼）
VM_ROOT_PASSWORD=
```

### 初始化腳本

系統會自動執行對應的初始化腳本：
```
sh/
├── debian_11.sh
├── debian_12.sh  
├── debian_13.sh
├── ubuntu_20.04.sh
├── ubuntu_22.04.sh
├── ubuntu_24.04.sh
├── rockylinux_8.sh
├── rockylinux_9.sh
└── rockylinux_10.sh
```

## 授權條款

此原始碼專案採用 [MIT](LICENSE) 授權條款。

## 作者

<img src="https://avatars.githubusercontent.com/u/25631760" align="left" width="96" height="96" style="margin-right: 0.5rem;">

<h4 style="padding-top: 0">邱敬幃 Pardn Chiu</h4>

<a href="mailto:dev@pardn.io" target="_blank">
  <img src="https://pardn.io/image/email.svg" width="48" height="48">
</a> <a href="https://linkedin.com/in/pardnchiu" target="_blank">
  <img src="https://pardn.io/image/linkedin.svg" width="48" height="48">
</a>

***

©️ 2025 [邱敬幃 Pardn Chiu](https://pardn.io)
