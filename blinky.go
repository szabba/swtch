package main

import (
	"fmt"
	"log"

	"github.com/kidoman/embd"
	_ "github.com/kidoman/embd/host/rpi"
)

func main() {
	fmt.Printf("Blinker\n")

	err := embd.InitGPIO()
	if err != nil {
		log.Fatal(err)
	}
	defer embd.CloseGPIO()

	swtch, err := NewSwitch("8", "7")
	if err != nil {
		log.Fatal(err)
	}
	defer swtch.Close()

	readLoop(swtch.In)
}

// Read from in until the value read is 0 or an error occurs.
func readLoop(in embd.DigitalPin) {
	in.SetDirection(embd.In)

	for {
		val, err := in.Read()

		if err != nil {
			log.Fatal(err)

		} else if val == 0 {
			break
		}

		fmt.Printf("Read %d from pin %d\n", val, in.N())
	}
}

// A Switch can be used to read when
type Switch struct {
	In, Out embd.DigitalPin
}

// Create a new Switch. In is the key of the switch to read from and out the
// one to write to.
func NewSwitch(in, out string) (s *Switch, err error) {
	s = new(Switch)

	s.In, err = embd.NewDigitalPin(in)
	if err != nil {
		goto Error
	}

	err = s.In.SetDirection(embd.In)
	if err != nil {
		goto Error
	}

	s.Out, err = embd.NewDigitalPin(out)
	if err != nil {
		goto Error
	}

	err = s.Out.SetDirection(embd.Out)
	if err != nil {
		goto Error
	}

	err = s.Out.Write(1)
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

	if s.In != nil {
		s.In.Close()
	}

	if s.Out != nil {
		err = s.Out.Write(0)
		s.Out.Close()
	}

	return err
}
