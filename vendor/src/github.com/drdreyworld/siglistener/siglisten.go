package siglistener

import (
	"log"
	"os"
	"os/signal"
)

type SigCommand int

type SigHandler func(cmd SigCommand)

type SigListener struct {
	cmdChan  chan SigCommand
	sigChan  chan os.Signal
	signals  map[os.Signal]SigCommand
	handlers map[SigCommand]SigHandler
}

func (self *SigListener) HandleSignal(sig os.Signal, cmd SigCommand, handler SigHandler) {
	if self.signals == nil {
		self.signals = make(map[os.Signal]SigCommand)
	}

	if self.handlers == nil {
		self.handlers = make(map[SigCommand]SigHandler)
	}

	self.signals[sig] = cmd
	self.handlers[cmd] = handler
}

func (self *SigListener) ListenCommands() {
	go func() {
		signals := make([]os.Signal, 0, len(self.signals))

		for sig := range self.signals {
			signals = append(signals, sig)
		}

		self.sigChan = make(chan os.Signal, 10)
		self.cmdChan = make(chan SigCommand, 1)

		signal.Notify(self.sigChan, signals...)

		defer close(self.sigChan)
		defer close(self.cmdChan)

		for {
			select {
			case sig := <-self.sigChan:
				if cmd, ok := self.signals[sig]; ok {
					self.cmdChan <- cmd
				} else {
					log.Println("Unknown signal", sig)
				}
			case cmd := <-self.cmdChan:
				if handler, ok := self.handlers[cmd]; ok {
					go handler(cmd)
				} else {
					log.Println("Unknown SigCommand", cmd)
				}
			}
		}
	}()
}
