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
	"bytes"
	"runtime"
	"time"
)

// nonwaiting represents the set of goroutine statuses that do not indicate waiting.
var nonwaiting = map[string]struct{}{
	"idle":      {},
	"runnable":  {},
	"running":   {},
	"syscall":   {},
	"dead":      {},
	"enqueue":   {},
	"copystack": {},
}

type waitpoint struct {
	watch map[string]struct{}
	buf   []byte
}

func newWaitpoint() *waitpoint {
	w := &waitpoint{
		buf: make([]byte, 1024),
	}
	w.Reset()
	return w
}

// grStatus returns the current goroutines and their status.
func (wp *waitpoint) grStatus() map[string]string {
	grs := map[string]string{}
	var n, p1, p2, ll int

	runtime.Gosched()
	for {
		n = runtime.Stack(wp.buf, true)
		if n < len(wp.buf) {
			for _, line := range bytes.Split(wp.buf[:n], []byte{'\n'}) {
				// only interested in lines like "goroutine 33 [running]:"
				if ll = len(line); ll < len("goroutine 1 []:") || !bytes.HasPrefix(line, []byte("goroutine ")) {
					continue
				}
				for p1 = len("goroutine "); p1+2 < ll && line[p1] >= '0' && line[p1] <= '9'; p1++ {
				}
				if line[p1] != ' ' || line[p1+1] != '[' {
					// unexpected format; ignore
					continue
				}
				for p2 = p1 + 2; p2 < ll && line[p2] >= 'a' && line[p2] <= 'z'; p2++ {
				}
				grs[string(line[len("goroutine "):p1])] = string(line[p1+2 : p2])
			}
			return grs
		}
		wp.buf = make([]byte, 2*len(wp.buf))
	}
}

func (wp *waitpoint) Reset() {
	grs := wp.grStatus()
	// fmt.Println(grs, "\n-----\n")
	wp.watch = make(map[string]struct{}, len(grs))
	for g, s := range grs {
		if _, found := nonwaiting[s]; !found {
			wp.watch[g] = struct{}{}
		}
	}
}

func (wp *waitpoint) Wait() {
	for func /*busy*/ () bool {
		grs := wp.grStatus()
		for g := range wp.watch {
			if s, found := grs[g]; found {
				if _, found := nonwaiting[s]; found {
					return true
				}
			}
		}
		// fmt.Println(grs, "\n-----\n")
		return false
	}() {
		time.Sleep(time.Millisecond)
	}
}
