package configs

import "github.com/spf13/viper"

func SetRedisDefaultConfig() {
	viper.SetDefault("REDIS.ADDR", "localhost:6379")
	viper.SetDefault("REDIS.DB", 0)
}
