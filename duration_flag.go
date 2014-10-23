package main

import "time"

// A Duration is a time.Duration with methods implementing flag.Value
type Duration time.Duration

func (d Duration) String() string {
	return time.Duration(d).String()
}

func (d *Duration) Set(src string) error {

	v, err := time.ParseDuration(src)
	if err != nil {
		*d = Duration(v)
	}
	return err
}

// Sleep for the duration called on
func (d Duration) Sleep() {
	time.Sleep(time.Duration(d))
}
