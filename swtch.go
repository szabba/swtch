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
	inKey, outKey  string
	noiseThreshold int
)

func init() {
	flag.StringVar(&inKey, "in", "7", "GPIO to read the switch state from")
	flag.StringVar(&outKey, "out", "8", "GPIO to supply current to the LED on")

	flag.IntVar(
		&noiseThreshold, "noise", 99,
		"number of times a 1 must be read before it's considered more than a "+
			"fluctuation",
	)

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
// A zero read from in is written immediately. A one is written only after it
// repeats 1+noiseThreshold times with roughly millisecond intervals between
// the reads.  This allows us to ignore temporary current fluctuations of the
// environment.
func Loop(in InPin, out OutPin, quit <-chan os.Signal) {

	var nZs int

	defer out.Write(0)

Out:
	for {
		val, err := in.Read()
		if err != nil {
			log.Fatal(err)

		} else if val != 0 {
			nZs++
		} else {
			nZs = 0
		}

		zero := val == 0
		nonZeroNotNoise := nZs == 1+noiseThreshold && !zero

		if zero || nonZeroNotNoise {
			err = out.Write(val)
			if err != nil {
				log.Fatal(err)
			}
		}

		select {
		case <-quit:
			break Out
		default:
		}

		time.Sleep(time.Millisecond)
	}
}

// Create a channel that receives notifications on os.Interrupt and os.Kill
func quitSignal() chan os.Signal {
	out := make(chan os.Signal, 1)
	signal.Notify(out, os.Interrupt, os.Kill)
	return out
}
