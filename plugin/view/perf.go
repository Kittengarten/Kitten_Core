package view

import (
	"fmt"
	"net/http"
	"os"
	"runtime"
	"runtime/debug"
	"strconv"
	"strings"
	"time"

	"github.com/Kittengarten/KittenCore/kitten"
	"github.com/Kittengarten/KittenCore/kitten/core"

	hm "github.com/dustin/go-humanize"
	probing "github.com/prometheus-community/pro-bing"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/mem"

	"github.com/FloatTech/floatbox/process"
	zero "github.com/wdvxdr1123/ZeroBot"
)

// 返回查看字符串
func viewString() string {
	m := getMem()
	switch runtime.GOOS {
	case `windows`:
		return fmt.Sprintf(`CPU：	%.2f%%%s
内存：	%.1f%%	（%s）
%s
体温：	%s℃

%s`,
			cpuPercent(), weight(),
			percent(m), use(m),
			diskUsedAll(),
			cpuTemperatureOnWindows(),
			getWTA())
	case `linux`:
		return fmt.Sprintf(`CPU：	%.2f%%%s
内存：	%.1f%%	（%s）
%s

%s`,
			cpuPercent(), weight(),
			percent(m), use(m),
			diskUsedAll(),
			getWTA())
	default:
		return fmt.Sprintf(`CPU：	%.2f%%%s
内存：	%.1f%%	（%s）

%s`,
			cpuPercent(), weight(),
			percent(m), use(m),
			getWTA())
	}
}

// CPU 使用率 %
func cpuPercent() float64 {
	p, err := cpu.Percent(time.Second, false)
	if nil != err {
		kitten.Warnln(`获取 CPU 使用率失败了喵！`, err)
		return 0
	}
	var avg float64
	for _, c := range p {
		avg += c
	}
	return avg / float64(len(p))
}

// 内存使用调用
func getMem() *mem.VirtualMemoryStat {
	m, err := mem.VirtualMemory()
	if nil != err {
		kitten.Warnln(`获取内存使用失败了喵！`, err)
		return &mem.VirtualMemoryStat{}
	}
	return m
}

// 内存使用率 %
func percent(m *mem.VirtualMemoryStat) float64 {
	return 100 * float64(m.Total-m.Free) / float64(m.Total)
}

// 内存使用情况
func use(m *mem.VirtualMemoryStat) string {
	return hm.IBytes(m.Total-m.Free) + ` / ` + hm.IBytes(m.Total)
}

// 磁盘使用调用
func getDisk() []*disk.UsageStat {
	p, err := disk.Partitions(false)
	if nil != err {
		kitten.Warnln(`获取磁盘分区失败了喵！`, err)
		return nil
	}
	d := make([]*disk.UsageStat, len(p), len(p))
	for i, s := range p {
		if d[i], err = disk.Usage(s.Mountpoint); nil != err {
			kitten.Warnln(`获取磁盘信息失败了喵！`, err)
		}
	}
	return d
}

// 全部磁盘使用情况
func diskUsedAll() string {
	var (
		b strings.Builder
		d = getDisk()
	)
	b.Grow(16 * len(d))
	for i, s := range d {
		b.WriteString(fmt.Sprintf(`磁盘 %d：	%.1f%%	（%s / %s）`,
			i, s.UsedPercent, hm.IBytes(s.Used), hm.IBytes(s.Total)))
		b.WriteByte('\n')
	}
	return b.String()[:b.Len()-1]
}

// Windows 系统下获取 CPU 温度，通过微星小飞机（需要自行安装配置，并确保温度在其 log 中的位置）
func cpuTemperatureOnWindows() string {
	const defaultPath = `C:\Program Files (x86)\MSI Afterburner\HardwareMonitoring.hml`
	n := core.FilePath(engine.DataFolder(), filePath)
	if err := core.InitFile(&n, defaultPath); nil != err {
		kitten.Error(err)
		return err.Error()
	}
	p := n.GetPath(defaultPath)
	if err := os.Remove(p.String()); nil != err {
		kitten.Error(err)
		return err.Error()
	}
	<-time.NewTimer(1 * time.Second).C
	file, err := p.ReadBytes()
	if nil != err {
		kitten.Error(err)
		return err.Error()
	}
	return string(file[329:331]) // 此处为温度在微星小飞机 log 中的位置
}

// 返回状态等级
func getPerf(cpu float64, mem float64, ts string) int {
	ti, err := strconv.Atoi(ts)
	if nil != err {
		kitten.Warn(err)
		return 5
	}
	if 0 >= ti || 100 <= ti {
		return 5
	}
	perf := 0.00005 * (cpu + mem) * float64(ti)
	kitten.Debugf(`%s的负荷评分是 %f……`, zero.BotConfig.NickName[0], perf)
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

// 服务不可用时重启 sign.service
func signd() {
	// 处理 panic，防止程序崩溃
	defer func() {
		if err := recover(); nil != err {
			kitten.Error(`signd 守护协程出现错误喵！`, err, string(debug.Stack()))
		}
	}()
	process.GlobalInitMutex.Lock()
	process.GlobalInitMutex.Unlock()
	type delayer struct {
		t *time.Timer // 定时器
		v bool        // 是否有效
	}
	const sign = `http://127.0.0.1:8080`
	var (
		t     = time.NewTicker(5 * time.Second) // 每 5 秒检测一次
		u     = kitten.MainConfig().SelfID
		bot   = zero.GetBot(u.Int())
		delay delayer
	)
	for {
		<-t.C
		data, err := core.GETData(sign)
		kitten.Debug(string(data))
		if needRestart(err) {
			// 需要重启
			if delay.v {
				// 定时器有效，无需重复定时
				continue
			}
			// 定时器无效，延时 1 分钟执行
			delay = delayer{
				t: time.AfterFunc(time.Minute, func() {
					doRestart(bot, err)
				}),
				v: true,
			}
			kitten.Warnln(`开始进入重启倒计时……`, err)
			continue
		}
		// 不需要重启
		if delay.v {
			// 如定时器有效，尝试停止
			if !delay.t.Stop() {
				// 如定时器已经过期或停止，则将通道放空
				<-delay.t.C
			}
			// 定时器标记为无效
			delay.v = false
			kitten.Info(`重启倒计时已取消。`)
		}
	}
}

func needRestart(err error) bool {
	if nil == err {
		// 如果没有错误，则不需要重启
		return false
	}
	httpErr, ok := err.(*core.HTTPErr)
	if !ok {
		// 如果不是 HTTP 错误，则需要重启
		return true
	}
	// 如果是 HTTP 4xx 错误，则不需要重启，否则需要重启
	return http.StatusBadRequest > httpErr.StatusCode ||
		http.StatusInternalServerError <= httpErr.StatusCode
}

func doRestart(bot *zero.Ctx, err error) {
	// 执行重启
	kitten.Restart(`sign.service`)
	if !kitten.CheckCtx(bot, kitten.Caller) {
		// 没有 APICaller ，无法发送
		kitten.Error(`没有 APICaller ，无法发送喵！`, bot)
		return
	}
	// 延时 10 秒发送重启结果
	time.AfterFunc(10*time.Second,
		func() {
			bot.SendPrivateMessage(kitten.MainConfig().SuperUsers[0].Int(),
				fmt.Errorf(`检测到 sign.service 因 %w 离线，已经重启程序。`, err))
		})
}

// Ping
func ping(ctx *zero.Ctx) {
	pingURL := kitten.GetArgs(ctx)
	pg, err := probing.NewPinger(pingURL)
	if nil != err {
		kitten.SendWithImage(ctx, `哈——？.png`, err)
		return
	}
	pg.Count = 4                  // 检测 4 次
	pg.Timeout = 16 * time.Second // 超时时间设置
	var nbytes int
	pg.OnSend = func(pkt *probing.Packet) {
		nbytes = pkt.Nbytes
	}
	var pm strings.Builder
	pm.Grow(32 * pg.Count)
	pg.OnRecv = func(pkt *probing.Packet) {
		pm.WriteString(fmt.Sprintf(`来自 %s 的回复：字节=%d 时间=%dms TTL=%d
`, pkt.IPAddr, pkt.Nbytes, pkt.Rtt.Milliseconds(), pkt.TTL))
	}
	var r strings.Builder
	pm.Grow(32 + 32*pg.Count)
	pg.OnFinish = func(st *probing.Statistics) {
		r.WriteString(fmt.Sprintf(`正在 Ping %s [%s] 具有 %d 字节的数据：
%v
%s 的 Ping 统计信息：
数据包：已发送 = %d，已接收 = %d，丢失 = %d（%.0f%% 丢失），
`,
			pingURL, st.IPAddr, nbytes,
			&pm,
			st.IPAddr,
			st.PacketsSent, st.PacketsRecv, st.PacketsSent-st.PacketsRecv, st.PacketLoss))
		if 100 > st.PacketLoss {
			r.WriteString(fmt.Sprintf(`往返行程的估计时间：
最短 = %dms，最长 = %dms，平均 = %dms`,
				st.MinRtt.Milliseconds(), st.MaxRtt.Milliseconds(), st.AvgRtt.Milliseconds()))
		}
	}
	if err := pg.Run(); nil != err {
		kitten.SendWithImageFail(ctx, err)
		return
	}
	kitten.SendText(ctx, true, &r)
}
