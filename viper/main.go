package main

import (
	"log"
	"time"

	"fmt"
	"github.com/fsnotify/fsnotify"
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
)

var Config *viper.Viper

func main() {
	r := gin.Default()
	Config = viper.New()
	Config.SetConfigName("config")
	Config.AddConfigPath("/tmp")
	if err := Config.ReadInConfig(); err != nil {
		panic(err)
	}

	go func() {
		for {
			time.Sleep(time.Second * 5)
			Config.WatchConfig()
			watch := func(e fsnotify.Event) {
				log.Printf("Config file is changed: %s n", e.String())
				InitConf(Config.Get("cache"))

			}
			Config.OnConfigChange(watch)

		}
	}()

	r.GET("/", func(c *gin.Context) {

		data := Config.GetString("cache")

		c.JSON(200, data)
	})
	r.Run(":4000")
}

func InitConf(x interface{}) {
	fmt.Println(x)
}
