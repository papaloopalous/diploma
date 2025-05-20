package configloading

import (
	"fmt"

	"load_balancer/internal/messages"

	"github.com/spf13/viper"
)

const (
	ServerAddr   = "server.address"
	BackendAddrs = "backends"
	Interval     = "interval"
	DBAddr       = "db.address"
	MaxTokens    = "maxTokens"
	Rate         = "rate"
	Salt         = "salt"
)

func LoadConfig() error {
	viper.SetConfigFile("./config/config.yaml")
	err := viper.ReadInConfig()

	if err != nil {
		return fmt.Errorf(messages.ErrReadConfig, err)
	}

	return nil
}

func SetParams() (serverAddr string, backendAddrs []string, interval int, dbAddr string, salt string, maxTokens, rate int) {
	serverAddr = viper.GetString(ServerAddr)
	backendAddrs = viper.GetStringSlice(BackendAddrs)
	interval = viper.GetInt(Interval)
	dbAddr = viper.GetString(DBAddr)
	salt = viper.GetString(Salt)
	maxTokens = viper.GetInt(MaxTokens)
	rate = viper.GetInt(Rate)
	return serverAddr, backendAddrs, interval, dbAddr, salt, maxTokens, rate
}
