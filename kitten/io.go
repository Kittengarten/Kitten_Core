package kitten

import (
	"bytes"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"go.uber.org/zap"

	"github.com/wdvxdr1123/ZeroBot/message"
)

// FilePath 文件路径构建
func FilePath[T string | Path](elem ...T) Path {
	var (
		l = len(elem)
		s = make([]string, l, l)
	)
	for k := range elem {
		s[k] = string(elem[k])
	}
	return Path(filepath.Join(s...))
}

// Read 文件读取
func (path Path) Read() ([]byte, error) {
	return os.ReadFile(path.String())
}

/*
Write 文件写入

如文件不存在会尝试新建
*/
func (path Path) Write(data []byte) (err error) {
	// 检查文件夹是否存在，不存在则创建
	if err = os.MkdirAll(filepath.Dir(path.String()), 0755); nil != err {
		return
	}
	// 写入文件
	return os.WriteFile(path.String(), data, 0644)
}

// Exists 判断文件或文件夹是否存在
func (path Path) Exists() (bool, error) {
	_, err := os.Stat(path.String())
	if nil == err {
		// 文件或文件夹存在
		return true, nil
	}
	if errors.Is(err, fs.ErrNotExist) {
		// 文件或文件夹不存在
		return false, nil
	}
	// 其它错误
	return false, err
}

// （私有）判断路径是否文件夹
func (path Path) isDir() (bool, error) {
	info, err := os.Stat(path.String())
	return info.IsDir(), err
}

// LoadPath 加载文件中保存的相对路径或绝对路径
func (path Path) LoadPath() (Path, error) {
	data, err := os.ReadFile(path.String())
	if nil != err {
		return path, err
	}
	if data = bytes.TrimSpace(data); filepath.IsAbs(string(data)) {
		return Path(`file://`) + FilePath(Path(data)), nil
	}
	return FilePath(Path(data)), nil
}

// GetImage 从图片的相对/绝对路径，或相对/绝对路径文件中保存的相对/绝对路径加载图片
func (path Path) GetImage(name Path) (message.MessageSegment, error) {
	if filepath.IsAbs(path.String()) {
		isDir, err := path.isDir()
		if nil != err {
			return message.MessageSegment{}, err
		}
		if isDir {
			return message.Image(fmt.Sprint(`file://`, FilePath(path, name))), nil
		}
		p, err := path.LoadPath()
		return message.Image(fmt.Sprint(`file://`, FilePath(p, name))), err
	}
	if isDir, err := path.isDir(); isDir {
		return message.Image(FilePath(path, name).String()), err
	}
	p, err := path.LoadPath()
	return message.Image(FilePath(p, name).String()), err
}

// Path 类型实现 Stringer 接口，并将路径规范化
func (path Path) String() string {
	return filepath.Clean(filepath.Join(string(path)))
}

/*
InitFile 初始化文本文件，要求传入路径事先规范化过

如果路径所指向的文件实际位于上级文件夹中，会相应地修改路径
*/
func InitFile[T string | Path](name *T, text string) error {
	var (
		n      = Path(*name)
		e, err = n.Exists()
	)
	// 如果发生错误或文件存在，直接返回
	if e || nil != err {
		return err
	}
	// 如果文件不存在
	if !filepath.IsAbs(n.String()) {
		// 如果不是绝对路径，搜索其上一级路径
		n = FilePath(`..`, n)
		e, err := n.Exists()
		if nil != err {
			return err
		}
		if e {
			// 如果在上一级目录中存在，则将路径修改为上一级
			*name = T(n)
			return nil
		}
	}
	// 如果文件不存在 && (是绝对路径 || 不是绝对路径但在上一级目录不存在)，初始化该文件
	return n.Write([]byte(text))
}

/*
从文件获取路径

d 为默认值
*/
func (p Path) GetPath(d Path) Path {
	return Path(p.GetString(d.String()))
}

/*
从文件获取字符串或路径

d 为默认值
*/
func (p Path) GetString(d string) string {
	if err := InitFile(&p, d); nil != err {
		zap.S().Errorf("初始化文件 %s 失败了喵！\n%v", p, err)
		return d
	}
	f, err := os.ReadFile(p.String())
	if nil != err {
		zap.S().Errorf("打开文件 %s 失败了喵！\n%v", p, err)
		return d
	}
	return string(f)
}
