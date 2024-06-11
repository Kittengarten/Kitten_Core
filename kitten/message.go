package kitten

import (
	"fmt"

	"github.com/wdvxdr1123/ZeroBot/message"
)

// 检查接口切片的每个元素中是否为错误，如果有则记录日志
func checkErr(v []any) {
	for _, i := range v {
		if err, ok := i.(error); ok {
			Error(err)
		}
	}
}

/*
Text 构建 message.MessageSegment 文本

格式同 fmt.Sprint
*/
func Text(text ...any) message.MessageSegment {
	checkErr(text)
	return message.Text(text...)
}

/*
TextOf 格式化构建 message.MessageSegment 文本

格式同 fmt.Sprintf
*/
func TextOf(format string, a ...any) message.MessageSegment {
	checkErr(a)
	return message.Text(fmt.Sprintf(format, a...))
}

// Image 将收到的图片文件名 | 绝对路径 | 网络 URL | Base64 编码转换为图片消息
func Image(file string, summary ...any) message.MessageSegment {
	return message.Image(file, summary...)
}
