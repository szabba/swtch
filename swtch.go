package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"time"

	"github.com/kidoman/embd"
	_ "github.com/kidoman/embd/host/rpi"
)

var (
	inKey, outKey string
	sleepTime     Duration = Duration(time.Millisecond)
)

func init() {
	flag.StringVar(&inKey, "in", "7", "GPIO to read the switch state from")
	flag.StringVar(&outKey, "out", "8", "GPIO to supply current to the LED on")

	flag.Var(&sleepTime, "sleep", "duration to wait between input reads")

	flag.Parse()
}

func main() {
	fmt.Printf("Blinker\n")

	quit := QuitSignal(1)

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

	Loop(in, out, quit)
}

// InPin only contains the Read method of embd.DigitalPin
type InPin interface {
	Read() (int, error)
}

// OutPin only contains the Write method of embd.DigitalPin
type OutPin interface {
	Write(int) error
}

// Writes from in to out until something is received on quit.
//
// When the input value changes. A one is written only after it
// repeats 1+noiseThreshold times with roughly millisecond intervals between
// the reads.  This allows us to ignore temporary current fluctuations of the
// environment.
func Loop(in InPin, out OutPin, quit <-chan struct{}) {

	var lastSent int

	out.Write(0)
	defer out.Write(0)

Out:
	for {
		val, err := in.Read()
		if err != nil {
			log.Fatal(err)
		}

		zeroNotNoise := val == 0 && lastSent != 0
		nonZeroNotNoise := val != 0 && lastSent == 0

		if zeroNotNoise || nonZeroNotNoise {
			err = out.Write(val)
			if err != nil {
				log.Fatal(err)
			}

			lastSent = val
		}

		select {
		case <-quit:
			break Out
		default:
		}

		sleepTime.Sleep()
	}
}


// Sends n events on channel once os.Kill or os.Interrupt is received
func QuitSignal(n int) chan struct{} {
	var (
		sig  = make(chan os.Signal, 1)
		quit = make(chan struct{})
	)
	signal.Notify(sig, os.Interrupt, os.Kill)

	go func() {
		defer signal.Stop(sig)

		<-sig
		for i := 0; i < n; i++ {
			quit <- struct{}{}
		}
	}()

	return quit
}
