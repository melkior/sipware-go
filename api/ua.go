package api

import (
	"os"
	"github.com/melkior/sipware-go/message"
)

type Ua interface {
	Name() string
	Print()
	Wait()
	Open(string)
	Exit() chan os.Signal
	Connect(string)
	Done()
	Destroy(exiting bool)
	Read([]byte) (n int, err error)
	Write(message.Msg) (err error)
	Request(message.Msg, func(message.Msg) error) error
	Reply(message.Msg)
	Register(interface{})
}
