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

package manualclock

import (
	"sync"
	"time"
)

type event struct {
	c       chan time.Time
	clock   *manualClock  // the clock that controls this event
	next    time.Time     // next event time
	d       time.Duration // (tickers only) time between ticks
	fn      func()        // (timers only) AfterFunc function
	stopped bool          // (timers only) stopped or expired
	removed bool          // (timers only) removed from event list
	mu      sync.RWMutex  // protection for `stopped` and `removed` flags, and `next` field
}

func (e *event) Next() time.Time {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.next
}

func (e *event) Stopped() bool {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.stopped
}

func (e *event) play(now time.Time) {
	if e.Stopped() {
		return
	}

	var f func()

	e.mu.Lock()
	if e.fn != nil {
		f = e.fn
		// time.Sleep(time.Millisecond)
	} else {
		f = func() {
			select {
			case e.c <- now:
				// time.Sleep(time.Millisecond)
			default:
			}
		}
	}
	if e.d != 0 {
		e.next = now.Add(e.d)
	} else {
		e.stopped = true
	}
	e.mu.Unlock()

	f()
}

// Timer represents an event timer similar to time.Timer, except it is controlled by a manual clock.
type Timer event

// C returns the timer-expiry channel
func (t *Timer) C() <-chan time.Time {
	return t.c
}

// Stop turns off the timer.
func (t *Timer) Stop() bool {
	t.mu.Lock()
	defer t.mu.Unlock()
	active := !t.stopped
	t.stopped = true
	return active
}

// Reset changes the expiry time of the timer, and reactivates it if it was stopped.
func (t *Timer) Reset(d time.Duration) bool {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.next = t.clock.Now().Add(d)
	active := !t.stopped
	t.stopped = false
	if t.removed {
		t.clock.addEvent((*event)(t))
		t.removed = false
	}
	return active
}

// Ticker represents a "ticker" similar to time.Ticker, except it is controlled by a manual clock.
type Ticker event

// C returns the tick channel
func (t *Ticker) C() <-chan time.Time {
	return t.c
}

// Stop turns off the ticker.
func (t *Ticker) Stop() {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.stopped = true
}

// events represents a list of sortable events.
type events []*event

func (a events) Len() int           { return len(a) }
func (a events) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a events) Less(i, j int) bool { return a[i].Next().After(a[j].Next()) }
