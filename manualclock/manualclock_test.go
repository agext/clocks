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
	"testing"
	"time"
)

type eventStamp struct {
	l string
	t time.Time
}

func Test_New(t *testing.T) {
	now := time.Now()
	clock := New()
	clock.Set(now)

	time.Sleep(time.Millisecond)

	if clock.Now() != now {
		t.Error(`a manual clock should not advance on its own`)
	}

	clock.Add(time.Hour)
	if exp, act := now.Add(time.Hour), clock.Now(); act != exp {
		t.Errorf(`Add() is incorrect: want %s got %s`, exp.Format(time.ANSIC), act.Format(time.ANSIC))
	}

	clock.Set(now.Add(2 * time.Hour))
	if exp, act := now.Add(2*time.Hour), clock.Now(); act != exp {
		t.Errorf(`Set() is incorrect: want %s got %s`, exp.Format(time.ANSIC), act.Format(time.ANSIC))
	}

	ch := make(chan eventStamp, 15)

	go func() {
		ch <- eventStamp{"start", clock.Now()}
		clock.Sleep(time.Millisecond)
		ch <- eventStamp{"Sleep", clock.Now()}
		rt := <-clock.After(time.Millisecond)
		ch <- eventStamp{"After", rt}
		clock.AfterFunc(time.Second, func() {
			ch <- eventStamp{"AfterFunc", clock.Now()}
		})
		timer := clock.NewTimer(time.Millisecond)
		rt = <-timer.C()
		ch <- eventStamp{"Timer", rt}
		timer.Reset(5 * time.Millisecond)
		rt = <-timer.C()
		ch <- eventStamp{"Timer 2", rt}
		timer = clock.AfterFunc(time.Second, func() {
			ch <- eventStamp{"AfterFunc 2", clock.Now()}
		})
		go func() {
			tc := clock.Tick(5 * time.Millisecond)
			clock.Sleep(2 * time.Millisecond)
			t := clock.NewTicker(5 * time.Millisecond)
			for i := 0; i < 10; i++ {
				select {
				case rt := <-tc:
					ch <- eventStamp{"Tick", rt}
				case rt := <-t.C():
					ch <- eventStamp{"Ticker", rt}
				}
			}
			t.Stop()
		}()
		timer.Stop()
		clock.Sleep(5 * time.Second)
		timer.Reset(time.Second)
	}()

	time.Sleep(5 * time.Millisecond)
	now = clock.Now()
	es := <-ch
	if es.l != "start" {
		t.Errorf(`unexpected event: want %q got %q`, "start", es.l)
	} else if es.t != now {
		t.Errorf(`start time is incorrect by %s`, es.t.Sub(now))
	}

	clock.Add(500 * time.Microsecond)

	select {
	case es = <-ch:
		t.Error(`events started too early`)
	default:
	}

	clock.Add(500 * time.Millisecond)

	select {
	case es = <-ch:
		if es.l != "Sleep" {
			t.Errorf(`unexpected event: want %q got %q`, "Sleep", es.l)
		} else if es.t != now.Add(time.Millisecond) {
			t.Errorf(`Sleep time is incorrect by %s`, es.t.Sub(now.Add(time.Millisecond)))
		}
	default:
		t.Fatal(`simulated sleep did not wake up on time`)
	}

	now = es.t
	es = <-ch
	if es.l != "After" {
		t.Errorf(`unexpected event: want %q got %q`, "After", es.l)
	} else if es.t != now.Add(time.Millisecond) {
		t.Errorf(`After time is incorrect by %s`, es.t.Sub(now.Add(time.Millisecond)))
	}

	now = es.t
	es = <-ch
	if es.l != "Timer" {
		t.Errorf(`unexpected event: want %q got %q`, "Timer", es.l)
	} else if es.t != now.Add(time.Millisecond) {
		t.Errorf(`Timer time is incorrect by %s`, es.t.Sub(now.Add(time.Millisecond)))
	}

	now = es.t
	select {
	case es = <-ch:
		if es.l != "Timer 2" {
			t.Errorf(`unexpected event: want %q got %q`, "Timer 2", es.l)
		} else if es.t != now.Add(5*time.Millisecond) {
			t.Errorf(`Timer 2 time is incorrect by %s`, es.t.Sub(now.Add(5*time.Millisecond)))
		}
	default:
		t.Fatal(`Timer reset did not reactivate timer`)
	}

	now = es.t
	for i := 0; i < 5; i++ {
		select {
		case es = <-ch:
		case <-time.After(time.Millisecond):
			t.Fatalf(`Tick event #%d did not fire`, i)
		}
		if es.l != "Tick" {
			t.Errorf(`unexpected event: want %q got %q`, "Tick", es.l)
		} else if es.t != now.Add(5*time.Millisecond) {
			t.Errorf(`Tick time is incorrect by %s`, es.t.Sub(now.Add(5*time.Millisecond)))
		}
		select {
		case es = <-ch:
		case <-time.After(time.Millisecond):
			t.Fatalf(`Ticker event #%d did not fire`, i)
		}
		if es.l != "Ticker" {
			t.Errorf(`unexpected event: want %q got %q`, "Ticker", es.l)
		} else if es.t != now.Add(7*time.Millisecond) {
			t.Errorf(`Ticker time is incorrect by %s`, es.t.Sub(now.Add(7*time.Millisecond)))
		}
		now = now.Add(5 * time.Millisecond)
	}

	select {
	case es = <-ch:
		t.Error(`AfterFunc event fired too early`)
	default:
	}

	now = clock.Now().Add(-500 * time.Microsecond)
	clock.Add(time.Second)

	es = <-ch
	if es.l != "AfterFunc" {
		t.Errorf(`unexpected event: want %q got %q`, "AfterFunc", es.l)
	} else if es.t != now.Add(502*time.Millisecond) {
		t.Errorf(`AfterFunc time is incorrect by %s`, es.t.Sub(now.Add(502*time.Millisecond)))
	}

	select {
	case es = <-ch:
		t.Error(`AfterFunc 2 event fired too early`)
	default:
	}

	clock.Add(5 * time.Second)
	select {
	case es = <-ch:
		if es.l != "AfterFunc 2" {
			t.Errorf(`unexpected event: want %q got %q`, "AfterFunc 2", es.l)
		} else if es.t != now.Add(5508*time.Millisecond) {
			t.Errorf(`AfterFunc 2 time is incorrect by %s`, es.t.Sub(now.Add(5508*time.Millisecond)))
		}
	default:
		t.Fatal(`AfterFunc timer reset did not reactivate timer`)
	}
}
