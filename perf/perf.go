// Package perf 查看服务器运行状况
package perf

import (
	"fmt"
	"math/rand"
	"os"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/Kittengarten/KittenCore/kitten"

	probing "github.com/prometheus-community/pro-bing"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/mem"
	"go.uber.org/zap"

	ctrl "github.com/FloatTech/zbpctrl"
	"github.com/FloatTech/zbputils/control"
	"github.com/FloatTech/zbputils/ctxext"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/extension/rate"
	"github.com/wdvxdr1123/ZeroBot/message"
)

const (
	_ = 1 << (10 * iota)
	// KiB 表示 KiB 所含字节数的变量
	KiB
	// MiB 表示 MiB 所含字节数的变量
	MiB
	// GiB 表示 GiB 所含字节数的变量
	GiB
	// TiB 表示 TiB 所含字节数的变量
	TiB
	// PiB 表示 PiB 所含字节数的变量
	PiB
	// EiB 表示 EiB 所含字节数的变量
	EiB
	// ZiB 表示 ZiB 所含字节数的变量
	ZiB
	// YiB 表示 YiB 所含字节数的变量
	YiB

	replyServiceName = `perf` // 插件名
	brief            = `查看运行状况`
	filePath         = `file.txt` // 保存微星小飞机温度配置文件路径的文件，非 Windows 系统或不使用可以忽略
	cView            = `查看`
)

var (
	config    = kitten.GetMainConfig()                                  // 主配置
	imagePath = kitten.FilePath(config.Path, replyServiceName, `image`) // 图片路径
	poke      = rate.NewManager[int64](5*time.Minute, 9)                // 戳一戳
)

func init() {
	var (
		nickname = config.NickName[0] // 默认昵称
		s        strings.Builder
	)
	for i := range config.NickName {
		s.WriteString(fmt.Sprintln(
			config.CommandPrefix, cView, config.NickName[i], ` // 可获取服务器运行状况`))
	}
	s.WriteString(fmt.Sprint(`戳一戳`, nickname, ` // 可得到响应`))
	// 注册插件
	engine := control.AutoRegister(&ctrl.Options[*zero.Ctx]{
		DisableOnDefault: false,
		Brief:            brief,
		Help:             s.String(),
	}).ApplySingle(ctxext.DefaultSingle)

	// 查看功能
	engine.OnCommand(cView).SetBlock(true).
		Limit(ctxext.NewLimiterManager(time.Hour, 5).LimitByGroup).Handle(func(ctx *zero.Ctx) {
		switch func() (who string) {
			ag, ok := ctx.State[`args`]
			if !ok {
				return
			}
			who, ok = ag.(string)
			if !ok {
				return
			}
			for k := range config.NickName {
				if who == config.NickName[k] {
					who = nickname
				}
			}
			return
		}() {
		case nickname:
			var (
				annoStr, chord, flower, elemental, imagery, err = kitten.GetWTAAnno()
				rAnno                                           string
			)
			if nil != err {
				zap.S().Error(`报时失败喵！`, err)
				rAnno = `喵？`
			} else {
				rAnno = fmt.Sprintf(`%s报时：现在是%s
琴弦：%s
花卉：%s
～%s元灵之%s～`,
					nickname, annoStr,
					chord,
					flower,
					elemental, imagery)
			}
			var (
				cpu = getCPUPercent()
				mem = getMemPercent()
				t   = `45`
			)
			var str string
			switch runtime.GOOS {
			case `windows`:
				t = getCPUTemperatureOnWindows(engine)
				str = fmt.Sprintf(`CPU：%.2f%%
内存：%.0f%%（%s）
%s
体温：%s℃
%s`,
					cpu,
					mem, getMemUsed(),
					getDiskUsedAll(),
					t,
					rAnno)
			case `linux`:
				str = fmt.Sprintf(`CPU：%.2f%%
内存：%.0f%%（%s）
%s
%s`,
					cpu,
					mem, getMemUsed(),
					getDiskUsedAll(),
					rAnno)
			default:
				str = fmt.Sprintf(`CPU：%.2f%%
内存：%.0f%%（%s）
%s`,
					cpu,
					mem, getMemUsed(),
					rAnno)
			}
			img, err := imagePath.GetImage(kitten.Path(strconv.Itoa(getPerf(cpu, mem, t)) + `.png`))
			if nil != err {
				kitten.SendWithImageFail(ctx, `%v`, err)
				return
			}
			kitten.SendMessage(ctx, true, img, message.Text(str))
		default:
			kitten.DoNotKnow(ctx)
		}
	})

	// Ping 功能
	engine.OnCommandGroup([]string{`Ping`, `ping`}, zero.AdminPermission).SetBlock(true).
		Limit(ctxext.NewLimiterManager(time.Minute, 1).LimitByGroup).Handle(func(ctx *zero.Ctx) {
		ag, ok := ctx.State[`args`]
		if !ok {
			return
		}
		pingURL, ok := ag.(string)
		if !ok {
			return
		}
		pg, err := probing.NewPinger(pingURL)
		if nil != err {
			zap.S().Warnf("Ping %s 时出现错误了喵！\n%v", pingURL)
			kitten.DoNotKnow(ctx)
			return
		}
		pg.Count = 4                  // 检测 4 次
		pg.Timeout = 16 * time.Second // 超时时间设置
		var nbytes int
		pg.OnSend = func(pkt *probing.Packet) {
			nbytes = pkt.Nbytes
		}
		var pm strings.Builder
		pg.OnRecv = func(pkt *probing.Packet) {
			pm.WriteString(fmt.Sprintf(`来自 %s 的回复：字节=%d 时间=%dms TTL=%v`, pkt.IPAddr, pkt.Nbytes, pkt.Rtt.Milliseconds(), pkt.TTL))
			pm.WriteByte('\n')
		}
		var r strings.Builder
		pg.OnFinish = func(st *probing.Statistics) {
			r.WriteString(fmt.Sprintf(`正在 Ping %s [%s] 具有 %d 字节的数据：
%s
%s 的 Ping 统计信息：
    数据包：已发送 = %d，已接收 = %d，丢失 = %d（%.0f%% 丢失），
`,
				pingURL, st.IPAddr, nbytes,
				pm.String(),
				st.IPAddr,
				st.PacketsSent, st.PacketsRecv, st.PacketsSent-st.PacketsRecv, st.PacketLoss))
			if 100 > st.PacketLoss {
				r.WriteString(fmt.Sprintf(`往返行程的估计时间：
    最短 = %dms，最长 = %dms，平均 = %dms`,
					st.MinRtt.Milliseconds(), st.MaxRtt.Milliseconds(), st.AvgRtt.Milliseconds()))
			}
		}
		if nil != pg.Run() {
			zap.S().Warn(err)
			kitten.SendWithImageFail(ctx, `%v`, err)
			return
		}
		kitten.SendText(ctx, true, r.String())
	})

	// 戳一戳
	engine.On(`notice/notify/poke`, zero.OnlyToMe).SetBlock(true).Handle(func(ctx *zero.Ctx) {
		var (
			g = ctx.Event.GroupID // 本群的群号
			u = ctx.Event.UserID  // 发出 poke 的 QQ 号
		)
		if `private` == ctx.Event.DetailType {
			g = -u
		}
		switch {
		case poke.Load(g).AcquireN(5):
			// 5 分钟共 9 块命令牌 一次消耗 5 块命令牌
			ctx.SendChain(message.Poke(u))
		case poke.Load(g).AcquireN(3):
			// 5 分钟共 9 块命令牌 一次消耗 3 块命令牌
			kitten.SendWithImageFail(ctx, `请不要拍%s >_<`, nickname)
		case poke.Load(g).Acquire():
			// 5 分钟共 9 块命令牌 一次消耗 1 块命令牌
			kitten.SendWithImageFail(ctx, "喂(#`O′) 拍%s干嘛！\n（好感 - %d）", nickname, 1+rand.Intn(100))
		default:
			// 频繁触发，不回复
		}
	})

	// 图片，用于让 Bot 发送图片，可通过 CQ 码、链接等，为防止滥用，仅管理员可用
	zero.OnCommand(`图片`, zero.AdminPermission).SetBlock(true).Handle(func(ctx *zero.Ctx) {
		if ag, ok := ctx.State[`args`]; ok {
			if img, ok := ag.(string); ok {
				kitten.SendMessage(ctx, true, message.Image(img))
			}
		}
	})
}

// CPU使用率%
func getCPUPercent() float64 {
	p, err := cpu.Percent(time.Second, false)
	if nil != err {
		zap.S().Warn(`获取 CPU 使用率失败了喵！`, err)
	}
	var avg float64
	for k := range p {
		avg += p[k]
	}
	return avg / float64(len(p))
}

// 内存使用调用
func getMem() (m *mem.VirtualMemoryStat) {
	m, err := mem.VirtualMemory()
	if nil != err {
		zap.S().Warn(`获取内存使用失败了喵！`, err)
	}
	return
}

// 内存使用率%
func getMemPercent() float64 {
	return getMem().UsedPercent
}

// 内存使用情况
func getMemUsed() string {
	return fmt.Sprintf(`%.2f MiB/%.2f MiB`, float64(getMem().Used)/MiB, float64(getMem().Total)/MiB)
}

// 磁盘使用调用
func getDisk() (d []*disk.UsageStat) {
	p, err := disk.Partitions(false)
	if nil != err {
		zap.S().Warn(`获取磁盘分区失败了喵！`, err)
		return
	}
	for i := range p {
		if d[i], err = disk.Usage(p[i].Mountpoint); nil != err {
			zap.S().Warn(`获取磁盘信息失败了喵！`, err)
		}
	}
	return
}

// 系统盘使用情况
func getDiskUsed() string {
	return fmt.Sprintf(`系统盘：%.2f%%（%.2f GiB/%.2f GiB）`,
		getDisk()[0].UsedPercent, float64(getDisk()[0].Used)/GiB, float64(getDisk()[0].Total)/GiB)
}

// 全部磁盘使用情况
func getDiskUsedAll() string {
	var (
		b strings.Builder
		d = getDisk()
	)
	for i := range d {
		b.WriteString(fmt.Sprintf("磁盘 %d：%.2f%%（%.2f GiB/%.2f GiB）",
			i, d[i].UsedPercent, float64(d[i].Used)/GiB, float64(d[i].Total)/GiB))
		if i < len(d)-1 {
			b.WriteByte('\n')
		}
	}
	return b.String()
}

// Windows 系统下获取 CPU 温度，通过微星小飞机（需要自行安装配置，并确保温度在其 log 中的位置）
func getCPUTemperatureOnWindows(e *control.Engine) string {
	n := kitten.FilePath(e.DataFolder(), filePath)
	if err := kitten.InitFile(&n, `C:\Program Files (x86)\MSI Afterburner\HardwareMonitoring.hml`); nil != err {
		zap.S().Warn(err)
	}
	p, err := n.LoadPath()
	if nil != err {
		zap.S().Warn(err)
	}
	if nil != os.Remove(p.String()) {
		zap.S().Warn(err)
	}
	<-time.NewTimer(1 * time.Second).C
	file, err := os.ReadFile(p.String())
	if nil != err {
		zap.S().Warn(err)
	}
	return string(file[329:331]) // 此处为温度在微星小飞机 log 中的位置
}

// 返回状态等级
func getPerf(cpu float64, mem float64, ts string) int {
	ti, err := strconv.Atoi(ts)
	if nil != err {
		zap.S().Warn(err)
		return 5
	}
	if 0 >= ti || 100 <= ti {
		return 5
	}
	perf := 0.00005 * (cpu + mem) * float64(ti)
	zap.S().Debugf(`%s的负荷评分是 %f……`, zero.BotConfig.NickName[0], perf)
	switch {
	case 0.1 > perf:
		return 0
	case 0.15 > perf:
		return 1
	case 0.2 > perf:
		return 2
	case 0.25 > perf:
		return 3
	case 0.3 > perf:
		return 4
	default:
		return 5
	}
}
