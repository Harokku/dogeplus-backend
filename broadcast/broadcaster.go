package broadcast

type Broadcaster interface {
	Connect() error
	Disconnect() error
	Send(message []byte) error
}
