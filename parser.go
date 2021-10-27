package krinder

type Parser interface {
	ParseMessage(message string) (Command, error)
}

type Command interface {
	Command() (string, error)
	Contents() (string, error)
	Reply(data []byte) error
}
