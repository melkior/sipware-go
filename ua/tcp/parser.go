package tcpua

import (
	"errors"
)

var (
	createPoolErr = errors.New("Create pool error")
	msgParseErr = errors.New("Message parse error")
	msgReadErr = errors.New("Message read error")
	parserStateErr = errors.New("Parser wrong state")
)

type ParserState int

const (
	PARSER_HEADER ParserState = iota
	PARSER_CONTENT
)
