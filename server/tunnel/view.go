package tunnel

import (
	"io"

	"golang.org/x/crypto/ssh/terminal"
)

type View io.Writer

type StatusView struct {
	term *terminal.Terminal
}

func NewStatusView(raw io.ReadWriter) StatusView {
	return StatusView{
		term: terminal.NewTerminal(raw, ""),
	}
}

func (view StatusView) Write(p []byte) (int, error) {
	return view.term.Write(p)
}
