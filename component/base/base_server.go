package base

import (
	utils "com.minigame.utils"
	"context"
	"encoding/json"
	"fmt"
	"github.com/go-redis/redis"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
	"math/rand"
	"os"
	"runtime/debug"
	"time"

	"com.minigame.component/amqp/rabbitmq"
	"com.minigame.component/db"
	"com.minigame.component/log"
)

var tag = "BaseServer"

type BaseServer struct {
	Context context.Context
	Fw      *Framework
	conf    *config
	Pro     Process
	Vipper  *viper.Viper
	RedisDb *redis.Client
	MysqlDb *db.MysqlDb
	MQConn  *rabbitmq.AmqpConn
}

func (s *BaseServer) Init(fw *Framework) error {
	rand.Seed(time.Now().UnixNano())
	s.Fw = fw

	s.Context = context.Background()
	s.conf = new(config)
	err := s.Pro.Initialize("", "")
	if err != nil {
		return err
	}
	var confStr string
	confStr, err = s.loadConf()
	if err != nil {
		return err
	}

	if s.conf.Log.Path != "" {
		s.Pro.DefaultLogDir = s.conf.Log.Path
	}

	err = log.InitLogger(s.Pro.Name, s.Pro.DefaultLogDir)
	if err != nil {
		return err
	}
	if s.conf.Log.Out != "" {
		err = log.ChooseLog(s.conf.Log.Out)
		if err != nil {
			return err
		}
	}

	if s.conf.Redis.Addr != "" {
		s.RedisDb, err = db.NewRedisClient(s.conf.Redis.Addr, s.conf.Redis.Password, s.conf.Redis.Db)
		if err != nil {
			return err
		}
	}
	if s.conf.Mysql.Ip != "" && s.conf.Mysql.User != "" {
		s.MysqlDb, err = db.NewMysqlClient(s.conf.Mysql.User, s.conf.Mysql.Password, s.conf.Mysql.Ip, s.conf.Mysql.Port,
			s.conf.Mysql.DbName)
		if err != nil {
			return err
		}
	}
	s.MQConn, err = rabbitmq.NewConn(context.Background(), &rabbitmq.RabbitInitConfig{
		UserName: "guest",
		Password: "guest",
		Host:     "127.0.0.1",
		Port:     5672,
	})
	if err != nil {
		return err
	}

	log.Infof(tag, "[Server] started successful!!! \nprocess: %+v\nconf: %s\n", s.Pro, confStr)
	return nil
}

func (s *BaseServer) loadConf() (confStr string, err error) {
	s.Vipper = viper.New()
	s.Vipper.AddConfigPath(s.Pro.DefaultCfgDir)
	s.Vipper.SetConfigName("server")
	s.Vipper.SetConfigType("toml")
	err = s.Vipper.ReadInConfig()
	err = errors.WithStack(err)
	if err != nil {
		return
	}
	err = s.Vipper.Unmarshal(s.conf)
	err = errors.WithStack(err)
	if err != nil {
		return
	}

	m := make(map[string]interface{})
	err = s.Vipper.Unmarshal(&m)
	err = errors.WithStack(err)
	if err != nil {
		return
	}
	marshal, err := json.Marshal(m)
	if err != nil {
		return
	}
	confStr = string(marshal)

	if s.conf.Log.Out == "" {
		s.conf.Log.Out = log.TERM
	}
	if s.conf.Redis.Addr == "" {
		s.conf.Redis.Addr = "127.0.0.1:6376"
	}
	//s.conf.Redis.Db %= 15
	return
}

func (s *BaseServer) CatchError(err interface{}) {
	if log.GetLevel() == log.DEBUG {
		//debug模式，写文件
		MainPanicToFile(err, s.Pro.DefaultLogDir, false)
	} else {
		//release模式，恢复
		fmt.Fprintln(os.Stderr, fmt.Sprintf("======catch error begin [%v]======", time.Now().Format(utils.TIME_FORMAT)))
		fmt.Fprintln(os.Stderr, err)
		//打印堆栈
		debug.PrintStack()
		fmt.Fprintln(os.Stderr, fmt.Sprintf("======catch error end   [%v]======", time.Now().Format(utils.TIME_FORMAT)))
		os.Stderr.Sync()
	}
}

func (s *BaseServer) MainLoop() {
	select {}
}

func (s *BaseServer) OnReload() {
	// TODO implement me
}

func (s *BaseServer) OnExit() {
	// TODO implement me
}

func (s *BaseServer) SetTag(str string) {
	tag = fmt.Sprintf("%s.(%s)", "BaseServer", str)
}
