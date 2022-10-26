package main

import (
	"bytes"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/go-kit/log"
	_ "github.com/lib/pq"
	"io/ioutil"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"testapp/internal/service"
	authdb "testapp/internal/store"
	httpTransport "testapp/internal/transport/http"
	"time"
)

func CreateKafkaConsumer(groupID string) *kafka.Consumer {

	consumer, err := kafka.NewConsumer(&kafka.ConfigMap{
		"bootstrap.servers": "localhost",
		"group.id":          groupID,
		"auto.offset.reset": "earliest",
	})

	if err != nil {
		panic(err)
	}

	return consumer
}

func SubscribeTopic(consumer *kafka.Consumer, topics []string) {
	//consumer.Subscribe(topic,nil)
	consumer.SubscribeTopics(topics, nil)
	fmt.Println("Subscribed!")
}

func ReadTopicMessages(consumer *kafka.Consumer) string {

	var message string
	for {
		msg, err := consumer.ReadMessage(-1)
		if err == nil {
			fmt.Printf("Message on %s: %s, key:%s \n", msg.TopicPartition, string(msg.Value), string(msg.Key))
			//message = message + string(msg.Value)
		} else {
			// The client will automatically try to recover from all errors.
			fmt.Printf("Consumer error: %v (%v)\n", err, msg)
		}
	}

	CloseConsumer(consumer)

	return message
}

func CloseConsumer(consumer *kafka.Consumer) {
	consumer.Close()
}

func RegisterConnector() *http.Response {
	plan, _ := ioutil.ReadFile("./connectors/debezium-connector.json")
	response, err := http.Post("http://localhost:8083/connectors/", "application/json", bytes.NewBuffer(plan))

	if err != nil {
		panic(err)
	}

	return response
}

func CheckConnector() {
	response, err := http.Get("http://localhost:8083/connectors/product_connector")
	defer response.Body.Close()

	if err != nil {
		panic(err)
	}
	if response.StatusCode != 200 {
		RegisterConnector()
	}

	// show full block.
	//body, _ := ioutil.ReadAll(response.Body)
	//fmt.Println(string(body))
}

func main() {
	var logger log.Logger
	{
		logger = log.NewLogfmtLogger(os.Stdout)
		logger = log.With(logger, "time", log.DefaultTimestampUTC)
	}

	//postgresql connection
	const (
		host     = "localhost"
		port     = 5432
		user     = "admin"
		password = "secret"
		dbname   = "auth"
	)
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s "+
		"password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)
	db, err := sql.Open("postgres", psqlInfo)
	sqlC := authdb.New(db)

	if err != nil {
		panic(err)
	}
	defer db.Close()
	//========================================

	//initialize
	var s service.Service
	{
		s = service.NewService(logger, sqlC)
	}

	var handler http.Handler
	{
		handler = httpTransport.MakeHTTPHandler(log.With(logger, "transport", "http"), s)
	}

	var httpServer *http.Server
	{
		httpServer = &http.Server{
			Addr:    "0.0.0.0:8081",
			Handler: handler,
		}
	}
	//================================================================

	//serving
	errs := make(chan error)
	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
		errs <- errors.New((<-c).String())
	}()

	go func() {
		_ = logger.Log("transport", "http", "address", "0.0.0.0:8081")

		err = httpServer.ListenAndServe()
		if errors.Is(err, http.ErrServerClosed) {
			errs <- err
		}
	}()
	//================================================================
	consumer := CreateKafkaConsumer("auth")
	CheckConnector()

	topics := make([]string, 0)
	topics = append(topics, "auth.events")
	SubscribeTopic(consumer, topics)
	go ReadTopicMessages(consumer)

	//shutdown
	err = <-errs
	_ = logger.Log("error", err.Error())

	ctx, cf := context.WithTimeout(context.Background(), time.Second*10)
	defer cf()
	if err := httpServer.Shutdown(ctx); err != nil {
		_ = logger.Log("error", err.Error())
	}
	//================================================================

}
