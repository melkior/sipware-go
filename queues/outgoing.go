package queue

import (
	"sync"
	"github.com/melkior/sipware-go/api"
	"github.com/melkior/sipware-go/message"
)

type Outgoing struct {
	api.Queue
	lock sync.Mutex
	items []message.Msg
}

func (q *Outgoing) Enqueue(msg message.Msg) {
	q.lock.Lock()
	q.items = append(q.items, msg)
	q.lock.Unlock()
}

func (q *Outgoing) Next() *message.Msg {
	if len(q.items) > 0 {
		q.lock.Lock()
		msg, tail := q.items[0], q.items[1:]
		q.items = tail
		q.lock.Unlock()
		return &msg
	}
	return nil
}

func (q *Outgoing) Process(f func(msg message.Msg)) {
	if len(q.items) > 0 {
		q.lock.Lock()
		msg, tail := q.items[0], q.items[1:]
		q.items = tail
		q.lock.Unlock()
		f(msg)
	}
}

func (q *Outgoing) Dequeue() *message.Msg {
	if len(q.items) > 0 {
		q.lock.Lock()
		msg, tail := q.items[0], q.items[1:]
		q.items = tail 
		q.lock.Unlock()
		return &msg
	}

	return nil
}
