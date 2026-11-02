// Package paths 统一应用的数据目录。
package paths

import (
	"os"
	"path/filepath"
)

// appDirName 是 %AppData% 下的目录名。
const appDirName = "win-red-apple"

// ConfigDir 返回应用配置目录并确保它存在。
func ConfigDir() (string, error) {
	base, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	dir := filepath.Join(base, appDirName)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return "", err
	}
	return dir, nil
}

// ProfilePath 返回匿名设备档案的路径。
func ProfilePath() (string, error) {
	dir, err := ConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "anonymous.json"), nil
}
