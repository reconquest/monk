package main

import (
	"fmt"
	"net"
	"os"
	"path/filepath"
	"time"
)

const F_SETPIPE_SZ = 1031
const F_GETPIPE_SZ = 1032

type Pipe struct {
	net.Conn
	listener net.Listener
	path     string
}

func (pipe *Pipe) Close() error {
	if pipe.listener != nil {
		pipe.listener.Close()
	}

	if pipe.Conn != nil {
		pipe.Conn.Close()
	}

	os.Remove(pipe.path)

	return nil
}

func StartPipe(dir string) (*Pipe, error) {
	path := filepath.Join(
		dir,
		"stream",
		fmt.Sprintf("%d.pipe", time.Now().UnixNano()),
	)

	listener, err := net.Listen("unix", path)
	if err != nil {
		return nil, err
	}

	return &Pipe{
		listener: listener,
		path:     path,
	}, nil
}

func (pipe *Pipe) WaitConnect() error {
	conn, err := pipe.listener.Accept()
	if err != nil {
		return err
	}

	pipe.Conn = conn

	return nil
}

func ConnectPipe(path string) (*Pipe, error) {
	conn, err := net.Dial("unix", path)
	if err != nil {
		return nil, err
	}

	return &Pipe{Conn: conn, path: path}, nil
}
