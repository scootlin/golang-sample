package intranet

type Server interface {
	Done()
	GetPort() int
	GetErrorChannel() chan error
	GetStruct() *server
	Listen()
	SetRecvCallBack(func([]byte, chan error))
	Send(request []byte)
}

type Client interface {
	Send([]byte) error
	SetRecvCallBack(func([]byte, chan error))
	Run()
	GetErrorChannel() chan error
}
