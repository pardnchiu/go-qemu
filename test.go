// * LLM 生成，用來檢查 API 用
package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const baseURL = "http://localhost:8080"

type VMInstallRequest struct {
	ID      int    `json:"id,omitempty"`
	NODE    string `json:"node,omitempty"`
	OS      string `json:"os"`
	Version string `json:"version"`
	Name    string `json:"name"`
	CPU     int    `json:"cpu"`
	Disk    string `json:"disk"`
	RAM     int    `json:"ram"`
}

type VMSetRequest struct {
	CPU    *int    `json:"cpu,omitempty"`
	Memory *int    `json:"memory,omitempty"`
	Disk   *string `json:"disk,omitempty"`
	Node   *string `json:"node,omitempty"`
}

func healthCheck() error {
	resp, err := http.Get(baseURL + "/api/health")
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	fmt.Println(string(body))
	return nil
}

func listVMs() error {
	resp, err := http.Get(baseURL + "/api/vm/list")
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	fmt.Println(string(body))
	return nil
}

func getVMStatus(vmID int) error {
	resp, err := http.Get(fmt.Sprintf("%s/api/vm/%d/status", baseURL, vmID))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	fmt.Println(string(body))
	return nil
}

func getInitScript(os string, version string) error {
	resp, err := http.Get(fmt.Sprintf("%s/sh/%s_%s.sh", baseURL, os, version))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	fmt.Println(string(body))
	return nil
}

func installVM(req VMInstallRequest) error {
	data, _ := json.Marshal(req)
	client := &http.Client{}
	httpReq, _ := http.NewRequest("POST", baseURL+"/api/vm/install", bytes.NewBuffer(data))
	httpReq.Header.Set("Accept", "text/event-stream")
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(httpReq)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	buf := make([]byte, 1024)
	for {
		n, err := resp.Body.Read(buf)
		if n > 0 {
			fmt.Print(string(buf[:n]))
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
	}
	return nil
}

func startVM(vmID int) error {
	client := &http.Client{}
	httpReq, _ := http.NewRequest("POST", fmt.Sprintf("%s/api/vm/%d/start", baseURL, vmID), nil)
	httpReq.Header.Set("Accept", "text/event-stream")
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(httpReq)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	buf := make([]byte, 1024)
	for {
		n, err := resp.Body.Read(buf)
		if n > 0 {
			fmt.Print(string(buf[:n]))
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
	}
	return nil
}

func stopVM(vmID int) error {
	httpReq, _ := http.NewRequest("POST", fmt.Sprintf("%s/api/vm/%d/stop", baseURL, vmID), nil)
	httpReq.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(httpReq)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	fmt.Println(string(body))
	return nil
}

func shutdownVM(vmID int) error {
	httpReq, _ := http.NewRequest("POST", fmt.Sprintf("%s/api/vm/%d/shutdown", baseURL, vmID), nil)
	httpReq.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(httpReq)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	fmt.Println(string(body))
	return nil
}

func rebootVM(vmID int) error {
	client := &http.Client{}
	httpReq, _ := http.NewRequest("POST", fmt.Sprintf("%s/api/vm/%d/reboot", baseURL, vmID), nil)
	httpReq.Header.Set("Accept", "text/event-stream")
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(httpReq)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	buf := make([]byte, 1024)
	for {
		n, err := resp.Body.Read(buf)
		if n > 0 {
			fmt.Print(string(buf[:n]))
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
	}
	return nil
}

func destroyVM(vmID int) error {
	httpReq, _ := http.NewRequest("POST", fmt.Sprintf("%s/api/vm/%d/destroy", baseURL, vmID), nil)
	httpReq.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(httpReq)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	fmt.Println(string(body))
	return nil
}

func setVMCPU(vmID int, cpu int) error {
	data := map[string]int{"cpu": cpu}
	body, _ := json.Marshal(data)

	httpReq, _ := http.NewRequest("POST", fmt.Sprintf("%s/api/vm/%d/set/cpu", baseURL, vmID), bytes.NewBuffer(body))
	httpReq.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(httpReq)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	respBody, _ := io.ReadAll(resp.Body)
	fmt.Println(string(respBody))
	return nil
}

func setVMMemory(vmID int, memory int) error {
	data := map[string]int{"memory": memory}
	body, _ := json.Marshal(data)

	httpReq, _ := http.NewRequest("POST", fmt.Sprintf("%s/api/vm/%d/set/memory", baseURL, vmID), bytes.NewBuffer(body))
	httpReq.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(httpReq)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	respBody, _ := io.ReadAll(resp.Body)
	fmt.Println(string(respBody))
	return nil
}

func setVMDisk(vmID int, disk string) error {
	data := map[string]string{"disk": disk}
	body, _ := json.Marshal(data)

	httpReq, _ := http.NewRequest("POST", fmt.Sprintf("%s/api/vm/%d/set/disk", baseURL, vmID), bytes.NewBuffer(body))
	httpReq.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(httpReq)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	respBody, _ := io.ReadAll(resp.Body)
	fmt.Println(string(respBody))
	return nil
}

func setVMNode(vmID int, node string) error {
	data := map[string]string{"node": node}
	body, _ := json.Marshal(data)

	client := &http.Client{}
	httpReq, _ := http.NewRequest("POST", fmt.Sprintf("%s/api/vm/%d/set/node", baseURL, vmID), bytes.NewBuffer(body))
	httpReq.Header.Set("Accept", "text/event-stream")
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(httpReq)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	buf := make([]byte, 1024)
	for {
		n, err := resp.Body.Read(buf)
		if n > 0 {
			fmt.Print(string(buf[:n]))
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
	}
	return nil
}

func main() {
	vmID := 190

	fmt.Println("=== 1. 安裝虛擬機 ===")
	if err := installVM(VMInstallRequest{
		ID: vmID,
		// NODE:    "pve1",
		OS:      "debian",
		Version: "13",
		Name:    "test-vm",
		CPU:     2,
		Disk:    "16G",
		RAM:     2048,
	}); err != nil {
		fmt.Printf("安裝失敗: %v\n", err)
		return
	}
	time.Sleep(5 * time.Second)

	fmt.Println("\n=== 2. 檢查虛擬機狀態 ===")
	if err := getVMStatus(vmID); err != nil {
		fmt.Printf("檢查狀態失敗: %v\n", err)
	}
	time.Sleep(2 * time.Second)

	fmt.Println("\n=== 3. 重啟虛擬機 ===")
	if err := rebootVM(vmID); err != nil {
		fmt.Printf("重啟失敗: %v\n", err)
		return
	}
	time.Sleep(5 * time.Second)

	fmt.Println("\n=== 4. 關機 (shutdown) ===")
	if err := shutdownVM(vmID); err != nil {
		fmt.Printf("關機失敗: %v\n", err)
		return
	}
	time.Sleep(20 * time.Second)

	fmt.Println("\n=== 5. 設定 CPU ===")
	if err := setVMCPU(vmID, 4); err != nil {
		fmt.Printf("設定 CPU 失敗: %v\n", err)
	}
	time.Sleep(1 * time.Second)

	fmt.Println("\n=== 6. 設定記憶體 ===")
	if err := setVMMemory(vmID, 4096); err != nil {
		fmt.Printf("設定記憶體失敗: %v\n", err)
	}
	time.Sleep(5 * time.Second)

	fmt.Println("\n=== 7. 設定硬碟 ===")
	if err := setVMDisk(vmID, "32G"); err != nil {
		fmt.Printf("設定硬碟失敗: %v\n", err)
	}
	time.Sleep(5 * time.Second)

	fmt.Println("\n=== 8. 設定節點 ===")
	if err := setVMNode(vmID, "r230"); err != nil {
		fmt.Printf("設定節點失敗: %v\n", err)
	}
	time.Sleep(5 * time.Second)

	fmt.Println("\n=== 9. 開機 ===")
	if err := startVM(vmID); err != nil {
		fmt.Printf("開機失敗: %v\n", err)
		return
	}
	time.Sleep(5 * time.Second)

	fmt.Println("\n=== 10. 關機 (stop) ===")
	if err := stopVM(vmID); err != nil {
		fmt.Printf("強制關機失敗: %v\n", err)
		return
	}
	time.Sleep(20 * time.Second)

	fmt.Println("\n=== 11. 摧毀虛擬機 ===")
	if err := destroyVM(vmID); err != nil {
		fmt.Printf("摧毀失敗: %v\n", err)
		return
	}

	fmt.Println("\n=== 測試完成 ===")
}
