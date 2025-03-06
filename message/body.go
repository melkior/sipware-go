package message

type Body struct {
	data []byte
}

func (b *Body) New() Body {
	return Body{}
}

func (b *Body) Set(data []byte) {
	b.data = data
}
