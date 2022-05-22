package adbclient

import (
	"io"
	"net"
	"runtime"

	adb "github.com/zach-klippenstein/goadb"
	"github.com/zach-klippenstein/goadb/wire"
)

type dialer struct {
	reader io.ReadWriteCloser
}

var _ adb.Dialer = &dialer{}

func (d *dialer) Dial(address string) (*wire.Conn, error) {
	netConn, err := net.Dial("tcp", address)
	if err != nil {
		return nil, err
	}

	d.reader = wire.MultiCloseable(netConn)

	runtime.SetFinalizer(d.reader, func(conn io.ReadWriteCloser) {
		conn.Close()
	})

	return &wire.Conn{
		Scanner: wire.NewScanner(d.reader),
		Sender:  wire.NewSender(d.reader),
	}, nil
}
