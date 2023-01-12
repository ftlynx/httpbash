package main

import (
	"encoding/base64"
	"os"
)

func InSlice[T int | string](v T, list []T) bool {
	for _, v2 := range list {
		if v == v2 {
			return true
		}
	}
	return false
}

// CommandConfigFile 适用于 xx -f xx.json 这种需要远程传输 xx.json 配置文件的命令
type CommandConfigFile struct {
	Base64Content string `json:"base64_content"` // 文件内容  base64
}

func (c *CommandConfigFile) StoreFile(filename string, dirname string) error {
	if c.Base64Content == "" {
		return nil
	}
	contentByte, err := base64.StdEncoding.DecodeString(c.Base64Content)
	if err != nil {
		return err
	}
	// 创建父目录
	if err := os.MkdirAll(dirname, 0755); err != nil {
		return err
	}
	path := dirname + filename
	if err := os.WriteFile(path, contentByte, 0666); err != nil {
		return err
	}

	return nil
}
