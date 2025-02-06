package logging

import "sync"

var messageQueue chan string
var messageQueueOnce sync.Once

func GetMessageQueue() chan string {
	messageQueueOnce.Do(func() {
		messageQueue = make(chan string, 8)
	})

	return messageQueue
}

func SendMessage(s string) {
	GetMessageQueue() <- s
}
