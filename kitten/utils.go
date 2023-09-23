package kitten

import (
	"math/rand"
	"strings"
	"time"
	"unicode"

	"github.com/Kittengarten/KittenAnno/wta"

	wr "github.com/mroth/weightedrand/v2"
)

// Choose 按权重抽取一个项目的序号
func (c Choices) Choose() (result int, err error) {
	choices := make([]wr.Choice[int, int], len(c), len(c))
	for i := range c {
		var (
			item   = c[i].GetID()
			weight = c[i].GetChance()
		)
		choices[i] = wr.Choice[int, int]{Item: item, Weight: weight}
	}
	chooser, err := wr.NewChooser(choices...)
	if nil != err {
		return -1, err
	}
	return chooser.Pick(), nil
}

// IsSameDate 判断两个时间是否在同一天
func IsSameDate(t1 time.Time, t2 time.Time) bool {
	return t1.Day() == t2.Day() && t1.Month() == t2.Month() && t1.Year() == t2.Year()
}

/*
CleanAll 清理字符串中全部不必要内容

lf 控制是否换行
*/
func CleanAll[T string | []rune | []byte](s T, lf bool) T {
	return T(strings.Map(func(r rune) rune {
		if remove := unicode.IsControl(r) || unicode.IsSpace(r);
		// 如果不换行，移除包括换行符在内的控制字符和空白字符
		(!lf && remove) ||
			// 如果换行，移除换行符以外的控制字符和空白字符
			(lf && remove && !strings.ContainsRune("\n\r", r)) ||
			// 移除可能引发排版和显示错误的字符
			strings.ContainsRune("\u061c\u200e\u200f\u202a\u202b\u202c\u202d\u202e\u2066\u2067\u2068\u2069\ufffd", r) {
			return -1
		}
		return r
	}, string(s)))
}

/*
GetMidText 获取中间字符串

pre 为前缀（不包含），suf 为后缀（不包含），str 为整个字符串
*/
func GetMidText(pre string, suf string, str string) string {
	return str[func() int {
		// 截掉前缀及之前部分
		if i := strings.Index(str, pre); -1 != i {
			return i + len(pre)
		}
		return 0
	}():func() int {
		// 截掉后缀及之后部分
		if i := strings.LastIndex(str, suf); -1 != i {
			return i
		}
		return len(str)
	}()]
}

// GetWTAAnno 获取世界树纪元的完整字符串和额外信息
func GetWTAAnno() (str string, chord string, flower string, elemental string, imagery string, err error) {
	anno, err := wta.GetAnno()
	if nil != err {
		return
	}
	str, chord = anno.GetAnnoStrSplit()
	flower, elemental, imagery = anno.Flower, anno.Elemental, anno.Imagery
	return
}

// GenerateRandomNumber 生成 count 个 [start, end) 范围的不重复的随机数
func GenerateRandomNumber(start, end, count int) []int {
	// 范围检查
	if end <= start || (end-start) < count || 0 == count {
		return nil
	}
	// 存放结果的集合（不重复）
	set := make(map[int]struct{}, count)
	for len(set) < count {
		// 生成随机数
		set[rand.Intn(end-start)+start] = struct{}{}
	}
	var (
		// 存放结果的切片
		nums = make([]int, count, count)
		// 切片下标
		i int
	)
	// 集合转换为切片
	for k := range set {
		nums[i] = k
		i++
	}
	return nums
}
