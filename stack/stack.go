package stack

import (
	"fmt"
	"kitten/kitten"
	"strconv"
	"strings"
	"time"

	ctrl "github.com/FloatTech/zbpctrl"
	"github.com/FloatTech/zbputils/control"

	"gopkg.in/yaml.v3"

	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"

	log "github.com/sirupsen/logrus"
)

const (
	replyServiceName = "叠猫猫" // 插件名
)

var (
	kittenConfig = kitten.LoadConfig()
	stackConfig  = LoadConfig()
)

func init() {
	go AutoExit("stack/data.yaml", stackConfig)
	go AutoExit("stack/exit.yaml", stackConfig)

	help := strings.Join([]string{"发送",
		fmt.Sprintf("%s叠猫猫 [参数]", kittenConfig.CommandPrefix),
		"参数可选：加入|退出|查看",
		fmt.Sprintf("最多可以叠%d只猫猫哦", stackConfig.MaxStack),
		fmt.Sprintf("在叠猫猫队列中超过%d小时后，会自动退出", stackConfig.MaxTime),
		fmt.Sprintf("主动退出叠猫猫，需要%d小时后，才能再次加入", stackConfig.GapTime),
	}, "\n")
	// 注册插件
	engine := control.Register(replyServiceName, &ctrl.Options[*zero.Ctx]{
		DisableOnDefault: false,
		Help:             help,
	})

	engine.OnCommand("叠猫猫").Handle(func(ctx *zero.Ctx) {
		ag := ctx.State["args"].(string)
		data := LoadData("stack/data.yaml")
		dataExit := LoadData("stack/exit.yaml")
		switch ag {
		case "加入":
			In(data, dataExit, stackConfig, ctx)
		case "退出":
			Exit(data, dataExit, ctx)
		case "查看":
			View(data, ctx)
		default:
			ctx.SendGroupMessage(ctx.Event.GroupID, help)
		}
	})
}

// 加载叠猫猫配置
func LoadConfig() (stackConfig Config) {
	yaml.Unmarshal(kitten.FileRead("stack/config.yaml"), &stackConfig)
	return stackConfig
}

// 加载叠猫猫数据
func LoadData(path string) (stackData Data) {
	yaml.Unmarshal(kitten.FileRead(path), &stackData)
	return stackData
}

// 加入叠猫猫
func In(data Data, dataExit Data, stackConfig Config, ctx *zero.Ctx) {
	permit := true
	id := ctx.Event.UserID
	var report string

	for _, meow := range dataExit {
		if id == meow.Id {
			report = fmt.Sprintf("退出叠猫猫不足%d小时，不能加入喵！", stackConfig.GapTime)
			permit = false
			log.Info(strconv.FormatInt(id, 10) + report)
		}
	}
	for _, meow := range data {
		if id == meow.Id {
			report = "已经加入叠猫猫了喵！"
			permit = false
			log.Info(strconv.FormatInt(id, 10) + report)
		}
	}

	if permit {
		if len(data) >= stackConfig.MaxStack {
			log.Info(strconv.FormatInt(id, 10) + stackConfig.OutOfStack)
			report = stackConfig.OutOfStack
		} else {
			var meow Kitten
			meow.Id = id
			meow.Name = ctx.Event.Sender.Card
			if meow.Name == "" {
				meow.Name = ctx.Event.Sender.NickName
			}
			meow.Time = time.Unix(ctx.Event.Time, 0)
			data = append(data, meow)
			stackData, err1 := yaml.Marshal(data)
			err2 := kitten.FileWrite("stack/data.yaml", stackData)
			if !kitten.Check(err1) || !kitten.Check(err2) {
				report = "叠猫猫失败了喵！"
				log.Warn(strconv.FormatInt(id, 10) + report)
			} else {
				report = fmt.Sprintf("叠猫猫成功，目前处于队列中第%d位喵～", len(data))
				log.Info(strconv.FormatInt(id, 10) + report)
			}
		}
	}
	messageToSend := message.Text(report)
	ctx.SendChain(message.At(id), messageToSend)
}

// 退出叠猫猫
func Exit(data Data, dataExit Data, ctx *zero.Ctx) {
	dataNew := data
	for idx, meow := range data {
		if ctx.Event.UserID == meow.Id {
			dataNew = append(data[:idx], data[idx+1:]...)
		}
	}
	if len(dataNew) == len(data) {
		report := "没有加入叠猫猫，不能退出喵！"
		log.Warn(strconv.FormatInt(ctx.Event.UserID, 10) + report)
		messageToSend := message.Text(report)
		ctx.SendChain(message.At(ctx.Event.UserID), messageToSend)
	} else {
		stackData, err1 := yaml.Marshal(dataNew)
		err2 := kitten.FileWrite("stack/data.yaml", stackData)
		if !kitten.Check(err1) || !kitten.Check(err2) {
			report := "退出叠猫猫失败喵！"
			log.Warn(strconv.FormatInt(ctx.Event.UserID, 10) + report)
			messageToSend := message.Text(report)
			ctx.SendChain(message.At(ctx.Event.UserID), messageToSend)
		} else {
			report := "退出叠猫猫成功喵！"
			log.Info(strconv.FormatInt(ctx.Event.UserID, 10) + report)
			messageToSend := message.Text(report)
			ctx.SendChain(message.At(ctx.Event.UserID), messageToSend)

			// 记录至退出日志
			var meowExit Kitten
			meowExit.Id = ctx.Event.UserID
			meowExit.Time = time.Unix(ctx.Event.Time, 0)
			dataExit = append(dataExit, meowExit)
			exitData, err1 := yaml.Marshal(dataExit)
			err2 := kitten.FileWrite("stack/exit.yaml", exitData)
			if !kitten.Check(err1) || !kitten.Check(err2) {
				report := "记录至退出日志失败了喵！"
				log.Warn(strconv.FormatInt(meowExit.Id, 10) + report)
			} else {
				report := "记录至退出日志成功喵！"
				log.Info(strconv.FormatInt(meowExit.Id, 10) + report)
			}
		}
	}
}

// 查看叠猫猫
func View(data Data, ctx *zero.Ctx) {
	const report = "【叠猫猫队列】"
	dataString := Reverse(data)                                              // 反序查看
	reports := fmt.Sprintf("%s\n%s", report, strings.Join(dataString, "\n")) // 生成播报
	if len(data) <= 0 {
		reports = fmt.Sprintf("%s暂时没有猫猫哦", reports)
	}
	ctx.SendGroupMessage(ctx.Event.GroupID, reports)
}

// 叠猫猫队列反序并写为字符串数组
func Reverse(data Data) []string {
	var dataStringReverse []string
	for idx := len(data) - 1; idx >= 0; idx-- {
		dataStringReverse = append(dataStringReverse,
			fmt.Sprintf("%s（%d）", data[idx].Name, data[idx].Id))
	}
	return dataStringReverse
}

// 自动退出队列
func AutoExit(path string, config Config) {
	// 处理panic，防止程序崩溃
	defer func() {
		if err := recover(); !kitten.Check(err) {
			log.Error(err)
		}
	}()

	var limitTimeHours int
	switch path {
	case "stack/data.yaml":
		limitTimeHours = config.MaxTime
	case "stack/exit.yaml":
		limitTimeHours = config.GapTime
	}

	for {
		data := LoadData(path)
		dataNew := data
		limitTime, _ := time.ParseDuration(strconv.Itoa(limitTimeHours) + "h")
		nextTime := time.Now().Add(limitTime)
		if len(data) > 0 {
			if time.Since(data[0].Time).Hours() > float64(limitTimeHours) {
				if len(data) > 1 {
					nextTime = data[1].Time.Add(limitTime)
				}
				dataNew = data[1:]
			} else {
				nextTime = data[0].Time.Add(limitTime)
			}
		}
		if len(dataNew) != len(data) {
			stackData, err1 := yaml.Marshal(dataNew)
			err2 := kitten.FileWrite(path, stackData)
			if !kitten.Check(err1) || !kitten.Check(err2) {
				log.Warn(fmt.Sprintf("定时退出%s失败喵！", path))
			} else {
				log.Info(fmt.Sprintf("定时退出%s成功喵！", path))
			}
		}
		log.Info(fmt.Sprintf("下次定时退出%s时间为：%s", path, nextTime.Format("2006-01-02 15:04:05")))
		time.Sleep(time.Until(nextTime))
	}
}