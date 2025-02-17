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

package backup

import (
	"fmt"
	"github.com/arangodb/go-driver"
	"github.com/arangodb/kube-arangodb/pkg/backup/utils"
	"strings"
)

func newTemporaryError(err error) error {
	return temporaryError{
		Causer:err,
	}
}

func newTemporaryErrorf(format string, a ... interface{}) error {
	return newTemporaryError(fmt.Errorf(format, a...))
}

type temporaryError struct {
	Causer error
}

func (t temporaryError) Cause() error {
	return t.Causer
}

func (t temporaryError) Error() string {
	return t.Causer.Error()
}

func newFatalError(err error) error {
	return fatalError{
		Causer:err,
	}
}

func newFatalErrorf(format string, a ... interface{}) error {
	return newFatalError(fmt.Errorf(format, a...))
}

type fatalError struct {
	Causer error
}

func (f fatalError) Cause() error {
	return f.Causer
}

func (f fatalError) Error() string {
	return f.Causer.Error()
}

func isTemporaryError(err error) bool {
	if _, ok := err.(temporaryError); ok {
		return true
	}

	if _, ok := err.(fatalError); ok {
		return false
	}

	if v, ok := err.(utils.Temporary); ok {
		if v.Temporary() {
			return true
		}
	}

	if v, ok := err.(driver.ArangoError); ok {
		if temporaryErrorNum.Has(v.ErrorNum) || temporaryCodes.Has(v.Code) {
			return true
		}
	}

	if v, ok := err.(utils.Causer); ok {
		return isTemporaryError(v.Cause())
	}

	// Check error string
	if strings.Contains(err.Error(), "context deadline exceeded") {
		return true
	}

	if strings.Contains(err.Error(), "connection refused") {
		return true
	}

	return false
}

func switchError(err error) error {
	if isTemporaryError(err) {
		return newTemporaryError(err)
	}

	return newFatalError(err)
}