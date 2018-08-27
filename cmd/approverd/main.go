package main

import (
	"flag"
	"log"
	"os"
	"time"

	"github.com/sjbodzo/review_system/queue"
)

var redisflags struct {
	pollSeconds   int
	reqQueueName  string
	procQueueName string
	endpoint      string
	port          int
}

func init() {
	flag.IntVar(&redisflags.port, "redisPort", 6379, "Port to connect to database with")
	flag.IntVar(&redisflags.pollSeconds, "pollSeconds", 5, "How many seconds to wait between polling")
	flag.StringVar(&redisflags.endpoint, "redisEndpoint", "", "Database endpoint to connect to")
	flag.StringVar(&redisflags.procQueueName, "redisProcQueueName", "proc_queue",
		"Name of redis queue to stage product reviews in while being reviewed")
	flag.StringVar(&redisflags.reqQueueName, "redisReqQueueName", "req_queue",
		"Name of redis queue where new or retried product review jobs go")
	flag.Parse()
}

func main() {
	f, err := os.OpenFile("/tmp/approverd.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatal("Unable to open log file for writing")
	}
	defer f.Close()
	log.SetOutput(f)

	if err := run(); err != nil {
		log.Fatalln(err)
	}
}

func run() error {
	pool := queue.NewWorkerPool(redisflags.endpoint, redisflags.port)
	ticker := time.NewTicker(time.Duration(redisflags.pollSeconds) * time.Second)
	for range ticker.C {
		go func() {
			err := pool.ProcessNextReview(redisflags.reqQueueName, redisflags.procQueueName)
			if err != nil {
				log.Println("Error:", err.Error())
			}
		}()
	}
	return nil
}
