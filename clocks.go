// Copyright 2016 ALRUX Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package clocks enables time travel (sort of)
//
// The Clock interface groups all the time-passage-dependent features from the
// standard time package.
//
// A "live" Clock is provided in this package giving pass-through access to the
// standard functionality.
//
// A "manual" Clock is included as a separate package, because it is mostly useful
// for testing and it is rarely if ever needed in the actual program.
package clocks

import (
	"time"
)

// Clock is the interface implemented by all clocks based on this package.
type Clock interface {
	Add(d time.Duration)
	Set(t time.Time)

	Now() time.Time

	Sleep(d time.Duration)

	After(d time.Duration) <-chan time.Time
	AfterFunc(d time.Duration, f func()) Timer
	NewTimer(d time.Duration) Timer

	Tick(d time.Duration) <-chan time.Time
	NewTicker(d time.Duration) Ticker
}

// Timer is a common interface for event timers. It is conceptually identical to
// time.Timer, except the channel is accessible via a method, rather than directly.
type Timer interface {
	C() <-chan time.Time
	Stop() bool
	Reset(time.Duration) bool
}

// Ticker is a common interface for "tickers". It is conceptually identical to
// time.Ticker, except the channel is accessible via a method, rather than directly.
type Ticker interface {
	C() <-chan time.Time
	Stop()
}

// New returns a live Clock instance.
func New() Clock {
	return &liveClock{}
}

// liveClock simply wraps the time package functions.
type liveClock struct{}

// Add is no-op on a live clock.
func (*liveClock) Add(d time.Duration) {}

// Set is no-op on a live clock.
func (*liveClock) Set(t time.Time) {}

// Now is a pass-through wrapper around time.Now
func (*liveClock) Now() time.Time { return time.Now() }

// Sleep is a pass-through wrapper around time.Sleep
func (*liveClock) Sleep(d time.Duration) { time.Sleep(d) }

// After is a pass-through wrapper around time.After
func (*liveClock) After(d time.Duration) <-chan time.Time { return time.After(d) }

// AfterFunc wraps the result of time.AfterFunc in a type that satisfies the Timer interface.
func (*liveClock) AfterFunc(d time.Duration, f func()) Timer {
	return liveTimer{time.AfterFunc(d, f)}
}

// NewTimer creates a new time.Timer with the provided duration, an wraps it in a
// type that satisfies the Timer interface.
func (*liveClock) NewTimer(d time.Duration) Timer {
	return liveTimer{time.NewTimer(d)}
}

// Tick is a pass-through wrapper around time.Tick
func (*liveClock) Tick(d time.Duration) <-chan time.Time { return time.Tick(d) }

// NewTicker creates a new time.Ticker with the provided duration, an wraps it in a
// type that satisfies the Ticker interface.
func (*liveClock) NewTicker(d time.Duration) Ticker {
	return liveTicker{time.NewTicker(d)}
}

type liveTimer struct {
	*time.Timer
}

func (t liveTimer) C() <-chan time.Time {
	return t.Timer.C
}

type liveTicker struct {
	*time.Ticker
}

func (t liveTicker) C() <-chan time.Time {
	return t.Ticker.C
}
