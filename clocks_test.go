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

package clocks

import (
	"testing"
	"time"
)

func Test_New(t *testing.T) {
	clock := New()

	clock.Add(time.Hour)
	clockNow := clock.Now()
	timeNow := time.Now()
	if clockNow.After(timeNow) {
		t.Error(`New().Add() is not no-op`)
	}

	clock.Set(clockNow.Add(time.Hour))
	clockNow = clock.Now()
	timeNow = time.Now()
	if clockNow.After(timeNow) {
		t.Error(`New().Set() is not no-op`)
	}

	if timeNow.Sub(clockNow) > time.Second {
		t.Error(`New().Now() is too far from time.Now()`)
	}

	timeout := map[string]<-chan time.Time{
		"Sleep": func() <-chan time.Time {
			ch := make(chan time.Time, 1)
			go func() {
				clock.Sleep(time.Microsecond)
				ch <- clock.Now()
			}()
			return ch
		}(),
		"After": clock.After(time.Microsecond),
		"AfterFunc": func() <-chan time.Time {
			ch := make(chan time.Time, 1)
			clock.AfterFunc(time.Microsecond, func() { ch <- clock.Now() })
			return ch
		}(),
		"NewTimer": func() <-chan time.Time {
			ch := make(chan time.Time, 1)
			go func() {
				nt := clock.NewTimer(10 * time.Microsecond)
				clock.Sleep(5 * time.Microsecond)
				nt.Reset(85 * time.Microsecond)
				<-nt.C()
				ch <- clock.Now()
				nt.Stop()
			}()
			return ch
		}(),
		"Tick": clock.Tick(10 * time.Millisecond),
		"NewTicker": func() <-chan time.Time {
			ch := make(chan time.Time, 1)
			go func() {
				nt := clock.NewTicker(50 * time.Millisecond)
				<-nt.C()
				ch <- clock.Now()
				nt.Stop()
			}()
			return ch
		}(),
	}

	for fn := range timeout {
		select {
		case <-timeout[fn]:
		case <-time.After(100 * time.Millisecond):
			t.Error(`New().` + fn + `() is taking too long - the clock may not be live`)
		}
	}
}
