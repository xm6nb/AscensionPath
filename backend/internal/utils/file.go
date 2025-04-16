package utils

import (
	"archive/zip"
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// ReadJson 将 JSON 文件读取到结构体中
func ReadJson[T any](filePath string, s *T) error {
	file, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}

	err = json.Unmarshal(file, s)
	if err != nil {
		return err
	}

	return nil
}

// SaveFile 保存文件到指定路径
func SaveFile(fileData []byte, fPath string) error {
	// 检查文件是否已存在
	if _, err := os.Stat(fPath); err == nil {
		return fmt.Errorf("文件已存在: %s", filepath.Base(fPath))
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("文件状态检查失败: %v", err)
	}

	// 创建目录如果不存在
	dir := filepath.Dir(fPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("创建目录失败: %v", err)
	}

	// 写入文件
	if err := os.WriteFile(fPath, fileData, 0644); err != nil {
		return fmt.Errorf("写入文件失败: %v", err)
	}

	return nil
}

// 新增解压函数
func Unzip(data []byte, spath string) error {
	// 计算数据的MD5哈希作为唯一标识
	hash := md5.Sum(data)
	hashStr := hex.EncodeToString(hash[:])

	// 创建标记文件路径
	markerFile := filepath.Join(spath, ".unzip_"+hashStr)

	// 检查标记文件是否存在
	if _, err := os.Stat(markerFile); err == nil {
		return fmt.Errorf("已解压过相同内容") // 已解压过相同内容，直接返回
	}

	// 创建内存中的zip reader
	r, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return fmt.Errorf("创建zip reader失败: %w", err)
	}

	// 创建目标目录
	if err := os.MkdirAll(spath, 0755); err != nil {
		return fmt.Errorf("创建目录失败: %w", err)
	}

	// 遍历zip文件
	for _, f := range r.File {
		// 构建安全路径
		targetPath := filepath.Join(spath, f.Name)

		// 防止路径穿越攻击
		if !strings.HasPrefix(targetPath, filepath.Clean(spath)+string(os.PathSeparator)) {
			return fmt.Errorf("非法文件路径: %s", f.Name)
		}

		// 创建目录结构
		if f.FileInfo().IsDir() {
			if err := os.MkdirAll(targetPath, f.Mode()); err != nil {
				return fmt.Errorf("创建目录失败: %w", err)
			}
			continue
		}

		// 检查文件是否已存在
		if _, err := os.Stat(targetPath); err == nil {
			continue // 文件已存在，跳过
		}

		// 创建文件
		if err := os.MkdirAll(filepath.Dir(targetPath), 0755); err != nil {
			return fmt.Errorf("创建父目录失败: %w", err)
		}

		// 写入文件内容
		dstFile, err := os.OpenFile(targetPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			return fmt.Errorf("创建文件失败: %w", err)
		}
		defer dstFile.Close()

		srcFile, err := f.Open()
		if err != nil {
			return fmt.Errorf("打开压缩文件失败: %w", err)
		}
		defer srcFile.Close()

		if _, err := io.Copy(dstFile, srcFile); err != nil {
			return fmt.Errorf("写入文件内容失败: %w", err)
		}
	}

	// 创建标记文件
	if err := os.WriteFile(markerFile, []byte{}, 0644); err != nil {
		return fmt.Errorf("创建标记文件失败: %w", err)
	}

	return nil
}

// IsSafePath 检查文件路径是否在指定基础目录下
func IsSafePath(baseDir string, targetPath string) bool {
	// 获取绝对路径
	absBase, err := filepath.Abs(baseDir)
	if err != nil {
		return false
	}

	absTarget, err := filepath.Abs(targetPath)
	if err != nil {
		return false
	}

	// 清理路径并添加路径分隔符
	cleanBase := filepath.Clean(absBase) + string(os.PathSeparator)
	cleanTarget := filepath.Clean(absTarget)

	// 检查目标路径是否在基础目录下
	return strings.HasPrefix(cleanTarget, cleanBase)
}

// IsPathExist 检查绝对路径是否存在
func IsPathExist(targetPath string) bool {
	// 获取绝对路径
	absPath, err := filepath.Abs(targetPath)
	if err != nil {
		return false
	}

	// 清理路径
	cleanPath := filepath.Clean(absPath)

	// 检查路径是否存在
	if _, err := os.Stat(cleanPath); err != nil {
		return !os.IsNotExist(err)
	}
	return true
}
