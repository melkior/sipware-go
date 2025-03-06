package api

import (
	"sipware/message"
)

type Queue interface {
	Enqueue(message.Msg)
	Process(func(message.Msg))
	Dequeue() *message.Msg
}
