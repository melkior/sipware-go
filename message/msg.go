package message

import (
	"fmt"
	"time"
        "mime"
	"strconv"
	"net/mail"
)

type Msg struct {
	Code int
	Date time.Time
	ContentLength int
	bodyLength int
	Id string
	Hop []string
	Method string
	MessageId string
	ContentType string

	headers map[string][]string
	body *Body
}

func (msg *Msg) Print() {
	fmt.Println("Code:", msg.Code);
	fmt.Println("Id:", msg.Id);
	fmt.Println("Hop:", msg.Hop);
	fmt.Println("Date:", msg.Date);
	fmt.Println("Method:", msg.Method);
	fmt.Println("Message id:", msg.MessageId);
	fmt.Println("Content Length:", msg.ContentLength);
	fmt.Println("Content Type:", msg.ContentType);
	fmt.Println("Headers:")

	for hdr, val := range msg.headers {
		fmt.Println(hdr, val)
	}
}

func (m *Msg) Get(header string) []string {
	return m.headers[header]
}

func (m *Msg) GetHeaders() map[string][]string {
	return m.headers
}

func (m *Msg) Set(header string, value string) {
	m.headers[header] = []string{value}
	// m.headers[header] := [value]
}

func (m *Msg) Add(header string, value string) {
	fmt.Println("header", header)
	fmt.Println("value", value)
	hdr := m.headers[header]

	if(len(hdr) > 0) {
		m.headers[header] = append(hdr, value)
	} else {
		m.headers[header] = []string{value}
	}
}

func (m *Msg) Body() []byte {
	return m.body.data
}

func (m *Msg) BodyLen() int {
	return m.bodyLength
}

func (m *Msg) SetBody(data []byte) {
	m.body.data = data
	m.bodyLength = len(m.body.data)
}

func (m *Msg) AddContent(data []byte) {
	m.body.data = append(m.body.data, data...)
	m.bodyLength = len(m.body.data)
}

func (msg *Msg) encodeHeaders() {
	for i, vals := range(msg.headers) {
		if len(vals) > 0 {
			arr := []string{}

			for j, val := range(vals) {
				fmt.Println("Us encode range for", j, msg.Id)
				h := mime.BEncoding.Encode("UTF-8", val)
				arr = append(arr, h)
			}
			msg.headers[i] = arr
		}
	}
}

func (msg *Msg) decodeHeaders() {
	decoder := new(mime.WordDecoder)

	for i, vals := range(msg.headers) {
		if len(vals) > 0 {
			arr := []string{}

			for j, val := range(vals) {
				fmt.Println("Ua decode range for", j, msg.Id)
				h, err := decoder.DecodeHeader(val)

				if err != nil {
					panic(err)
				}
				arr = append(arr, h)
			}
			msg.headers[i] = arr
		}
	}
}

func NewMsg(headers map[string][]string, body interface{}) (Msg, error) {
	fmt.Println("New message", headers)
	msg := Msg{headers: headers, body: &Body{}}
	msg.decodeHeaders()
	hop := msg.headers["Hop"]

	if len(hop) > 0 {
		msg.Hop = hop
	}

	id := msg.headers["Reqid"]

	if len(id) > 0 {
		msg.Id = id[0]
	}

	dateStr := msg.headers["Date"]

	if len(dateStr) > 0 {
		t, err := mail.ParseDate(dateStr[0])

		if err == nil {
			msg.Date = t
		}
	}

	if len(msg.headers["Message-Id"]) > 0 {
		msg.MessageId = msg.headers["Message-Id"][0]
	}

	contentLength := msg.headers["Content-Length"]

	if len(contentLength) > 0 {
	        contentLen, err := strconv.Atoi(contentLength[0])

		if(err != nil) {
			return msg, err
		} else {
			msg.ContentLength = contentLen
			fmt.Println("New msg content length", contentLen, msg.ContentLength)
		}
	}

	contentType := msg.headers["Content-Type"]

	if len(contentType) > 0 {
		msg.ContentType = contentType[0]
	}

	method := msg.headers["Method"]

	fmt.Println("Method", method)

	if len(method) > 0 {
		msg.Method = method[0]
	}

	codeStr := msg.headers["Code"]

	if len(codeStr) > 0 {
		code, err := strconv.Atoi(codeStr[0]);

		if err != nil {
			fmt.Println("Error during conversion", code)
		}

		msg.Code = code
	}
	return msg, nil
}
