package message

var MessageLog []string

func AddMessage(x string) {
	MessageLog = append(MessageLog, x)
}
