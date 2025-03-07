package queue

import (
	"fmt"
	"sync"
	"github.com/melkior/sipware-go/api"
	"github.com/melkior/sipware-go/message"
)

type Active struct {
	sipware.Queue
	lock sync.Mutex
	items []message.Msg
}

func (q *Active) Enqueue(msg message.Msg) {
	q.lock.Lock()
	q.items = append(q.items, msg)
	q.lock.Unlock()
}

func (q *Active) Next() *message.Msg {
	if len(q.items) > 0 {
		q.lock.Lock()
		msg, tail := q.items[0], q.items[1:]
		q.items = tail
		q.lock.Unlock()
		return &msg
	}
	return nil
}

func (q *Active) Process(f func(msg message.Msg)) {
	if len(q.items) > 0 {
		q.lock.Lock()
		msg, tail := q.items[0], q.items[1:]
		q.items = tail
		q.lock.Unlock()
	}
}

func (q *Active) Dequeue() *message.Msg {
	if len(q.items) > 0 {
		q.lock.Lock()
		msg, tail := q.items[0], q.items[1:]
		q.items = tail 
		q.lock.Unlock()
		return &msg
	}
	return nil
}
