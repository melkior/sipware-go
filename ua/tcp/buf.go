package tcpua

import (
	// "os"
	"fmt"
	"sync"
	"bytes"
	"strings"
	"net/mail"
	"unicode/utf8"
	"github.com/melkior/sipware-go/message"
)

type SipwareTcpBuf struct {
	len int
	cap int
	idx int
	chunkLen int
	cur *message.Msg
	state ParserState
	pool sync.Pool
	chunks []SipwareTcpChunk
	msgs []message.Msg
}

func (q *SipwareTcpBuf) CreatePool(chunkLen int) error {
	q.chunkLen = chunkLen
	q.pool = sync.Pool {
		New: func() any {
			return make([]byte, chunkLen)
		},
	}

	return nil
}

func (q *SipwareTcpBuf) NewSlice() []byte {
	return q.pool.Get().([]byte)
}

func (q *SipwareTcpBuf) NewChunk() SipwareTcpChunk {
	// buff := q.pool.Get().([]byte)
        buff := make([]byte, q.chunkLen)

	return SipwareTcpChunk{buff: buff}
}

func (q *SipwareTcpBuf) DestroyChunk(chunk SipwareTcpChunk) {
	fmt.Println("Destroy chunk")
	chunk.offset = 0
	chunk.written = 0
	// q.pool.Put(chunk.buff)
}

func (q *SipwareTcpBuf) Push(chunk SipwareTcpChunk) {
	q.chunks = append(q.chunks, chunk)
}

func (q *SipwareTcpBuf) Shift() SipwareTcpChunk {
	first, chunks := q.chunks[0], q.chunks[1:]

	if(first.offset < first.written) {
		delta := first.written - first.offset
		chunk := q.NewChunk()
		buffer := first.buff[first.offset:first.written]

		for i := 0; i < delta; i++ {
			chunk.buff[i] = buffer[i]
		}
		q.chunks = append([]SipwareTcpChunk{chunk}, chunks...)
	} else {
		q.chunks = chunks
	}

	q.DestroyChunk(first)
	return first
}

func (q *SipwareTcpBuf) parseHeaders(data []byte) (message.Msg, error) {
        reader := strings.NewReader(string(data))
        mail, err := mail.ReadMessage(reader)

        if err != nil {
                fmt.Println("Parse header error", err)
                return message.Msg{}, err
        }

        msg, err := message.NewMsg(mail.Header, "")
        return msg, err
}

func (q *SipwareTcpBuf) readHeaders(buf []byte) (message.Msg, error) {
	msg, err := q.parseHeaders(buf)
	if err != nil {
		return message.Msg{}, err
	}

	return msg, nil
}

func (q *SipwareTcpBuf) nextContent() (message.Msg, error) {
	if len(q.chunks) == 0 {
		fmt.Println("Next content empty chunks", q.chunks)
		return message.Msg{}, msgParseErr
		// os.Exit(4)
	}

	msg := q.cur

	if msg == nil {
		return message.Msg{}, msgParseErr
	}

	i := 0

	for i < len(q.chunks) {
		chunk := &q.chunks[i]

		if(msg.ContentLength > 0) {
			length := msg.BodyLen()
			delta := chunk.written - chunk.offset

			if length + delta >= msg.ContentLength {
				content := chunk.buff[chunk.offset:chunk.offset + msg.ContentLength - length]
				chunk.offset = chunk.offset + msg.ContentLength - length
				msg.AddContent(content)
			} else {
				content := chunk.buff[chunk.offset:chunk.written]
				chunk.offset = chunk.written
				msg.AddContent(content)
			}

			length = msg.BodyLen()

			if chunk.offset >= chunk.written {
				q.Shift()
				i--
			}

			if length >= msg.ContentLength {
				q.cur = nil
				q.state = PARSER_HEADER
				return *msg, nil
			}
		}
		i++
	}

	return message.Msg{}, msgParseErr
}

func (q *SipwareTcpBuf) nextHeaders() (message.Msg, error) {
	if len(q.chunks) == 0 {
		fmt.Println("Next Headers empty chunks", q.chunks)
		return message.Msg{}, msgParseErr
	}

	chunk := &q.chunks[0]
	idx, offset := find(chunk.buff[chunk.offset:chunk.written], []byte("\r\n\r\n"))

	if idx < 0 || offset < 0 {
		return message.Msg{}, msgParseErr
	}

	msg, err := q.readHeaders(chunk.buff[chunk.offset:chunk.offset + offset])

	if err != nil {
		return msg, err
	}

	chunk.offset = chunk.offset + offset

	if msg.ContentLength == 0 {
		q.cur = nil

		if chunk.offset >= chunk.written {
			q.Shift()
		}

		return msg, nil
	}

	if msg.ContentLength > 0 {
		q.cur = &msg
		q.state = PARSER_CONTENT
		return q.nextContent()
	}

	return msg, nil
}

func (q *SipwareTcpBuf) next() (message.Msg, error) {
	if q.state == PARSER_HEADER {
		return q.nextHeaders()
	} else if q.state == PARSER_CONTENT {
		return q.nextContent()
	} else {
		fmt.Println("Next wrong state", q.state)
		return message.Msg{}, parserStateErr
	}
}

func (q *SipwareTcpBuf) Read() ([]message.Msg, error) {
	fmt.Println("Read")
	msgs := []message.Msg{}

	for {
		msg, err := q.next()

		if err != nil {
			return msgs, err
		} else {
			msgs = append(msgs, msg)
		}
	}
	return msgs, nil
}

// bytes.explode
func explode(s []byte, n int) [][]byte {
        if n <= 0 || n > len(s) {
                n = len(s)
        }
        a := make([][]byte, n)
        var size int
        na := 0
        for len(s) > 0 {
                if na+1 >= n {
                        a[na] = s
                        na++
                        break
                }
                _, size = utf8.DecodeRune(s)
                a[na] = s[0:size:size]
                s = s[size:]
                na++
        }
        return a[0:na]
}

// func split(s, sep []byte, n int) (int, int) {
func find(s, sep []byte) (int, int) {
	/*
	if n == 0 {
		return -1, -1
	}
	if n < 0 {
		n = bytes.Count(s, sep) + 1
	}
	*/

        idx := bytes.Index(s, sep)

	if idx < 0 {
		return -1, -1
	}

	return idx, idx + len(sep)
}
