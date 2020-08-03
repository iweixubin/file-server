package internal

import (
	"flag"
	"github.com/spf13/viper"
	"os"
)

var Cfg *Config

type Config struct {
	WebEngine string
	Host      string
	SrcDir    string

	SupportExtResize []string
	SizeRegexp       string
	SizeLimits       []SizeLimit

	HttpExpires            int64
	CacheMode              CacheMode
	CacheModeDiskDir       string
	CacheModeMemoryExpires int64
	CacheModeRedisConn     string
	CacheModeRedisExpires  int64
}

type SizeLimit struct {
	Width  int
	Height int
}

func init() {
	initViper()
	initCfg()
	initSizeRegexp()
	initCache()
}

func initCfg() {
	Cfg = &Config{}

	Cfg.WebEngine = viper.GetString("WebEngine")
	Cfg.Host = viper.GetString("Host")
	Cfg.SrcDir = viper.GetString("SrcDir")

	Cfg.SupportExtResize = viper.GetStringSlice("SupportExtResize")
	Cfg.SizeRegexp = viper.GetString("SizeRegexp")
	sizeLimit := viper.GetIntSlice("SizeLimit")
	if len(sizeLimit) != 0 {
		if len(sizeLimit)%2 != 0 {
			panic("配置文件中的 SizeLimit 不是偶数个！")
		}

		var sl SizeLimit
		for i, v := range sizeLimit {
			if i%2 == 0 {
				sl = SizeLimit{}
				sl.Width = v
			} else {
				sl.Height = v
				Cfg.SizeLimits = append(Cfg.SizeLimits, sl)
			}
		}
	}

	Cfg.HttpExpires = viper.GetInt64("HttpExpires")
	switch viper.GetString("CacheMode") {
	case "none":
		Cfg.CacheMode = CacheModeNone
	case "disk":
		Cfg.CacheMode = CacheModeDisk
	case "memory":
		Cfg.CacheMode = CacheModeMemory
	case "redis":
		Cfg.CacheMode = CacheModeRedis
	default:
		println("CacheMode 没有设置规定的值：none、disk、memory、redis")
	}

	Cfg.CacheModeDiskDir = viper.GetString("CacheModeDisk.Dir")

	Cfg.CacheModeMemoryExpires = viper.GetInt64("CacheModeMemory.Expires")

	Cfg.CacheModeRedisConn = viper.GetString("CacheModeRedis.Conn")
	Cfg.CacheModeRedisExpires = viper.GetInt64("CacheModeRedis.Expires")

}

func initViper() {
	var configPath string

	flag.StringVar(&configPath, "c", "", "The config file path")
	flag.Parse()

	if configPath != "" {
		viper.SetConfigFile(configPath)
	} else {
		viper.SetConfigName("config")
		viper.SetConfigType("yaml")
		viper.AddConfigPath(".")
		viper.AddConfigPath(os.Getenv("FileServerConfigPath"))
	}

	if e := viper.ReadInConfig(); e != nil {
		panic(e.Error())
	}
}
