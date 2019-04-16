package waithost

import (
	"errors"
	"net"
	"reflect"
	"strings"
	"testing"
	"time"
)

func Test_extractTarget(t *testing.T) {
	type args struct {
		target string
	}
	targets := []struct {
		name    string
		args    args
		want    *WaitHost
		wantErr bool
	}{
		{"http-full", args{"http://goland.org:80"}, &WaitHost{Scheme: "http", Host: "goland.org", Port: 80, Timeout: 0, ConnectTimeout: time.Second, logger: defaultLogger, retryMessage: DefaultRetryMessage}, false},
		{"http-no-port", args{"http://goland.org"}, &WaitHost{Scheme: "http", Host: "goland.org", Port: 80, Timeout: 0, ConnectTimeout: time.Second, logger: defaultLogger, retryMessage: DefaultRetryMessage}, false},
		{"https-full", args{"https://goland.org:443"}, &WaitHost{Scheme: "https", Host: "goland.org", Port: 443, Timeout: 0, ConnectTimeout: time.Second, logger: defaultLogger, retryMessage: DefaultRetryMessage}, false},
		{"https-no-port", args{"https://goland.org"}, &WaitHost{Scheme: "https", Host: "goland.org", Port: 443, Timeout: 0, ConnectTimeout: time.Second, logger: defaultLogger, retryMessage: DefaultRetryMessage}, false},
		{"tcp-full", args{"tcp://127.0.0.1:1123"}, &WaitHost{Scheme: "tcp", Host: "127.0.0.1", Port: 1123, Timeout: 0, ConnectTimeout: time.Second, logger: defaultLogger, retryMessage: DefaultRetryMessage}, false},
		{"tcp-no-scheme", args{"127.0.0.1:1123"}, &WaitHost{Scheme: "tcp", Host: "127.0.0.1", Port: 1123, Timeout: 0, ConnectTimeout: time.Second, logger: defaultLogger, retryMessage: DefaultRetryMessage}, false},
		{"tcp-no-addr", args{":1123"}, &WaitHost{Scheme: "tcp", Host: "localhost", Port: 1123, Timeout: 0, ConnectTimeout: time.Second, logger: defaultLogger, retryMessage: DefaultRetryMessage}, false},
		{"error", args{"goland.org"}, &WaitHost{Scheme: "tcp", Host: "goland.org", Port: 0, Timeout: 0, ConnectTimeout: time.Second, logger: defaultLogger, retryMessage: DefaultRetryMessage}, false},
	}
	for _, tt := range targets {
		t.Run(tt.name, func(t *testing.T) {
			got, err := extractTarget(tt.args.target)
			if (err != nil) != tt.wantErr {
				t.Errorf("extractTarget() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("extractTarget() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestHttpWait(t *testing.T) {
	type args struct {
		target string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{"http-full", args{"http://goland.org:80"}, false},
		{"http-no-port", args{"http://goland.org"}, false},
		{"https-full", args{"https://goland.org:443"}, false},
		{"https-no-port", args{"https://goland.org"}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := Wait(tt.args.target); (err != nil) != tt.wantErr {
				t.Errorf("Wait() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestTcpWait(t *testing.T) {

	srv, _ := NewServer("tcp", ":1123")
	go func() {
		_ = srv.Run()
	}()
	defer func() {
		_ = srv.Close()
	}()

	type args struct {
		target string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{"tcp-full", args{"tcp://localhost:1123"}, false},
		{"tcp-no-scheme", args{"localhost:1123"}, false},
		{"tcp-no-addr", args{":1123"}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := Wait(tt.args.target); (err != nil) != tt.wantErr {
				t.Errorf("Wait() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestTcpDelayWait(t *testing.T) {

	DefaultTimeout = time.Duration(10) * time.Second

	srv, _ := NewServer("tcp", ":1123")
	go func() {
		time.Sleep(time.Duration(2) * time.Second)
		_ = srv.Run()
	}()
	defer func() { _ = srv.Close() }()

	type args struct {
		target string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{"tcp-full", args{"tcp://localhost:1123"}, false},
		{"tcp-no-scheme", args{"localhost:1123"}, false},
		{"tcp-no-addr", args{":1123"}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := Wait(tt.args.target); (err != nil) != tt.wantErr {
				t.Errorf("Wait() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestTcpWaitError(t *testing.T) {

	DefaultTimeout = time.Duration(2) * time.Second

	type args struct {
		target string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{"tcp-no-addr", args{":1123"}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := Wait(tt.args.target); (err != nil) != tt.wantErr {
				t.Errorf("Wait() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

type Server interface {
	Run() error
	Close() error
}
type TCPServer struct {
	addr   string
	server net.Listener
}

func NewServer(protocol, addr string) (Server, error) {
	switch strings.ToLower(protocol) {
	case "tcp":
		return &TCPServer{
			addr: addr,
		}, nil
	}
	return nil, errors.New("invalid protocol given")
}
func (t *TCPServer) Run() (err error) {
	t.server, err = net.Listen("tcp", t.addr)
	if err != nil {
		return
	}
	for {
		conn, err := t.server.Accept()
		if err != nil {
			err = errors.New("could not accept connection")
			break
		}
		if conn == nil {
			err = errors.New("could not create connection")
			break
		}
		_ = conn.Close()
	}
	return
}
func (t *TCPServer) Close() (err error) {
	return t.server.Close()
}
