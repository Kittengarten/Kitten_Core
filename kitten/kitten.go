// Package kitten 包含了 KittenCore 以及各插件的核心依赖结构体、方法和函数
package kitten

import (
	"fmt"

	"gopkg.in/yaml.v3"
)

const (
	Empty  = `[]`                // YAML 空数组
	Layout = `2006.1.2 15:04:05` // 日期时间格式
	Guild  = `暂不支持频道喵！`
)

var (
	config    Config // 来自 Bot 的配置文件
	imagePath Path   // 图片路径
)

func init() {
	// 配置文件加载
	d, err := Path(`config.yaml`).Read()
	if nil != err {
		fmt.Println(err)
		return
	}
	if err := yaml.Unmarshal(d, &config); nil != err {
		fmt.Println(err)
		return
	}
	// 图片路径
	imagePath = FilePath(config.Path, `image`)
}

// GetImagePath 获取图片路径
func GetImagePath() Path {
	return imagePath
}

// GetMainConfig 获取主配置
func GetMainConfig() Config {
	return config
}
