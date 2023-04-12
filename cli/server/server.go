package main

import (
	"bufio"
	"context"
	"fmt"
	"github.com/half2me/antgo/ant"
	"github.com/redis/go-redis/v9"
	"log"
	"net"
	"time"
)

const (
	HOST = "localhost"
	PORT = "9999"
	TYPE = "tcp"
)

func main() {
	// redis://<user>:<pass>@localhost:6379/<db>
	opt, err := redis.ParseURL("redis://localhost:6379/0")
	if err != nil {
		panic(err)
	}

	client := redis.NewClient(opt)

	listen, err := net.Listen(TYPE, HOST+":"+PORT)
	if err != nil {
		log.Fatal(err)
	}
	// close listener
	defer listen.Close()

	// Write to Redis
	ch := make(chan ant.BroadcastMessage)
	go redisWriter(client, ch)

	// Server loop
	for {
		conn, err := listen.Accept()
		if err != nil {
			log.Fatal(err)
		}
		go handle(conn, ch)
	}
}

func handle(conn net.Conn, ch chan ant.BroadcastMessage) {
	log.Println("New connection!")
	defer conn.Close()
	r := bufio.NewReader(conn)

	for {
		msg, err := ant.ReadMsg(r)
		if err != nil {
			break
		}
		// skip non-broadcast messages
		if msg.Class() != ant.MESSAGE_TYPE_BROADCAST {
			continue
		}
		ch <- ant.BroadcastMessage(msg)

	}
	log.Println("Client disconnected!")
}

func redisWriter(client *redis.Client, ch chan ant.BroadcastMessage) {
	ctx := context.Background()
	for msg := range ch {
		// deduplicate: TODO: replace with lua script?
		hash := "dedup:" + string(hashAntMsg(msg))
		res := client.SetNX(ctx, "dedup:"+string(hash), "", time.Second*5).Val()
		if res {
			// not a duplicate, we can continue
			//log.Println(msg)

			key := fmt.Sprintf("ant:%d:%s", msg.DeviceNumber(), msg.DeviceTypeString())
			result := client.GetSet(ctx, key, []byte(msg)).Val() // TODO: replace with SetArgs to have TTL here
			if result != "" {
				old := ant.BroadcastMessage(result)
				switch msg.DeviceType() {
				case ant.DEVICE_TYPE_POWER:
					pow := ant.PowerMessage(msg)
					if pow.DataPageNumber() == 0x10 {
						fmt.Printf("%5d %.2fW\n", msg.DeviceNumber(), pow.AveragePower(ant.PowerMessage(old)))
					}
				case ant.DEVICE_TYPE_SPEED_AND_CADENCE:
					s := ant.SpeedAndCadenceMessage(msg)
					oldS := ant.SpeedAndCadenceMessage(old)
					cadence, _ := s.Cadence(oldS)
					speed, _ := s.Speed(oldS, 0.98)
					fmt.Printf("%5d %.2frpm %.2fm/s\n", msg.DeviceNumber(), cadence, speed)
				}
			}
		}
	}
}

func hashAntMsg(msg ant.BroadcastMessage) []byte {
	// content + extended content: device number [:2] & type [2]
	return append(msg.Content(), msg.ExtendedContent()[:3]...)
}
