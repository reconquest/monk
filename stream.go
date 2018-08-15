package main

import (
	"errors"
	"fmt"
	"io"
	"sync"
	"time"
)

const (
	StreamChunkSize = 32 * 1024
)

type StreamBuffer struct {
	conn *StreamConnection
	data []byte
}

type Stream struct {
	started      bool
	startedMutex sync.Mutex

	conns      []*StreamConnection
	connsMutex sync.Mutex

	buffer chan StreamBuffer

	maxBufferSize int

	transferCond sync.Cond
}

type StreamConnection struct {
	machine string
	pipe    io.ReadWriteCloser
	once    sync.Once
	onClose func()
}

func (conn *StreamConnection) Close() {
	conn.once.Do(func() {
		conn.pipe.Close()
		if conn.onClose != nil {
			conn.onClose()
		}
	})
}

func NewStream(bufferSize int) *Stream {
	return &Stream{
		buffer: make(chan StreamBuffer, bufferSize/StreamChunkSize),
		transferCond: sync.Cond{
			L: &sync.Mutex{},
		},
	}
}

func (stream *Stream) write(conn *StreamConnection, data []byte) error {
	select {
	case stream.buffer <- StreamBuffer{conn, data}:
		return nil
	default:
		return errors.New("buffer is full")
	}
}

func (stream *Stream) Close() {
	close(stream.buffer)

	conns := []*StreamConnection{}
	stream.connsMutex.Lock()
	for _, conn := range stream.conns {
		conns = append(conns, conn)
	}
	stream.connsMutex.Unlock()

	for _, conn := range conns {
		conn.Close()
	}
}

func (stream *Stream) transfer() {
	for {
		item, ok := <-stream.buffer
		if !ok {
			return
		}

		stream.transferCond.L.Lock()
		for {
			if stream.broadcast(item.conn, item.data) {
				break
			}

			stream.transferCond.Wait()
		}

		stream.transferCond.L.Unlock()
	}
}

func (stream *Stream) Serve(machine string, pipe io.ReadWriteCloser) *sync.WaitGroup {
	stream.start()

	logger.Infof("stream: connected %s", machine)

	worker := &sync.WaitGroup{}
	worker.Add(1)

	conn := &StreamConnection{
		machine: machine,
		pipe:    pipe,
	}

	conn.onClose = func() {
		defer worker.Done()
		stream.removeConnection(conn)
	}

	stream.connsMutex.Lock()
	stream.conns = append(stream.conns, conn)
	stream.connsMutex.Unlock()

	go stream.read(conn)

	stream.transferCond.Signal()

	return worker
}

func (stream *Stream) start() {
	stream.startedMutex.Lock()
	defer stream.startedMutex.Unlock()

	if stream.started {
		return
	}

	stream.started = true

	go stream.transfer()
}

func (stream *Stream) removeConnection(conn *StreamConnection) {
	stream.connsMutex.Lock()
	for i, item := range stream.conns {
		if item == conn {
			stream.conns = append(
				stream.conns[:i],
				stream.conns[i+1:]...,
			)

			logger.Infof("stream: disconnected %s", conn.machine)

			break
		}
	}
	stream.connsMutex.Unlock()
}

func (stream *Stream) broadcast(owner *StreamConnection, data []byte) bool {
	running := 0

	stream.connsMutex.Lock()
	wrote := make(chan *StreamConnection, len(stream.conns))
	for _, conn := range stream.conns {
		if conn == owner {
			continue
		}

		running++
		go func(conn *StreamConnection) {
			_, err := conn.pipe.Write(data)
			if err != nil {
				errorh(err, "unable to write to stream: %s", conn.machine)
				conn.Close()
			}

			wrote <- conn
		}(conn)
	}
	stream.connsMutex.Unlock()

	if running == 0 {
		return false
	}

	after := time.After(time.Second * 5)
	done := map[*StreamConnection]struct{}{}
waiting:
	for {
		select {
		case conn := <-wrote:
			done[conn] = struct{}{}
			running--
			if running == 0 {
				break waiting
			}

		case <-after:
			break waiting
		}
	}

	var toClose []*StreamConnection
	var result bool

	stream.connsMutex.Lock()
	for _, conn := range stream.conns {
		if conn == owner {
			continue
		}
		_, ok := done[conn]
		if !ok {
			errorh(
				fmt.Errorf("deadline exceeded: %s", time.Second*5),
				"unable to write to stream: %s",
				conn.machine,
			)

			toClose = append(toClose, conn)
		} else {
			result = true
		}
	}
	stream.connsMutex.Unlock()

	for _, conn := range toClose {
		conn.Close()
	}

	return result
}

func (stream *Stream) read(
	conn *StreamConnection,
) {
	defer conn.Close()

	buffer := make([]byte, StreamChunkSize)

	for {
		read, err := conn.pipe.Read(buffer)
		if err != nil {
			if err == io.EOF {
				break
			}

			errorh(err, "can't read stream of %s", conn.machine)
			break
		}

		clone := make([]byte, read)
		copy(clone, buffer[:read])

		err = stream.write(conn, clone)
		if err != nil {
			errorh(err, "can't write message of %s to stream", conn.machine)
		}
	}
}
