package stack2

import (
	"strings"
	"sync"
	"time"
)

type loc byte // 地区

const cockroachDoNotAnalysis = `蟑螂不会分析，蟑螂只会勇敢地创上去`

const (
	cat       loc = iota // 叠猫猫
	cockroach            // 叠蟑螂
)

var (
	// 各地区的字符串
	l10nStr = map[string]map[loc]string{
		`猫猫`:         {cat: `猫猫`, cockroach: `蟑螂`},
		`压坏`:         {cat: `压坏`, cockroach: `压爆浆`},
		`体重`:         {cat: `体重`, cockroach: `翼展`},
		`绒布球`:        {cat: `绒布球`, cockroach: `蟑螂卵鞘`},
		`奶猫`:         {cat: `奶猫`, cockroach: `德国蟑螂`},
		`抱枕`:         {cat: `抱枕`, cockroach: `美洲蟑螂`},
		`小可爱`:        {cat: `小可爱`, cockroach: `广东蟑螂`},
		`大可爱`:        {cat: `大可爱`, cockroach: `澳洲蟑螂`},
		`幼年猫娘`:       {cat: `幼年猫娘`, cockroach: `幼年蟑螂娘`},
		`猫娘萝莉`:       {cat: `猫娘萝莉`, cockroach: `蟑螂萝莉`},
		`猫娘少女`:       {cat: `猫娘少女`, cockroach: `蟑螂少女`},
		`成年猫娘`:       {cat: `成年猫娘`, cockroach: `成年蟑螂娘`},
		`小老虎`:        {cat: `小老虎`, cockroach: `精英蟑螂娘`},
		`大老虎`:        {cat: `大老虎`, cockroach: `蟑螂母体`},
		`猫车`:         {cat: `猫车`, cockroach: `蟑螂恶霸`},
		`猫猫巴士`:       {cat: `猫猫巴士`, cockroach: `蟑螂基地车`},
		`猫卡`:         {cat: `猫卡`, cockroach: `蟑螂运输船`},
		`虎式坦克`:       {cat: `虎式坦克`, cockroach: `蟑螂轨道炮`},
		`■■■`:        {cat: `■■■`, cockroach: `蟑螂歼星舰`},
		`发生平地摔`:      {cat: `发生平地摔`, cockroach: `翻了个身`},
		`猫堆`:         {cat: `猫堆`, cockroach: `蟑螂巢穴`},
		`触发特效`:       {cat: `触发特效`, cockroach: `获得康复新液`},
		`猫娘`:         {cat: `猫娘`, cockroach: `蟑螂娘`},
		`平地摔`:        {cat: `平地摔`, cockroach: `翻了个身`},
		`休息`:         {cat: `休息`, cockroach: `蛰伏`},
		`kg`:         {cat: `kg`, cockroach: `cm`},
		`触发了清空猫堆的特效`: {cat: `触发了清空猫堆的特效`, cockroach: `获得了康复新液`},
		`平地摔了喵`:      {cat: `平地摔了喵`, cockroach: `翻了个身`},
		`🙀`:          {cat: `🙀`, cockroach: `🪳`},
		`😿`:          {cat: `😿`, cockroach: `🪳`},
		`🐅`:          {cat: `🐅`, cockroach: `🪳`},
		`🐯`:          {cat: `🐯`, cockroach: `🪳`},
		`喵！`:         {cat: `喵！`, cockroach: `！`},
		`喵～`:         {cat: `喵～`, cockroach: `～`},
		`总重量为`:       {cat: `总重量为`, cockroach: `总长度为`},
		`老虎`:         {cat: `老虎`, cockroach: `精英级以上蟑螂`},
	}
	// 字符替换器
	l10nReplacer = func(l loc) *strings.Replacer {
		return strings.NewReplacer(mapToSlice(l10nStr, l)...)
	}
	// 地区标记位
	globalLocation loc
	// 地区互斥锁
	muLocation sync.Mutex
)

func mapToSlice(m map[string]map[loc]string, l loc) (s []string) {
	if cat == l {
		// 叠猫猫无需替换
		return
	}
	for _, v := range m {
		old, ok := v[cat]
		if !ok {
			continue
		}
		new, ok := v[l]
		if !ok {
			continue
		}
		s = append(s, []string{old, new}...)
	}
	return
}

// 叠蟑螂活动日期判断，在愚人节的前三天或后七天范围内返回 true
func checkCockroachDate() bool {
	switch time.Now().Month() {
	case time.March:
		return 28 < time.Now().Day()
	case time.April:
		return 9 > time.Now().Day()
	default:
		return false
	}
}
