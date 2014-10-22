package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"

	"github.com/kidoman/embd"
	_ "github.com/kidoman/embd/host/rpi"
)

var in string

func init() {
	flag.StringVar(&in, "in", "8", "GPIO to read the switch state from")

	flag.Parse()
}

func main() {
	fmt.Printf("Blinker\n")

	quit := quitSignal()
	defer signal.Stop(quit)

	err := embd.InitGPIO()
	if err != nil {
		log.Fatal(err)
	}
	defer embd.CloseGPIO()

	swtch, err := NewSwitch(in)
	if err != nil {
		log.Fatal(err)
	}
	defer swtch.Close()

	Loop(swtch.in, quit)
}

// Read from in until the value read is 0, an error occurs or something is sent
// on quit.
func Loop(in embd.DigitalPin, quit <-chan os.Signal) {

Out:
	for {
		val, err := in.Read()

		if err != nil {
			log.Fatal(err)

		} else if val == 0 {
			break
		}

		fmt.Printf("Read %d from pin %d\n", val, in.N())

		select {
		case <-quit:
			break Out
		default:
		}
	}
}

// Create a channel that receives notifications on os.Interrupt and os.Kill
func quitSignal() chan os.Signal {
	out := make(chan struct{})
	signal.Notify(out, os.Interrupt, os.Kill)
	return out
}

// A Switch can be used to read when
type Switch struct {
	in embd.DigitalPin
}

// Create a new Switch. In is the key of the switch to read from and out the
// one to write to.
func NewSwitch(in string) (s *Switch, err error) {
	s = new(Switch)

	s.in, err = embd.NewDigitalPin(in)
	if err != nil {
		goto Error
	}

	err = s.in.SetDirection(embd.In)
	if err != nil {
		goto Error
	}

	return s, nil
Error:
	s.Close()
	return nil, err
}

// Free associated resources
func (s Switch) Close() error {

	var err error

	if s.in != nil {
		s.in.Close()
	}

	return err
}
