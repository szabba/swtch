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

var inKey, outKey string

func init() {
	flag.StringVar(&inKey, "in", "7", "GPIO to read the switch state from")
	flag.StringVar(&outKey, "out", "8", "GPIO to supply current to the LED on")

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

	in, err := embd.NewDigitalPin(inKey)
	if err != nil {
		log.Fatal(err)
	}
	defer in.Close()

	out, err := embd.NewDigitalPin(outKey)
	if err != nil {
		log.Fatal(err)
	}
	defer out.Close()
	out.SetDirection(embd.Out)
	out.Write(0)

	Loop(in, quit)
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
	out := make(chan os.Signal, 1)
	signal.Notify(out, os.Interrupt, os.Kill)
	return out
}
