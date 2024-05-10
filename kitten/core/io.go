package core

import (
	"bytes"
	"errors"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/wdvxdr1123/ZeroBot/message"
	"go.uber.org/zap"
)

type Path string // Path 是一个表示文件路径的字符串

// FilePath 文件路径构建
func FilePath[T ~string](elem ...T) Path {
	s := make([]string, len(elem), len(elem))
	for i, e := range elem {
		s[i] = string(e)
	}
	return Path(filepath.Join(s...))
}

// Read 文件读取
func (p Path) Read() ([]byte, error) {
	return os.ReadFile(p.String())
}

/*
Write 文件写入

如文件不存在会尝试新建
*/
func (p Path) Write(data []byte) (err error) {
	// 检查文件夹是否存在，不存在则创建
	if err = os.MkdirAll(filepath.Dir(p.String()), 0o755); nil != err {
		return
	}
	// 写入文件
	return os.WriteFile(p.String(), data, 0o644)
}

// Exists 判断文件或文件夹是否存在
func (p Path) Exists() (bool, error) {
	_, err := os.Stat(p.String())
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
func (p Path) isDir() (bool, error) {
	info, err := os.Stat(p.String())
	if nil != err {
		return false, err
	}
	return info.IsDir(), err
}

// LoadPath 加载文件中保存的相对路径或绝对路径
func (p Path) LoadPath() (Path, error) {
	data, err := p.Read()
	if nil != err {
		return p, err
	}
	if data = bytes.TrimSpace(data); filepath.IsAbs(string(data)) {
		return Path(`file://`) + FilePath(Path(data)), nil
	}
	return FilePath(Path(data)), nil
}

// Image 从图片的相对/绝对路径，或相对/绝对路径文件中保存的相对/绝对路径加载图片
func (p Path) Image(name Path) (message.MessageSegment, error) {
	if filepath.IsAbs(p.String()) {
		isDir, err := p.isDir()
		if nil != err {
			return message.MessageSegment{}, err
		}
		if isDir {
			return message.Image(`file://` + FilePath(p, name).String()), nil
		}
		p, err := p.LoadPath()
		return message.Image(`file://` + FilePath(p, name).String()), err
	}
	if isDir, err := p.isDir(); isDir {
		return message.Image(FilePath(p, name).String()), err
	}
	p, err := p.LoadPath()
	return message.Image(FilePath(p, name).String()), err
}

// String 实现 fmt.Stringer，返回路径规范化后的字符串表示
func (p Path) String() string {
	return filepath.Clean(filepath.Join(string(p)))
}

/*
InitFile 初始化文本文件，要求传入路径事先规范化过

如果路径所指向的文件实际位于上级文件夹中，会相应地修改路径
*/
func InitFile(name *Path, text string) error {
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
			*name = n
			return nil
		}
	}
	// 如果文件不存在 && (是绝对路径 || 不是绝对路径但在上一级目录不存在)，初始化该文件
	return (*name).Write([]byte(text))
}

/*
GetPath 从文件获取路径

d 为默认值
*/
func (p Path) GetPath(d Path) Path {
	return Path(p.GetString(d.String()))
}

/*
GetString 从文件获取字符串或路径

d 为默认值
*/
func (p Path) GetString(d string) string {
	if err := InitFile(&p, d); nil != err {
		zap.S().Errorf(`初始化文件 %s 失败了喵！%s`, p, err)
		return d
	}
	f, err := p.Read()
	if nil != err {
		zap.S().Errorf(`打开文件 %s 失败了喵！%s`, p, err)
		return d
	}
	return string(f)
}

// GetImage 从 url 下载图片到 path
func GetImage(url string, path Path) error {
	// 获取 HTTP 响应体，失败则返回
	d, err := GETData(url)
	if nil != err {
		return err
	}
	return path.Write(d)
}
