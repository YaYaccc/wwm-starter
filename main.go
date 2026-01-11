package main

import (
	"crypto/sha256"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"time"
	"unsafe"
)

const (
	// SourceExe 源文件名
	SourceExe = "yysls.exe"
	// TargetExe 目标文件名
	TargetExe = "wwm.exe"
)

func main() {
	// 0. 检查并请求管理员权限
	if !isAdmin() {
		fmt.Println("当前无管理员权限，正在尝试提升权限...")
		runMeElevated()
		return
	}

	// 1. 获取当前程序所在目录
	exePath, err := os.Executable()
	if err != nil {
		logAndPause(fmt.Sprintf("错误: 无法获取当前程序路径: %v", err))
	}
	baseDir := filepath.Dir(exePath)

	// 2. 构建路径
	// 程序默认放置在 yysls_medium\Engine\Binaries\Win64r 下，直接在当前目录下查找文件
	sourcePath := filepath.Join(baseDir, SourceExe)
	targetPath := filepath.Join(baseDir, TargetExe)

	fmt.Printf("当前工作目录: %s\n", baseDir)

	// 3. 检查源文件是否存在
	if _, err := os.Stat(sourcePath); os.IsNotExist(err) {
		logAndPause(fmt.Sprintf("错误: 源文件不存在: %s", sourcePath))
	}

	// 4. 检查目标文件是否存在及一致性
	targetExists := false
	if _, err := os.Stat(targetPath); err == nil {
		targetExists = true
	}

	if !targetExists {
		fmt.Println("目标文件不存在，正在复制...")
		if err := copyFile(sourcePath, targetPath); err != nil {
			logAndPause(fmt.Sprintf("复制文件失败: %v", err))
		}
		fmt.Println("复制成功。")
	} else {
		fmt.Println("目标文件存在，正在检查一致性...")
		consistent, err := checkConsistency(sourcePath, targetPath)
		if err != nil {
			logAndPause(fmt.Sprintf("检查一致性失败: %v", err))
		}

		if !consistent {
			fmt.Println("文件不一致，正在更新目标文件...")
			// 先尝试删除旧文件
			if err := os.Remove(targetPath); err != nil {
				logAndPause(fmt.Sprintf("删除旧文件失败: %v。请检查权限。", err))
			}
			if err := copyFile(sourcePath, targetPath); err != nil {
				logAndPause(fmt.Sprintf("复制文件失败: %v", err))
			}
			fmt.Println("更新成功。")
		} else {
			fmt.Println("文件一致，无需更新。")
		}
	}

	// 5. 启动目标进程
	fmt.Printf("正在启动 %s...\n", targetPath)
	cmd := exec.Command(targetPath)
	cmd.Dir = baseDir // 通常将工作目录设置为二进制文件所在目录
	// 不绑定 Stdout 和 Stderr，避免输出日志
	// cmd.Stdout = os.Stdout
	// cmd.Stderr = os.Stderr

	if err := cmd.Start(); err != nil {
		logAndPause(fmt.Sprintf("启动应用程序失败: %v。请检查权限。", err))
	}

	fmt.Println("应用程序启动成功。")
	// 等待几秒让用户看到成功信息，然后自动退出
	time.Sleep(2 * time.Second)
}

// isAdmin 检查当前进程是否具有管理员权限
// 通过尝试打开物理磁盘设备来判断（通常只有管理员能打开）
func isAdmin() bool {
	_, err := os.Open("\\\\.\\PHYSICALDRIVE0")
	return err == nil
}

// runMeElevated 尝试以管理员权限重启当前程序
func runMeElevated() {
	verb := "runas"
	exe, _ := os.Executable()
	cwd, _ := os.Getwd()
	args := strings.Join(os.Args[1:], " ")

	verbPtr, _ := syscall.UTF16PtrFromString(verb)
	exePtr, _ := syscall.UTF16PtrFromString(exe)
	cwdPtr, _ := syscall.UTF16PtrFromString(cwd)
	argsPtr, _ := syscall.UTF16PtrFromString(args)

	shell32 := syscall.NewLazyDLL("shell32.dll")
	procShellExecuteW := shell32.NewProc("ShellExecuteW")
	_, _, _ = procShellExecuteW.Call(
		0,
		uintptr(unsafe.Pointer(verbPtr)),
		uintptr(unsafe.Pointer(exePtr)),
		uintptr(unsafe.Pointer(argsPtr)),
		uintptr(unsafe.Pointer(cwdPtr)),
		1, // SW_NORMAL
	)
}

// checkConsistency 通过 SHA256 哈希比较文件
// 确保目标文件与源文件完全一致
func checkConsistency(src, dst string) (bool, error) {
	srcHash, err := getFileHash(src)
	if err != nil {
		return false, err
	}
	dstHash, err := getFileHash(dst)
	if err != nil {
		return false, err
	}
	return srcHash == dstHash, nil
}

// getFileHash 计算文件的 SHA256 哈希值
func getFileHash(filePath string) (string, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}

	return fmt.Sprintf("%x", h.Sum(nil)), nil
}

// copyFile 复制文件内容
func copyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, sourceFile)
	return err
}

// logAndPause 打印错误信息并暂停，等待用户确认后退出
func logAndPause(msg string) {
	log.Println(msg)
	fmt.Println("按回车键退出...")
	fmt.Scanln()
	os.Exit(1)
}
