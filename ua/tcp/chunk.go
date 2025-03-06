package tcpua

type SipwareTcpChunk struct {
	written int
	offset int
	buff []byte
}

func (c *SipwareTcpChunk) Get() []byte {
	return c.buff
}

func (c *SipwareTcpChunk) SetN(written int) {
	c.written = written
}

func (c *SipwareTcpChunk) SetOffset(offset int) {
	c.offset = offset
}

func (c *SipwareTcpChunk) Read(n int, offset int) ([]byte, error) {
	return c.buff[c.offset:n], nil
}

func (c *SipwareTcpChunk) Write(data []byte, n int, offset int) error {
	j := 0

	for i := offset; i < n; i++ {
		c.buff[j] = data[i]
		j++
	}

	c.written = j
	c.offset = offset
	return nil
}
