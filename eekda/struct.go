package eekda

import (
	"time"

	"github.com/Kittengarten/KittenCore/kitten"

	zero "github.com/wdvxdr1123/ZeroBot"
)

const (
	breakfast byte = iota // 早餐
	lunch                 // 午餐
	lowtea                // 下午茶
	dinner                // 晚餐
	supper                // 夜宵
)

type (
	// 今天吃什么
	today struct {
		Time      time.Time `yaml:"time"`      // 生成时间
		Breakfast int64     `yaml:"breakfast"` // 早餐
		Lunch     int64     `yaml:"lunch"`     // 午餐
		LowTea    int64     `yaml:"lowtea"`    // 下午茶
		Dinner    int64     `yaml:"dinner"`    // 晚餐
		Supper    int64     `yaml:"supper"`    // 夜宵
	}

	// Stat 饮食统计数据
	stat []food

	// 猫猫数据
	food struct {
		ID        int64  `yaml:"id"`        // QQ
		Name      string `yaml:"name"`      // 群名片或昵称
		Breakfast int    `yaml:"breakfast"` // 早餐次数
		Lunch     int    `yaml:"lunch"`     // 午餐次数
		LowTea    int    `yaml:"lowtea"`    // 下午茶次数
		Dinner    int    `yaml:"dinner"`    // 晚餐次数
		Supper    int    `yaml:"supper"`    // 夜宵次数
	}
)

// 生成今日的任意一餐
func newFoodToday(ctx *zero.Ctx, td today, meal byte) food {
	switch meal {
	case breakfast:
		// 早餐
		return food{
			ID:        td.Breakfast,
			Name:      getLine(ctx, kitten.QQ(td.Breakfast)),
			Breakfast: 1}
	case lunch:
		// 午餐
		return food{
			ID:    td.Lunch,
			Name:  getLine(ctx, kitten.QQ(td.Lunch)),
			Lunch: 1,
		}
	case lowtea:
		// 下午茶
		return food{
			ID:     td.LowTea,
			Name:   getLine(ctx, kitten.QQ(td.LowTea)),
			LowTea: 1,
		}
	case dinner:
		// 晚餐
		return food{
			ID:     td.Dinner,
			Name:   getLine(ctx, kitten.QQ(td.Dinner)),
			Dinner: 1,
		}
	case supper:
		// 夜宵
		return food{
			ID:     td.Supper,
			Name:   getLine(ctx, kitten.QQ(td.Supper)),
			Supper: 1,
		}
	default:
		return food{}
	}
}
