package api

import (
	"github.com/melkior/sipware-go/message"
)

type Queue interface {
	Enqueue(message.Msg)
	Process(func(message.Msg))
	Dequeue() *message.Msg
}
