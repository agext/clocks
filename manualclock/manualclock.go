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

// Package manualclock provides a clock that only advances when explicitly told to.
package manualclock

import (
	"sort"
	"sync"
	"time"

	"github.com/agext/clocks"
)

// manualClock represents a clock that only advances when explicitly told to.
// A pointer to it satisfies the Clock interface.
type manualClock struct {
	now            time.Time    // current time
	events         events       // dependent events (e.g. tickers & timers)
	newEvents      events       // buffer for adding dependent events
	nowMutex       sync.RWMutex // protection for current time
	eventsMutex    sync.Mutex   // protection for event list
	newEventsMutex sync.Mutex   // protection for event buffer
}

// New returns a manual clock instance set to the current time.
func New() clocks.Clock {
	return &manualClock{now: time.Now()}
}

// addEvent adds the provided event to the list maintained and controlled by this clock.
func (mc *manualClock) addEvent(e *event) {
	mc.newEventsMutex.Lock()
	mc.newEvents = append(mc.newEvents, e)
	mc.newEventsMutex.Unlock()
}

// Add moves the the manual clock by the specified duration, triggering any ticker or
// timer activity that would occur in the time interval between the old and new times.
func (mc *manualClock) Add(d time.Duration) {
	mc.Set(mc.Now().Add(d))
}

// Set moves the manual clock to the specified time, triggering any ticker or
// timer activity that would occur in the time interval between the old and new times.
func (mc *manualClock) Set(now time.Time) {
	mc.eventsMutex.Lock()
	defer func() {
		mc.nowMutex.Lock()
		mc.now = now
		mc.nowMutex.Unlock()
		mc.eventsMutex.Unlock()
	}()

	mc.newEventsMutex.Lock()
	mc.events = append(mc.events, mc.newEvents...)
	mc.newEvents = mc.newEvents[:0]
	mc.newEventsMutex.Unlock()

	if len(mc.events) == 0 {
		return
	}

	sort.Sort(mc.events)
	last, first := 0, len(mc.events)-1
	for last <= first && mc.events[last].Next().After(now) {
		last++
	}

	if last > first {
		return
	}

	for e := mc.events[first]; !e.Next().After(now); e = mc.events[first] {
		next := e.Next()
		mc.nowMutex.Lock()
		mc.now = next
		mc.nowMutex.Unlock()
		wp := newWaitpoint()
		e.play(next)
		wp.Wait()
		if e.Stopped() {
			mc.events = mc.events[:first]
			e.mu.Lock()
			e.removed = true
			e.mu.Unlock()
		}
		mc.newEventsMutex.Lock()
		mc.events = append(mc.events, mc.newEvents...)
		mc.newEvents = mc.newEvents[:0]
		mc.newEventsMutex.Unlock()

		first = len(mc.events) - 1
		if first < 0 {
			break
		}

		sort.Sort(mc.events[last:])

		for last <= first && mc.events[last].Next().After(now) {
			last++
		}

		if last > first {
			break
		}
	}
}

// Now returns the current time on the manual clock.
func (mc *manualClock) Now() time.Time {
	mc.nowMutex.RLock()
	defer mc.nowMutex.RUnlock()
	return mc.now
}

// Sleep pauses the current goroutine for the given duration on the manual clock.
// The clock must be moved forward in another goroutine.
func (mc *manualClock) Sleep(d time.Duration) {
	<-mc.After(d)
}

// After waits for the duration to elapse and then sends the current time on the returned channel.
func (mc *manualClock) After(d time.Duration) <-chan time.Time {
	return mc.NewTimer(d).C()
}

// AfterFunc waits for the duration to elapse and then executes a function.
// A Timer is returned that can be stopped.
func (mc *manualClock) AfterFunc(d time.Duration, f func()) clocks.Timer {
	t := mc.NewTimer(d).(*Timer)
	t.fn = f
	return t
}

// NewTimer returns a new instance of Timer, controlled by the manual clock.
func (mc *manualClock) NewTimer(d time.Duration) clocks.Timer {
	t := &Timer{
		c:     make(chan time.Time, 1),
		clock: mc,
		next:  mc.Now().Add(d),
	}
	mc.addEvent((*event)(t))
	return t
}

// Tick is a convenience function for Ticker().
// It will return a ticker channel that cannot be stopped.
func (mc *manualClock) Tick(d time.Duration) <-chan time.Time {
	return mc.NewTicker(d).C()
}

// NewTicker returns a new instance of Ticker, controlled by the manual clock.
func (mc *manualClock) NewTicker(d time.Duration) clocks.Ticker {
	t := &Ticker{
		c:     make(chan time.Time, 1),
		clock: mc,
		d:     d,
		next:  mc.Now().Add(d),
	}
	mc.addEvent((*event)(t))
	return t
}
