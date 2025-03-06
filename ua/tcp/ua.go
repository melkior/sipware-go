package tcpua

import (
	"fmt"
	"log"
	"os"
	"net"
	"sync"
	"time"
	"mime"
	"bytes"
	"syscall"
	"context"
	"os/signal"
	"math/rand"
	"sipware/api"
	"sipware/trans"
	"sipware/message"
	"sipware/queues"
)

type User struct {
	DisplayName string
	Email string
	Alias string
}

type RegisterConfig struct {
        From string `json:"From"`
        To string `json:"To"`
        Method string `json:"Method"`
        Alias string `json:"Alias"`
        Password string `json:"Password"`
}

type ContactConfig struct {
        File string `json:"contact"`
        Value string `json:"value"`
        Expires string `json:"expires"`
}

type CacheConfig struct {
        Contact ContactConfig `json:"contact"`
}

type Config struct {
        Open string `json:"open"`
        Register RegisterConfig `json:"register"`
        Cache CacheConfig `json:"cache"`
}

type SipwareTcpUa struct {
	sipware.Ua
	ctx context.Context
	name string
	host string
	seed rand.Rand
	user User
	conn net.Conn
	exit chan os.Signal
	wg sync.WaitGroup
	buf SipwareTcpBuf
	tr map[string] trans.SipwareTr
	inCh chan message.Msg
	outCh chan message.Msg
	Queues map[string]sipware.Queue
}

func New(name string) *SipwareTcpUa {
	seed := *rand.New(rand.NewSource(time.Now().UnixNano()))

	fmt.Println("Seed", seed)

	tr := make(map[string](trans.SipwareTr))
	queues := make(map[string](sipware.Queue))

	wg := sync.WaitGroup{}
	wg.Add(1)

	return &SipwareTcpUa {
		ctx: context.Background(),
		name: name, seed: seed,
		wg: wg, tr: tr, Queues: queues,
	}
}

func (ua *SipwareTcpUa) Done() {
	ua.wg.Done()
}

func (ua *SipwareTcpUa) SetExitHandler() {
	ua.exit = make(chan os.Signal)
	signal.Notify(ua.exit, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
}

func (ua *SipwareTcpUa) Exit() chan os.Signal {
	return ua.exit
}

func (ua *SipwareTcpUa) randomString(length int, charset string) string {
	b := make([]byte, length)

	for i := range b {
		b[i] = charset[ua.seed.Intn(len(charset))]
	}

	return string(b)
}

func (ua *SipwareTcpUa) Name() string {
        fmt.Println("Sipware Tcp Ua Name", ua.name)
        return ua.name
}

func (ua *SipwareTcpUa) Host() string {
        fmt.Println("Sipware Tcp Ua Name", ua.name)
        return ua.host
}

func (ua *SipwareTcpUa) Print() {
        fmt.Println("Sipware Tcp Ua Print");
}

func (ua *SipwareTcpUa) Wait() {
	ua.wg.Wait();
}

func (ua *SipwareTcpUa) startReading() {
	for {
		chunk := ua.buf.NewChunk()
		n, err := ua.conn.Read(chunk.Get())

		if err != nil {
			fmt.Println("Error:", err)
			os.Exit(1);
		}

		chunk.SetN(n)
		ua.buf.Push(chunk);
		msgs, err := ua.buf.Read()
		fmt.Println("Ua start loop err", err)
		fmt.Println("Ua start loop msgs", len(msgs))

		for idx, msg := range msgs {
			fmt.Println("Us start for range", idx, msg.Id)
			if(len(msg.Id) > 0) {
				ua.inCh <- msg
			}
		}
	}
}

func (ua *SipwareTcpUa) dispatchRequest(msg message.Msg) {
	active := ua.Queues["active"]
	incoming := ua.Queues["incoming"]

	incoming.Enqueue(msg)
	incoming.Process(func(m message.Msg) {
		active.Enqueue(m)
	})

}

func (ua *SipwareTcpUa) dispatchResponse(msg message.Msg) {
	id := msg.Id

	if(len(id) > 0) {
		tr := ua.tr[id]

		tr.Write(msg)
	}

}

func (ua *SipwareTcpUa) startDispatcher() {
	for {
		select {
		case msg := <- ua.inCh:
			fmt.Println("Ua dispatch", msg.Id)

			if msg.Code == 0 {
				ua.dispatchRequest(msg)
			} else {
				ua.dispatchResponse(msg)
			}
		}
	}
} 

func (ua *SipwareTcpUa) Start() {
	// Create a buffer to read data into
	fmt.Println("Ua start")
	// ua.buf := SipwareTcpBuf{}
	ua.buf.CreatePool(4096)

	ua.inCh = make(chan message.Msg)
	ua.outCh = make(chan message.Msg)

	ua.Queues["incoming"] = &SipwareQueue.Incoming{}
	ua.Queues["outgoing"] = &SipwareQueue.Outgoing{}
	ua.Queues["active"] = &SipwareQueue.Active{}

	go ua.startReading()
	go ua.startDispatcher()
}

func (ua *SipwareTcpUa) Open(host string) {
	fmt.Println("Sipware Tcp Ua Open", host);
	dst, err := net.ResolveTCPAddr("tcp", host)

	if err != nil {
		fmt.Println("Tcp resolve failed:", err.Error())
		os.Exit(1)
	}

	ua.host = host
	wg := sync.WaitGroup{}
	wg.Add(1)

	go func() {
		defer wg.Done()
		conn, err := net.DialTCP("tcp", nil, dst)

		if err != nil {
			fmt.Println("Dial failed:", err.Error())
			os.Exit(1)
		}
		// conn.Close()
		ua.conn = conn
	} ()
	wg.Wait()
}

func (ua *SipwareTcpUa) Destroy(exiting bool) {
	fmt.Println("Ua destroy")

	ua.conn.Close()

	if(exiting) {
		ua.exit <- syscall.SIGINT
	}
}

func (ua *SipwareTcpUa) Read(data []byte) (n int, err error) {
	return ua.conn.Read(data)
}

func (ua *SipwareTcpUa) write(p []byte) (int, error) {
	fmt.Println("Sipware Tcp Ua Write", string(p))
	n, err := ua.conn.Write(p);

	return n, err
}

func (ua *SipwareTcpUa) Write(m message.Msg) error {
        buf := &bytes.Buffer{}
	headers := m.GetHeaders()
	body := m.Body()

        if len(headers) > 0 {
                for key, val := range headers {
                        for _, v := range val {
                                buf.WriteString(key)
                                buf.WriteString(": ")
                                buf.WriteString(mime.BEncoding.Encode("utf-8", v))
                                buf.WriteString("\r\n")
                        }
                }
        }

        buf.WriteString("\r\n")
        // fmt.Println("WRITE HDR BUF", body, "|", buf, "|")

        if len(body) > 0 {
                buf.Write(body[:len(body)])
        }

        b := buf.Bytes()
        _, err := ua.write(b);
        return err
}

func (ua *SipwareTcpUa) Request(msg message.Msg, f func(message.Msg) error) error {
	fmt.Println("Sipware Tcp Ua Request");

	err := ua.Write(msg)

	if err == nil {
		reqid := msg.Get("Reqid")
		ua.tr[reqid[0]] = trans.New()
		ua.tr[reqid[0]].Add(1)

		go func() {
			i := 1
			ctx, cancel := context.WithTimeout(ua.ctx, time.Second)

			defer cancel()

			for {
				select {
				case <-ctx.Done():
					fmt.Println("REQUEST CTX DONE")
					return;
				case msg := <- ua.tr[reqid[0]].Read():
					fmt.Println("Ua response transaction", i, msg.Id)
					i++
					ua.tr[reqid[0]].Done()
					cancel()
					f(msg)
				}
			}
		} ()
		ua.tr[reqid[0]].Wait()
	}

	return err
}

func (ua *SipwareTcpUa) Reply(m message.Msg) {
	fmt.Println("Sipware Tcp Ua Reply", m);
	err := ua.Write(m)
	fmt.Println("Sipware Tcp Ua Reply Write", err);
}

func (ua *SipwareTcpUa) Register(cf interface{}) {
	cfg := cf.(RegisterConfig)
	fmt.Println("Sipware Tcp Ua Register", cfg);

	reqid := ua.randomString(16, "ascii");

	headers := map[string][]string {
		"From": []string{cfg.From},
		"To": []string{cfg.To},
		"Method": []string{"register"},
		"Password": []string{cfg.Password},
		"Alias": []string{cfg.Alias},
		"Reqid": []string{reqid},
		"Content-Length": []string{"0"},
	}

	msg, err := message.NewMsg(headers, "")

	if err != nil {
		return;
	}

	wg := sync.WaitGroup{}
	wg.Add(1)

	go func() {
		defer wg.Done()
		err := ua.Request(msg, func(resp message.Msg) error {
			fmt.Println("Ua register response", resp)

			displayName := resp.Get("Display_name")
			email := resp.Get("Email")
			alias := resp.Get("Alias")

			if len(displayName) != 0 {
				ua.user.DisplayName = displayName[0]
			}

			if len(email) != 0 {
				ua.user.Email = email[0]
			}

			if len(alias) != 0 {
				ua.user.Alias = alias[0]
			}
			fmt.Println("REGISTER USER", ua.user)
			return nil
		})

		if err != nil {
			log.Fatalf("failed to write mail to STDOUT: %s", err)
		}
	} ()

	wg.Wait()
}
