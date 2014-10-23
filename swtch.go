package main

import (
	"flag"
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

	conn := make(chan int)
	go InLoop(in, conn, quit)
	OutLoop(out, conn)
}

// InPin only contains the Read method of embd.DigitalPin
type InPin interface {
	Read() (int, error)
}

// OutPin only contains the Write method of embd.DigitalPin
type OutPin interface {
	Write(int) error
}

func InLoop(in InPin, out chan<- int, quit <-chan struct{}) {

	var lastSent int

	out <- 0
	defer close(out)

Out:
	for {
		val, err := in.Read()
		if err != nil {
			log.Fatal(err)
		}

		zeroNotNoise := val == 0 && lastSent != 0
		nonZeroNotNoise := val != 0 && lastSent == 0

		if zeroNotNoise || nonZeroNotNoise {
			out <- val
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

func OutLoop(out OutPin, in <-chan int) {
	var blink, on bool
	defer out.Write(embd.Low)

Out:
	for {
		select {
		case v, open := <-in:

			if !open {
				break Out
			}
			blink = v != 0
			if blink {
				on = true
			} else {
				on = false
				out.Write(embd.Low)
			}

		case <-time.After(100 * time.Millisecond):

			if blink {
				if on {
					out.Write(embd.High)
				} else {
					out.Write(embd.Low)
				}
				on = !on
			}
		}
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
