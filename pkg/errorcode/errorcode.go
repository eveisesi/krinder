package errorcode

import "fmt"

type ErrorCode string

const (
	Internal ErrorCode = "INTERNAL"
)

type instance struct {
	code    ErrorCode
	message string
}

func (i instance) Error() string {
	return fmt.Sprintf("[%s] %s", i.code, i.message)
}

func New(code ErrorCode, message string) instance {
	return instance{
		code:    code,
		message: message,
	}
}
