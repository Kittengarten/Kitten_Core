// Package stack2 叠猫猫 v2
package stack2

import (
	"errors"
	"fmt"
	"math"
	"math/rand"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/Kittengarten/KittenCore/kitten"

	"go.uber.org/zap"
	"gopkg.in/yaml.v3"

	ctrl "github.com/FloatTech/zbpctrl"
	"github.com/FloatTech/zbputils/control"
	"github.com/FloatTech/zbputils/ctxext"
	zero "github.com/wdvxdr1123/ZeroBot"
)

const (
	replyServiceName = `stack2` // 插件名
	brief            = `一起来玩叠猫猫 v2`
	dataFile         = `data.yaml` // 叠猫猫数据文件
	exitFile         = `exit.yaml` // 叠猫猫退出日志文件
	cStack           = `叠猫猫`
	cIn              = `加入`
	cView            = `查看`
)

var (
	configFile       = kitten.FilePath(replyServiceName, `config.yaml`) // 叠猫猫配置文件名
	stackConfig, err = loadConfig(configFile)                           // 叠猫猫配置文件
	mu               sync.Mutex
)

func init() {
	if nil != err {
		zap.Error(err)
		return
	}
	// 初始化叠猫猫配置文件
	err := kitten.InitFile(&configFile, `gaptime: 1        # 每千克体重的冷却时间（小时数）
mingaptime: 1     # 最小冷却时间（小时数）`)
	if nil != err {
		zap.Error(err)
		return
	}
	var (
		help = fmt.Sprintf(`%s%s %s|%s
被别的猫猫压坏；叠猫猫失败摔下来——这些情况需要休息一段时间后，才能再次加入`,
			kitten.GetMainConfig().CommandPrefix, cStack, cIn, cView)
		// 注册插件
		engine = control.AutoRegister(&ctrl.Options[*zero.Ctx]{
			DisableOnDefault:  false,
			Brief:             brief,
			Help:              help,
			PrivateDataFolder: replyServiceName,
		})
	)

	engine.OnCommand(`叠猫猫`).SetBlock(true).
		Limit(ctxext.NewLimiterManager(time.Minute, 5).LimitByGroup).Handle(func(ctx *zero.Ctx) {
		ag, ok := ctx.State[`args`]
		if !ok {
			return
		}
		op, ok := ag.(string)
		if !ok {
			return
		}
		mu.Lock()
		defer mu.Unlock()
		d, err := loadData(kitten.FilePath(engine.DataFolder(), dataFile))
		if nil != err {
			zap.S().Error(err)
			kitten.SendWithImageFail(ctx, `%v`, err)
		}
		if `guild` == ctx.Event.DetailType {
			ctx.Send(kitten.Guild)
			return
		}
		switch op {
		case `加入`:
			d.in(ctx, engine)
		case `查看`:
			m := d.getStack()
			m.view(ctx)
		default:
			ctx.Send(help)
		}
	})
}

// 加入叠猫猫
func (d *data) in(ctx *zero.Ctx, en *control.Engine) {
	var (
		u = ctx.Event.UserID // 叠猫猫的 QQ
		c time.Duration      // 剩余的冷却时间
	)
	if slices.ContainsFunc(*d, func(k meow) bool {
		c = k.Time.Sub(time.Unix(ctx.Event.Time, 0))
		return u == k.ID && !k.Status && 0 < c
	}) {
		kitten.SendWithImageFail(ctx, `还需要休息 %s 小时才能加入喵！`, c.Round(time.Second))
		return
	}
	if slices.ContainsFunc(*d, func(k meow) bool { return u == k.ID && k.Status }) {
		kitten.SendWithImageFail(ctx, `已经加入叠猫猫了喵！`)
		return
	}
	// 叠入的猫猫及其下标
	k, i := d.getMeow(u)
	switch i {
	case -1:
		// 如果是首次叠猫猫
		name := kitten.QQ(u).GetTitleCardOrNickName(ctx)
		k = meow{
			ID:     u,
			Name:   name,
			Weight: len(name),
			Time:   time.Unix(ctx.Event.Time, 0),
		}
	default:
		// 如果是已经存在的猫猫，更新其名称
		k.Name = kitten.QQ(u).GetTitleCardOrNickName(ctx)
	}
	dn := d.getNoStack()
	*d = d.getStack()
	dr := slices.Clone(*d)
	var r strings.Builder
	r.WriteString(`叠猫猫失败，杂鱼～杂鱼～`)
	r.WriteByte('\n')
	switch n := d.getStackResult(ctx, k); n {
	case 0: // 如果没有猫猫摔下来，叠猫猫初步成功
		p := d.getPressResult(ctx, k)
		if 0 != p {
			// 压坏了别的猫猫
			r.WriteString(`有 %d 只猫猫被压坏了喵！需要休息一段时间。`)
			r.WriteByte('\n')
			exit(ctx, &k, press, p)
			e := dr[:p]
			r.WriteString(e.toString())
			kitten.SendWithImage(ctx, kitten.Path(`杂鱼.png`), r.String(), p)
			break
		}
		// 如果没有压坏猫猫，叠猫猫成功
		k.Status = true
		kitten.SendTextOf(ctx, true, `叠猫猫成功，目前处于队列中第 %d 位喵～`, 1+len(*d))
	case -1: // 如果平地摔
		r.WriteString(`你平地摔了喵！需要休息约 %.2f 小时。`)
		kitten.SendWithImage(ctx, kitten.Path(`杂鱼.png`), r.String(), float64(k.Weight*stackConfig.GapTime)/10)
		exit(ctx, &k, flat, 0)
	default: // 如果叠猫猫失败，有猫猫摔下来
		r.WriteString(`上面 %d 只猫猫摔下来了喵！需要休息一段时间。`)
		r.WriteByte('\n')
		exit(ctx, &k, fall, len(*d))
		e := dr[len(dr)-n:]
		r.WriteString(e.toString())
		kitten.SendWithImage(ctx, kitten.Path(`杂鱼.png`), r.String(), n)
	}
	// 合并当前未叠猫猫与叠猫猫的队列
	*d = append(dn, *d...)
	// 将叠入的猫猫追加入原始数据
	*d = append(*d, k)
	// 存储叠猫猫数据
	if err := d.save(kitten.FilePath(en.DataFolder(), dataFile)); nil != err {
		r := "叠猫猫文件存储失败喵！"
		zap.S().Error(u, r, err)
		kitten.SendWithImageFail(ctx, r)
	}
}

// 获取并返回叠猫猫队列
func (d *data) getStack() data {
	// 删除不在叠猫猫中的猫猫，得到叠猫猫队列
	return slices.DeleteFunc(slices.Clone(*d), func(k meow) bool { return !k.Status })
}

// 获取并返回不在叠猫猫的队列
func (d *data) getNoStack() data {
	// 删除叠猫猫中的猫猫，得到不在叠猫猫的队列
	return slices.DeleteFunc(slices.Clone(*d), func(k meow) bool { return k.Status })
}

// 查看叠猫猫
func (d *data) view(ctx *zero.Ctx) {
	s := d.getStack()
	if 0 >= len(s) {
		kitten.SendTextOf(ctx, true, "【叠猫猫队列】\n暂时没有猫猫哦")
		return
	}
	kitten.SendTextOf(ctx, true, `【叠猫猫队列】%s`, s.toString())
}

// 叠猫猫数据文件存储
func (d *data) save(path kitten.Path) (err error) {
	b, err := yaml.Marshal(d)
	err = errors.Join(err, path.Write(b))
	return
}

/*
提取猫猫及其下标

无此猫猫则返回空结构体及 -1
*/
func (d *data) getMeow(u int64) (m meow, i int) {
	i = slices.IndexFunc(*d, func(k meow) bool { return u == k.ID })
	if 0 > i {
		m = meow{}
		return
	}
	m = (*d)[i]
	*d = slices.Delete(*d, i, 1+i)
	return
}

// 从叠猫猫队列生成字符串
func (d *data) toString() string {
	// 克隆一份防止修改源数据
	m := slices.Clone(*d)
	// 按“后来居上”排列叠猫猫队列
	slices.Reverse(m)
	var s strings.Builder
	for i := range *d {
		s.WriteString(fmt.Sprintf("\n%s（%d，%.1f kg）", m[i].Name, m[i].ID, float64(m[i].Weight)/10))
	}
	return s.String()
}

// 获取全队列的总重量
func (d *data) getTotalWeight() (w int) {
	for i := range *d {
		w += (*d)[i].Weight
	}
	return
}

/*
检查最下方的猫猫是否被压坏

如果没有被压坏则返回 true
*/
func (d *data) checkPress() bool {
	if 1 >= len(*d) {
		// 如果只有一只猫猫或者没有猫猫，直接返回，避免下标越界
		return true
	}
	a := (*d)[1:]
	return float64(a.getTotalWeight()-(*d)[0].Weight)/float64(d.getTotalWeight()) <= rand.Float64()
}

/*
获取被压坏猫猫的数量，并将被压坏的猫猫标记为未在叠猫猫

不含叠入的猫猫
*/
func (d *data) getPressResult(ctx *zero.Ctx, k meow) int {
	var (
		s  = append(*d, k) // 将叠入的猫猫纳入队列重量计算
		ld = len(*d)
	)
	for i := range *d {
		if a := s[i:]; !a.checkPress() {
			// 如果压猫猫失败，则直接返回
			return i
		}
		// 去除压坏的猫猫后，再检查队列
		exit(ctx, &(*d)[i], pressed, ld-i)
	}
	return ld
}

/*
检查是否因为叠猫猫失败摔下来

m 为上方的猫猫，n 为下方的猫猫

如果没有摔下来则返回 true
*/
func (m meow) checkStack(n meow) bool {
	return float64(m.Weight)/float64(m.Weight+n.Weight) <= rand.Float64()
}

/*
获取叠猫猫失败摔下来猫猫的数量，并将摔下来的猫猫标记为未在叠猫猫

不含叠入的猫猫

如果平地摔则返回 -1
*/
func (d *data) getStackResult(ctx *zero.Ctx, k meow) int {
	ld := len(*d) // 初始猫猫数量
	if 0 == ld {
		// 当叠猫猫队列为空， 1% 概率平地摔
		if 0.01 > rand.Float64() {
			return -1
		}
		return 0
	}
	// 从队列的最上部开始遍历（后来居上）
	for i := range *d {
		if k.checkStack((*d)[ld-i-1]) {
			// 如果这只猫猫没有摔下来，则直接返回
			return i
		}
		k = (*d)[ld-i-1]
		// 去除队列中最上方摔下来的猫猫
		exit(ctx, &(*d)[ld-i-1], fall, ld-i)
	}
	return ld
}

/*
去除退出的猫猫 k

t 为类型，h 为 摔下来的高度 | 压坏的猫猫总数 | 上方的猫猫总数
*/
func exit(ctx *zero.Ctx, k *meow, t byte, h int) {
	// 去除
	k.Status = false
	// 冷却
	k.Time = time.Unix(ctx.Event.Time, 0).Add(max(
		time.Hour*time.Duration(stackConfig.MinGapTime),
		time.Hour*time.Duration(stackConfig.GapTime*k.Weight)/10))
	switch t {
	case flat:
		// 平地摔，体重翻倍
		if k.Weight > math.MaxInt/2 {
			k.Weight = math.MaxInt
		} else {
			k.Weight *= 2
		}
	case fall:
		// 摔下来，体重 - 100g × 当前高度
		k.Weight = max(1, k.Weight-h)
	case press:
		// 压坏了猫猫，体重 + 100g × 压坏的猫猫总数
		k.Weight = min(k.Weight, math.MaxInt-h) + h
	case pressed:
		// 被压坏，体重 + 100g × 上方的猫猫总数
		k.Weight = min(k.Weight, math.MaxInt-h) + h
	}
}

// 加载叠猫猫配置
func loadConfig[T string | kitten.Path](p T) (c config, err error) {
	if err = kitten.InitFile(&p, kitten.Empty); nil != err {
		return
	}
	d, err := kitten.Path(p).Read()
	if nil != err {
		return
	}
	err = yaml.Unmarshal(d, &c)
	return
}

// 加载叠猫猫数据
func loadData[T string | kitten.Path](p T) (d data, err error) {
	if err = kitten.InitFile(&p, kitten.Empty); nil != err {
		return
	}
	b, err := kitten.Path(p).Read()
	if nil != err {
		return
	}
	err = yaml.Unmarshal(b, &d)
	return
}
