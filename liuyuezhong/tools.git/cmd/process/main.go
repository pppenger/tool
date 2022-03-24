package main

import (
	"flag"
	"fmt"
	"github.com/go-redis/redis"
	"log"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"sync/atomic"
	"syscall"
	"time"
)

var (
	exit       int32
	del        bool
	batch      int
	host       string
	port       int
	passwd     string
	expire     int
	db         int
	cluster    bool
	nodeId     string
	skipPrefix string
	debug      bool
)

func init() {
	log.SetFlags(log.Lshortfile | log.LstdFlags)

	flag.StringVar(&host, "h", "127.0.0.1", "redis server host")
	flag.IntVar(&port, "p", 6379, "redis server port")
	flag.IntVar(&db, "db", 0, "redis db")
	flag.StringVar(&passwd, "a", "", "redis server password")
	flag.BoolVar(&cluster, "c", false, "is redis cluster")
	flag.StringVar(&nodeId, "node", "", "redis cluster node only")
	flag.IntVar(&batch, "b", 100, "batch to process")
	flag.IntVar(&expire, "e", 0, "expire in seconds")
	flag.BoolVar(&del, "d", false, "to del keys")
	flag.StringVar(&skipPrefix, "skip", "", "skip process key prefix")
	flag.BoolVar(&debug, "debug", false, "debug mode")
}

func Process(c *redis.Client, keys []string) {
	if len(keys) == 0 {
		return
	}

	if atomic.LoadInt32(&exit) == 1 {
		return
	}

	if del {
		log.Println("del keys: ", keys)
		if err := c.Del(keys...).Err(); err != nil {
			log.Fatalf("del err: %v", err)
		}

		return
	}

	if expire > 0 {
		for _, key := range keys {
			duration, err := c.TTL(key).Result()
			if err != nil {
				if err != redis.Nil {
					log.Fatalf("ttl key[%s] err: %v", key, err)
				}
				continue
			}

			if duration == -time.Second {
				log.Println("expire key: ", key)
				if err := c.Expire(key, time.Duration(expire)*time.Second).Err(); err != nil {
					log.Fatalf("expire key[%s] err: %v", key, err)
				}
			}
		}

		return
	}

	log.Println("hit keys: ", keys)
}

func DoJob(c *redis.Client, prefix, node string) {
	const scanPer = 100
	const outputPer = 10 * 10000

	iterator := uint64(0)
	count := 1
	outputCounter := uint64(0)
	cursor := uint64(0)
	argArray := []interface{}{
		"scan", cursor, "match", fmt.Sprintf("%s*", prefix), "count", scanPer,
	}
	if node != "" {
		argArray = append(argArray, node)
	}

	buff := make([]string, 0)
	for atomic.LoadInt32(&exit) == 0 {
		value, err := c.Do(argArray...).Result()
		if err != nil {
			fmt.Println(err)
			break
		}
		values := value.([]interface{})
		cursor, err = strconv.ParseUint(values[0].(string), 10, 64)
		if err != nil {
			log.Fatalf("invalid cursor[%s]: %v", values[0].(string), err)
		}
		keys := values[1].([]interface{})

		iterator += scanPer
		outputCounter += scanPer
		if debug {
			if outputCounter >= outputPer {
				outputCounter = 0
				log.Printf("scan iterator[%v] cursor[%v]\n", iterator, cursor)
			}
		}

		if len(keys) > 0 {
			for _, key := range keys {
				if skipPrefix != "" {
					if strings.HasPrefix(key.(string), skipPrefix) {
						continue
					}
				}

				buff = append(buff, key.(string))
			}

			count += len(buff)
			if len(buff) >= batch {
				Process(c, buff)
				buff = buff[0:0]
			}
		}

		if cursor == 0 {
			break
		}
		argArray[1] = cursor
	}
	Process(c, buff)
}

func main() {
	flag.Parse()

	args := flag.Args()
	if len(args) == 0 {
		log.Fatalf("prefix argument is empty")
	}

	client := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", host, port),
		Password: passwd,
		DB:       db,
	})

	if err := client.Ping().Err(); err != nil {
		log.Fatalf("redis ping test error: %v", err)
	}

	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGHUP, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT)
		for {
			s := <-sigChan
			switch s {
			case syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT:
				atomic.StoreInt32(&exit, 1)
				log.Println("exit...")
				return
			case syscall.SIGHUP:
			default:
			}
		}
	}()

	prefix := args[0]
	log.Println("handle prefix: ", prefix)

	// 指定节点
	if nodeId != "" {
		log.Printf("scan node[%s]\n", nodeId)
		DoJob(client, prefix, nodeId)
		return
	}

	// 非集群
	if !cluster {
		DoJob(client, prefix, "")
		return
	}

	nodes := make([]string, 0)
	result, err := client.Do("cluster", "nodes").Result()
	if err != nil {
		log.Fatalf("get cluster node err: %v", err)
	}

	lines := strings.Split(result.(string), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		info := strings.SplitN(line, " ", 4)
		if len(info) != 4 {
			log.Fatalf("invalid node info: %v", info)
		}

		if strings.Contains(info[2], "master") {
			nodes = append(nodes, info[0])
		}
	}

	log.Printf("get %d nodes\n", len(nodes))
	for _, node := range nodes {
		log.Printf("scan node[%s]\n", node)
		DoJob(client, prefix, node)
	}
}
