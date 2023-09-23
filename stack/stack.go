// Package stack 叠猫猫
package stack

import (
	"errors"
	"fmt"
	"math"
	"math/rand"
	"runtime/debug"
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
	replyServiceName = `stack` // 插件名
	brief            = `一起来玩叠猫猫`
	dataFile         = `data.yaml` // 叠猫猫数据文件
	exitFile         = `exit.yaml` // 叠猫猫退出日志文件
	cStack           = `叠猫猫`
	cIn              = `加入`
	cExit            = `退出`
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
	err := kitten.InitFile(&configFile, `maxstack: 10    # 叠猫猫队列上限
maxtime: 2      # 叠猫猫时间上限（小时数）
gaptime: 1      # 叠猫猫主动退出或者被压坏后重新加入所需的时间（小时数）
outofstack: "不能再叠了，下面的猫猫会被压坏的喵！"  # 叠猫猫队列已满的回复
maxcount: 5     # 被压次数上限
failpercent: 1  # 叠猫猫每层失败概率百分数`)
	if nil != err {
		zap.Error(err)
		return
	}
	var (
		help = fmt.Sprintf(`%s%s %s|%s|%s
叠猫猫每层高度有 %d%% 概率会失败
最多可以叠 %d 只猫猫哦
在叠猫猫队列中超过 %d 小时后，会自动退出
主动退出叠猫猫；试图压别的猫猫；被压超过 %d 次且位于下半部分；叠猫猫失败摔下来——这些情况需要 %d 小时后，才能再次加入`,
			kitten.GetMainConfig().CommandPrefix, cStack, cIn, cExit, cView,
			stackConfig.FailPercent,
			stackConfig.MaxStack,
			stackConfig.MaxTime,
			stackConfig.MaxCount, stackConfig.GapTime)
		// 注册插件
		engine = control.AutoRegister(&ctrl.Options[*zero.Ctx]{
			DisableOnDefault:  false,
			Brief:             brief,
			Help:              help,
			PrivateDataFolder: replyServiceName,
		})
	)

	// 自动退出的协程
	go autoExit(dataFile, stackConfig, engine)
	go autoExit(exitFile, stackConfig, engine)

	engine.OnCommand(cStack).SetBlock(true).
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
			return
		}
		e, err := loadData(kitten.FilePath(engine.DataFolder(), exitFile))
		if nil != err {
			zap.S().Error(err)
			kitten.SendWithImageFail(ctx, `%v`, err)
			return
		}
		if !ok {
			return
		}
		if `guild` == ctx.Event.DetailType {
			ctx.Send(kitten.Guild)
			return
		}
		switch op {
		case cIn:
			d.in(ctx, e, engine)
		case cExit:
			d.exit(ctx, engine)
		case cView:
			d.view(ctx)
		default:
			ctx.Send(help)
		}
	})
}

// 压猫猫
func (d data) press(ctx *zero.Ctx, u int64, en *control.Engine) {
	var r strings.Builder
	r.WriteString(stackConfig.OutOfStack)
	ld := len(d)
	if exitLabel := -1; checkStack(ld + 1) {
		// 只有下半的猫猫会被压坏
		for i := range d {
			if d[i].Count++; d[i].Count > stackConfig.MaxCount && i < ld/2 {
				exitLabel = i // 最上面一只压坏的猫猫的位置
			}
		}
		if 0 <= exitLabel {
			// 如果有猫猫被压坏
			exitData := d[:exitLabel+1]
			// 反序排列退出队列
			slices.Reverse(exitData)
			// 将被压坏的的猫猫记录至退出日志
			for k := range exitData {
				if err := logExit(ctx, exitData[k].ID, en); nil != err {
					zap.S().Error(err)
					return
				}
			}
			d = d[1+exitLabel:]
			d.save(kitten.FilePath(en.DataFolder(), dataFile))
			r.WriteString(fmt.Sprintf(`

压猫猫成功，下面的猫猫对你的好感度下降了！你在 %d 小时内无法加入叠猫猫。

有 %d 只猫猫被压坏了喵！需要休息 %d 小时。
%s`,
				stackConfig.GapTime,
				exitLabel+1, stackConfig.GapTime,
				strings.Join(exitData.toString(), "\n")))
		} else {
			// 如果没有猫猫被压坏
			d.save(kitten.FilePath(en.DataFolder(), dataFile))
			r.WriteString(fmt.Sprintf("\n\n压猫猫成功，下面的猫猫对你的好感度下降了！你在 %d 小时内无法加入叠猫猫。", stackConfig.GapTime))
		}
	} else {
		r.WriteString(fmt.Sprintf("\n\n压猫猫失败了喵！你在 %d 小时内无法加入叠猫猫。", stackConfig.GapTime))
	}
	// 将压猫猫的猫猫记录至退出日志
	if err := logExit(ctx, u, en); nil != err {
		zap.S().Error(err)
		kitten.SendWithImageFail(ctx, `%v`, err)
		return
	}
	kitten.SendWithImageFail(ctx, r.String())
}

// 加入叠猫猫
func (d data) in(ctx *zero.Ctx, e data, en *control.Engine) {
	id := ctx.Event.UserID
	if slices.ContainsFunc(e, func(k meow) bool { return id == k.ID }) {
		kitten.SendWithImageFail(ctx, fmt.Sprintf(`休息不足 %d 小时，不能加入喵！`, stackConfig.GapTime))
		return
	}
	if slices.ContainsFunc(d, func(k meow) bool { return id == k.ID }) {
		kitten.SendWithImageFail(ctx, `已经加入叠猫猫了喵！`)
		return
	}
	ld := len(d)
	if ld >= stackConfig.MaxStack {
		// 压猫猫
		d.press(ctx, id, en)
	} else if checkStack(ld + 1) {
		// 如果叠猫猫成功
		d = append(d, meow{
			ID:   id,
			Name: kitten.QQ(id).GetTitleCardOrNickName(ctx),
			Time: time.Unix(ctx.Event.Time, 0),
		})
		if err := d.save(kitten.FilePath(en.DataFolder(), dataFile)); nil != err {
			rt := `叠猫猫文件存储失败喵！`
			zap.S().Error(id, rt, err)
			kitten.SendWithImageFail(ctx, rt)
			return
		}
		kitten.SendTextOf(ctx, true, `叠猫猫成功，目前处于队列中第 %d 位喵～`, len(d))
	} else {
		// 如果叠猫猫失败
		if ld != 0 {
			// 如果不是平地摔
			exitCount := int(math.Ceil(float64(ld) * rand.Float64()))
			if 0 == exitCount {
				exitCount = 1
			}
			exitData := d[ld-exitCount:]
			// 将摔下来的的猫猫记录至退出日志
			for k := range exitData {
				err := logExit(ctx, exitData[k].ID, en)
				if nil != err {
					zap.S().Error(err)
					kitten.SendWithImageFail(ctx, `%v`, err)
					return
				}
			}
			d = d[:ld-exitCount]
			if err := d.save(kitten.FilePath(en.DataFolder(), dataFile)); nil != err {
				rt := `叠猫猫文件存储失败喵！`
				zap.S().Error(id, rt, err)
				kitten.SendWithImageFail(ctx, rt)
				return
			}
			// 反序排列退出队列
			slices.Reverse(exitData)
			// 构建退出报告
			kitten.SendWithImageFail(ctx, "叠猫猫失败，上面 %d 只猫猫摔下来了喵！需要休息 %d 小时。\n%s",
				exitCount, stackConfig.GapTime, strings.Join(exitData.toString(), "\n"))
		} else {
			// 如果是平地摔
			kitten.SendWithImageFail(ctx, `叠猫猫失败，你平地摔了喵！需要休息 %d 小时。`, stackConfig.GapTime)
		}
		// 将叠猫猫失败的猫猫记录至退出日志
		if err := logExit(ctx, id, en); nil != err {
			zap.S().Error(err)
			kitten.SendWithImageFail(ctx, `%v`, err)
		}
	}
}

// 退出叠猫猫
func (d data) exit(ctx *zero.Ctx, en *control.Engine) {
	var (
		u = ctx.Event.UserID
		i = slices.IndexFunc(d, func(m meow) bool { return u == m.ID }) // 退出叠猫猫的猫猫位置
	)
	if -1 == i {
		kitten.SendWithImageFail(ctx, `没有加入叠猫猫，不能退出喵！`)
		return
	}
	// 从叠猫猫队列中删除退出的猫猫
	d = slices.Delete(d, i, 1+i)
	if err := d.save(kitten.FilePath(en.DataFolder(), dataFile)); nil != err {
		r := `退出叠猫猫时，文件存储失败喵！`
		zap.S().Error(u, r, err)
		kitten.SendWithImageFail(ctx, r)
		return
	}
	// 将退出的猫猫记录至退出日志
	if err := logExit(ctx, u, en); nil != err {
		zap.S().Error(err)
		kitten.SendWithImageFail(ctx, `%v`, err)
		return
	}
	kitten.SendText(ctx, true, `退出叠猫猫成功喵！`)
}

// 查看叠猫猫
func (d data) view(ctx *zero.Ctx) {
	const h = `【叠猫猫队列】`
	var r strings.Builder
	r.WriteString(h)
	r.WriteByte('\n')
	// 反序排列叠猫猫队列
	slices.Reverse(d)
	r.WriteString(strings.Join(d.toString(), "\n")) // 生成播报
	if 0 >= len(d) {
		r.Reset()
		r.WriteString(`暂时没有猫猫哦`)
	}
	ctx.Send(r.String())
}

// 从叠猫猫队列生成字符串切片
func (d data) toString() (s []string) {
	for i, ld := 0, len(d); i < ld; i++ {
		s = append(s, fmt.Sprintf(`%s（%d）`, d[i].Name, d[i].ID))
	}
	return
}

// 叠猫猫数据文件存储
func (d data) save(path kitten.Path) (err error) {
	b, err := yaml.Marshal(d)
	err = errors.Join(err, path.Write(b))
	return
}

// 自动退出队列
func autoExit[T string | kitten.Path](f T, c config, e *control.Engine) {
	// 处理 panic，防止程序崩溃
	defer func() {
		if err := recover(); nil != err {
			zap.S().Error(err)
			debug.PrintStack()
		}
	}()
	var limitTime time.Duration
	switch f {
	case dataFile:
		limitTime = time.Hour * time.Duration(c.MaxTime)
	case exitFile:
		limitTime = time.Hour * time.Duration(c.GapTime)
	}
	if 0 == limitTime {
		limitTime = time.Hour
	}
	for {
		if err := func() error {
			p := kitten.FilePath(e.DataFolder(), string(f))
			return kitten.InitFile(&p, kitten.Empty)
		}; nil != err() {
			zap.Error(err())
			return
		}
		d, err := loadData(kitten.FilePath(e.DataFolder(), string(f)))
		if nil != err {
			zap.S().Error(err)
		}
		nextTime := time.Now().Add(limitTime)
		ld := len(d) // 退出前的猫猫数量
		if 0 < ld {
			if limitTime < time.Since(d[0].Time) {
				if 1 < ld {
					nextTime = d[1].Time.Add(limitTime)
				}
				d = d[1:]
			} else {
				nextTime = d[0].Time.Add(limitTime)
			}
		}
		if ld != len(d) {
			d.save(kitten.FilePath(e.DataFolder(), string(f)))
		}
		zap.S().Infof(`下次定时退出 %s 时间为：%s`, kitten.FilePath(e.DataFolder(), string(f)), nextTime.Format(kitten.Layout))
		<-time.NewTimer(time.Until(nextTime)).C
	}
}

// 记录至退出日志
func logExit(ctx *zero.Ctx, u int64, e *control.Engine) (err error) {
	dataExit, err := loadData(kitten.FilePath(e.DataFolder(), exitFile))
	if nil != err {
		return
	}
	dataExit = append(dataExit, meow{
		ID:    u,
		Name:  kitten.QQ(u).GetTitleCardOrNickName(ctx),
		Time:  ctx.Event.RawEvent.Time(),
		Count: 0,
	})
	return dataExit.save(kitten.FilePath(e.DataFolder(), exitFile))
}

// 根据高度 h 检查压猫猫或叠猫猫是否成功
func checkStack(h int) bool {
	return 0.01*float64(h*stackConfig.FailPercent) <= rand.Float64()
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
