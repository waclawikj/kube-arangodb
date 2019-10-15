//
// DISCLAIMER
//
// Copyright 2018 ArangoDB GmbH, Cologne, Germany
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// Copyright holder is ArangoDB GmbH, Cologne, Germany
//
// Author Adam Janikowski
//

package helper

import (
	"fmt"
	"time"
)

type Interrupt struct {
}

func (i Interrupt) Error() string {
	return "interrupted"
}

func isInterrupt(err error) bool {
	_, ok := err.(Interrupt)
	return ok
}

func Timeout(interval, timeout time.Duration, action func() error) error {
	intervalT := time.NewTicker(interval)
	defer intervalT.Stop()

	timeoutT := time.NewTimer(timeout)
	defer timeoutT.Stop()

	for {
		select {
		case <-intervalT.C:
			err := action()
			if err != nil {
				if isInterrupt(err) {
					return nil
				}
				return err
			}
			break
		case <-timeoutT.C:
			return fmt.Errorf("function timeouted")
		}
	}
}
