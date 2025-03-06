package queue

import (
	"fmt"
	"sync"
	"sipware/api"
	"sipware/message"
)

type Incoming struct {
	sipware.Queue
	lock sync.Mutex
	items []message.Msg
}

func (q *Incoming) Enqueue(msg message.Msg) {
	q.lock.Lock()
	q.items = append(q.items, msg)
	q.lock.Unlock()
}

func (q *Incoming) Next() *message.Msg {
	if len(q.items) > 0 {
		q.lock.Lock()
		msg, tail := q.items[0], q.items[1:]
		q.items = tail
		q.lock.Unlock()
		return &msg
	}
	return nil
}

func (q *Incoming) Process(f func(msg message.Msg)) {
	if len(q.items) > 0 {
		q.lock.Lock()
		msg, tail := q.items[0], q.items[1:]
		q.items = tail
		q.lock.Unlock()
		f(msg)
	}
}

func (q *Incoming) Dequeue() *message.Msg {
	if len(q.items) > 0 {
		q.lock.Lock()
		msg, tail := q.items[0], q.items[1:]
		q.items = tail 
		q.lock.Unlock()
		return &msg
	}
	return nil
}
