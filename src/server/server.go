package server

import (
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
	"time"
)

var err error

const (
	CMD_RESTART = iota
	CMD_GRACE
	CMD_STOP
)

func NewServer() *server {
	result := new(server)
	result.server = &http.Server{}
	return result
}

type server struct {
	server   *http.Server
	listener net.Listener
	signals  chan os.Signal
	commands chan int
	stopchan chan int
	stopped  bool
	stopcmd  int
}

func (self *server) Server() *http.Server {
	return self.server
}

func (self *server) Start(gracefulChild bool) {
	self.signals = make(chan os.Signal, 1)
	self.commands = make(chan int, 10)
	self.stopchan = make(chan int, 1)

	defer close(self.signals)
	defer close(self.commands)
	defer close(self.stopchan)

	signal.Notify(self.signals,
		syscall.SIGUSR1,
		syscall.SIGUSR2,
		syscall.SIGQUIT,
	)

	go func() {
		for {
			switch <-self.signals {
			case syscall.SIGUSR1:
				self.commands <- CMD_GRACE
			case syscall.SIGUSR2:
				self.commands <- CMD_RESTART
			case syscall.SIGQUIT, syscall.SIGTERM:
				self.commands <- CMD_STOP
			}
		}
	}()

	go func() {
		for {
			switch <-self.commands {
			case CMD_GRACE:
				self.Grace()
			case CMD_RESTART:
				self.Restart()
			case CMD_STOP:
				self.Stop()
			}
		}
	}()

	if gracefulChild {
		f := os.NewFile(3, "")
		self.listener, err = net.FileListener(f)
	} else {
		self.listener, err = net.Listen("tcp", self.server.Addr)
	}

	FatalIfError(err)
	self.server.Serve(self.listener)

	switch <-self.stopchan {
	case CMD_GRACE:
		time.Sleep(time.Second)
	}
}

func (self *server) Stop() {
	if !self.stopped {
		self.stopped = true
		log.Println("Stop")
		go func() {
			self.listener.Close()
			self.server = nil
			self.stopchan <- CMD_STOP
		}()
	}
}

func (self *server) Restart() {
	if !self.stopped {
		self.stopped = true
		log.Println("Restart")
		go func() {
			self.listener.Close()
			args := make([]string, 0, len(os.Args)-1)
			for _, arg := range os.Args[1:] {
				if arg != "-grace" {
					args = append(args, arg)
				}
			}

			cmd := exec.Command(os.Args[0], args...)
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			cmd.Start()
			self.stopchan <- CMD_RESTART
		}()
	}
}

func (self *server) Grace() {
	if !self.stopped {
		self.stopped = true
		log.Println("Graceful restart")
		go func() {
			self.listener.Close()
			file, err := self.listener.(*net.TCPListener).File()
			FatalIfError(err)

			args := make([]string, 0, len(os.Args))
			for _, arg := range os.Args[1:] {
				if arg != "-grace" {
					args = append(args, arg)
				}
			}

			args = append(args, "-grace")
			cmd := exec.Command(os.Args[0], args...)
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			cmd.ExtraFiles = []*os.File{
				file,
			}
			cmd.Start()
			self.stopchan <- CMD_GRACE
		}()
	}
}
