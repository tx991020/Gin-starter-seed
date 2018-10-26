package main

import (
	"fmt"
	"github.com/RichardKnop/machinery/example/tasks"
	"log"

	"errors"
	machinery "github.com/RichardKnop/machinery/v1"
	"github.com/RichardKnop/machinery/v1/config"
	"github.com/spf13/viper"
)

var (
	broker        string
	resultBackend string
	exchange      string
	exchangeType  string
	defaultQueue  string
	bindingKey    string

	cnf    config.Config
	server *machinery.Server
	worker *machinery.Worker
)

func init() {
	viper.SetConfigName("config") // no need to include file extension

	viper.AddConfigPath("/Users/andy/GoLang/src/doozer/api-server/")

	err := viper.ReadInConfig()

	if err != nil { // Handle errors reading the config file
		log.Fatal(err)
	}

	broker = viper.GetString("dozer.broker")
	resultBackend = viper.GetString("dozer.result_backend")
	exchange = viper.GetString("dozer.exchange")
	exchangeType = viper.GetString("dozer.exchange_type")
	defaultQueue = viper.GetString("dozer.default_queue")
	bindingKey = viper.GetString("dozer.binding_key")


	cnf = config.Config{
		Broker:        broker,
		ResultBackend: resultBackend,
		AMQP:          &config.AMQPConfig{Exchange: exchange, ExchangeType: exchangeType, BindingKey: bindingKey},
		DefaultQueue:  defaultQueue,
	}

	server, err := machinery.NewServer(&cnf)
	errors.New("Could not initialize server")

	// Register tasks
	tasks := map[string]interface{}{
		"add":      exampletasks.Add,
		"multiply": exampletasks.Multiply,
	}
	server.RegisterTasks(tasks)

	// The second argument is a consumer tag
	// Ideally, each worker should have a unique tag (worker1, worker2 etc)
	worker = server.NewWorker("machinery_worker",4)
}

func main() {
	err := worker.Launch()
	fmt.Println(err)
}
