package config

import (
	"fmt"
	"os"

	"github.com/spf13/viper"
)

type Config struct {
	ServerHost   string
	ServerPort   string
	ElasticHost  string
	ElasticPort  string
	ElasticIndex string
}

func ReadInConfig() Config {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("./configs/")
	err := viper.ReadInConfig()
	if err != nil {
		fmt.Printf("Fatal error config file: %s \n", err.Error())
		os.Exit(1)
	}

	return Config{
		ElasticHost:  viper.GetString("elasticsearch.host"),
		ElasticPort:  viper.GetString("elasticsearch.port"),
		ElasticIndex: viper.GetString("elasticsearch.index"),
		ServerHost:   viper.GetString("server.host"),
		ServerPort:   viper.GetString("server.port"),
	}
}
