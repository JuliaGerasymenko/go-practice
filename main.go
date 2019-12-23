package main

import (
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"fmt"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"log"
	"strings"
	"time"
)

type Payload struct {
	Time   string `json:"time"`
	Wisdom string `json:"wisdom"`
	Secret string `json:"secret"`
	Team   string `json:"team"`
}

func main() {
	opts := mqtt.NewClientOptions().
		AddBroker(fmt.Sprintf("tcp://%s:1883", "localhost"))
	client := mqtt.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		panic(token.Error())
	}

	wait := make(chan []string)
	var num byte
	var secret string
	s1 := make([]string, 4)

	client.Subscribe("/test/inception", 0, func(client mqtt.Client, message mqtt.Message) {
		var m Payload
		if err := json.Unmarshal(message.Payload(), &m); err != nil {
			return
		}
		secret = m.Secret
		if strings.Index(m.Wisdom, ":") != -1 {
			num = m.Wisdom[0] - 48
			s1[num-1] = m.Wisdom[3:]
		}

		for i := 0; i < 4; i++ {
			if s1[i] == "" {
				break
			}
			if i == 3 {
				wait <- s1
			}
		}
	})

	arr := <-wait
	var message Payload

	message.Wisdom = strings.Join(arr, " ")

	message.Team = "inna_vlad_julia the best"

	hasher := sha1.New()
	hasher.Write([]byte(secret + " inna_vlad_julia"))
	bs := hasher.Sum(nil)
	message.Secret = string(hex.EncodeToString(bs))

	message.Time = time.Now().String()

	data, err := json.Marshal(&message)
	if err != nil {
		log.Fatal(err)
	}
	timer := time.NewTicker(2 * time.Second)
	client.Publish("/test/result", 0, false, data)
	<-timer.C
}
