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
	"os"
	"strings"
	"testing"
)

type Scope string

const (
	SMOKE Scope = "SMOKE"
	INTEGRATION Scope = "INTEGRATION"
	ACCEPTANCE Scope = "ACCEPTANCE"

	SCOPE_ENV = "TEST_SCOPE"
)

func (s Scope) Level() int {
	switch s {
	case SMOKE:
		return 1
	case INTEGRATION:
		return 2
	case ACCEPTANCE:
		return 3
	default:
		return 0
	}
}

func (s Scope) IsInScope(target Scope) bool {
	return s.Level() <= s.Level()
}

func GetScope(scope string) Scope {
	return Scope(strings.ToUpper(scope))
}

func GetScopeParam() Scope {
	return GetScope(os.Getenv(SCOPE_ENV))
}

func MarkTest(t *testing.T, scope Scope, allowParallels bool) {
	if !GetScopeParam().IsInScope(scope) {
		t.Skipf("Test is not in scope %s", GetScopeParam())
		return
	}

	if allowParallels {
		t.Parallel()
	}
}

func MarkSmokeTest(t *testing.T, allowParallels bool) {
	MarkTest(t, SMOKE, allowParallels)
}

func MarkIntegrationTest(t *testing.T, allowParallels bool) {
	MarkTest(t, INTEGRATION, allowParallels)
}

func MarkAcceptanceTest(t *testing.T, allowParallels bool) {
	MarkTest(t, ACCEPTANCE, allowParallels)
}