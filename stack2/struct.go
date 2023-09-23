package stack2

import (
	"time"
)

const (
	flat    byte = iota // 平地摔
	fall                // 摔下来
	press               // 压坏
	pressed             // 被压坏
)

type (
	// 叠猫猫配置
	config struct {
		DefaultWeight int `yaml:"defaultweight"` // 默认体重（0.1kg 数）
		GapTime       int `yaml:"gaptime"`       // 每千克体重的冷却时间（小时数）
		MinGapTime    int `yaml:"mingaptime"`    // 最小冷却时间（小时数）
	}

	data []meow // 叠猫猫数据

	// 猫猫数据值
	meow struct {
		ID     int64     `yaml:"id"`     // QQ
		Name   string    `yaml:"name"`   // 群名片或昵称
		Weight int       `yaml:"weight"` // 体重（0.1kg 数）
		Status bool      `yaml:"status"` // 是否在叠猫猫中
		Time   time.Time `yaml:"time"`   // 如果在叠猫猫中，叠入的时间；如果不在叠猫猫中，冷却结束的时间
		// Stat             // 统计信息
	}

	Stat struct {
		In        `yaml:"in"`   // 加入次数
		Exit      `yaml:"exit"` // 退出次数
		Time      time.Time     `yaml:"time"`      // 总时长
		Max       int           `yaml:"max"`       // 曾经达到的最大高度
		MaxWeight int           `yaml:"maxweight"` // 曾经达到的最大重量
	}

	In struct {
		Success int  // 成功
		Fall    Fail // 摔下去
		Press   Fail // 压坏
		Flat    int  // 平地摔次数
	}

	Fail struct {
		Count int    // 失败次数
		Max   Record // 单次导致退出猫猫的最大值
		Total Record // 导致退出的猫猫总和
	}

	Record struct {
		Count  int // 最大数量
		Weight int // 最大重量（0.1kg 数）
	}

	Exit struct {
		Fall    int // 摔下去次数
		Pressed int // 被压坏次数
		Active  int // 主动退出次数
	}
)
