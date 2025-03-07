package trans

import (
	"sync"
	"github.com/melkior/sipware-go/message"
)

type SipwareTr struct {
        wg *sync.WaitGroup
        ch chan message.Msg
}

func New() SipwareTr {
	tr := SipwareTr {
		wg: &sync.WaitGroup{},
		ch: make(chan message.Msg),
	}
	return tr
}

func (tr SipwareTr) Add(i int) {
	tr.wg.Add(i)
}

func (tr SipwareTr) Done() {
	tr.wg.Done()
}

func (tr SipwareTr) Wait() {
	tr.wg.Wait()
}

func (tr SipwareTr) Read() chan message.Msg {
	return tr.ch
}

func (tr SipwareTr) Write(m message.Msg) {
	tr.ch <- m
}

