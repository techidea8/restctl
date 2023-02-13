package wxapp

import (
	"fmt"

	"github.com/ArtisanCloud/PowerWeChat/src/kernel"
	"github.com/ArtisanCloud/PowerWeChat/v3/src/miniProgram"
)

type UserConfigOption func(*miniProgram.UserConfig)

func UseRedisCache(addr, password string) UserConfigOption {
	return func(cfg *miniProgram.UserConfig) {
		cfg.Cache = kernel.NewRedisClient(&kernel.RedisOptions{
			Addr:     addr,
			Password: password,
			DB:       0,
		})
	}
}
func InitializeMiniapp(conf WxConf) {
	app, err := miniProgram.NewMiniProgram(&miniProgram.UserConfig{
		AppID:     conf.AppId,  // 小程序appid
		Secret:    conf.Secret, // 小程序app secret
		HttpDebug: conf.HttpDebug,
		Log: miniProgram.Log{
			Level: conf.Log.Level,
			File:  conf.Log.File,
		},
		Cache: kernel.NewRedisClient(&kernel.RedisOptions{
			Addr:     conf.CacheConf.Addr,
			Password: conf.CacheConf.Password,
			DB:       conf.CacheConf.DbIndex,
		}),
	})

	if err != nil {
		fmt.Errorf(err.Error())
	} else {
		MiniProgramApp = app
	}
}
