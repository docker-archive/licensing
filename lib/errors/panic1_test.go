package errors

import (
	"fmt"
)

func panicTop1() (err error) {
	defer func() {
		r := recover()
		rerr, ok := r.(error)
		if !ok {
			rerr = fmt.Errorf("panic: %v", r)
		}
		err = WithStack(rerr)
	}()
	return panicMiddle1()
}

func panicMiddle1() error {
	return panicBottom1()
}

func panicBottom1() error {
	panic("panicBottom1")
}
