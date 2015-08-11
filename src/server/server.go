package server

import (
	"errors"
	"github.com/drdreyworld/siglistener"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
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
	stopchan chan int
	stopped  bool
	stopcmd  int
	siglistener.SigListener
}

func (self *server) Server() *http.Server {
	return self.server
}

func (self *server) Start(gracefulChild bool) (err error) {
	self.stopchan = make(chan int, 1)
	defer close(self.stopchan)

	self.HandleSignal(syscall.SIGUSR1, CMD_GRACE, func(cmd siglistener.SigCommand) { LogIfError(self.Grace()) })
	self.HandleSignal(syscall.SIGUSR2, CMD_RESTART, func(cmd siglistener.SigCommand) { LogIfError(self.Restart()) })
	self.HandleSignal(syscall.SIGQUIT, CMD_STOP, func(cmd siglistener.SigCommand) { LogIfError(self.Stop()) })
	self.ListenCommands()

	if gracefulChild {
		f := os.NewFile(3, "")
		self.listener, err = net.FileListener(f)
	} else {
		self.listener, err = net.Listen("tcp", self.server.Addr)
	}

	if err == nil {
		self.server.Serve(self.listener)

		switch <-self.stopchan {
		case CMD_GRACE:
			time.Sleep(time.Second)
		}
	}

	return err
}

func (self *server) Stop() (err error) {
	if self.stopped {
		err = errors.New("Server already stopped")
	} else {
		self.stopped = true
		log.Println("Stop")

		if err = self.listener.Close(); err == nil {
			self.server = nil
			self.stopchan <- CMD_STOP
		}
	}
	return err
}

func (self *server) Restart() (err error) {
	if self.stopped {
		err = errors.New("Server already stopped")
	} else {
		self.stopped = true
		log.Println("Restart")

		if err = self.listener.Close(); err == nil {
			args := make([]string, 0, len(os.Args)-1)
			for _, arg := range os.Args[1:] {
				if arg != "-grace" {
					args = append(args, arg)
				}
			}
			cmd := exec.Command(os.Args[0], args...)
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr

			if err = cmd.Start(); err == nil {
				self.stopchan <- CMD_RESTART
			}
		}
	}
	return err
}

func (self *server) Grace() (err error) {
	if self.stopped {
		err = errors.New("Server already stopped")
	} else {
		self.stopped = true
		log.Println("Graceful restart")

		if err = self.listener.Close(); err == nil {
			if file, err := self.listener.(*net.TCPListener).File(); err == nil {
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
				cmd.ExtraFiles = []*os.File{file}

				if err = cmd.Start(); err == nil {
					self.stopchan <- CMD_GRACE
				}
			}
		}
	}
	return err
}
