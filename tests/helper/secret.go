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
	"github.com/stretchr/testify/require"
	core "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	typedCore "k8s.io/client-go/kubernetes/typed/core/v1"
	"testing"
	"time"
)

func KubernetesSecretClient(t *testing.T) typedCore.SecretInterface {
	return KubernetesClient(t).CoreV1().Secrets(Namespace(t))
}

func CleanKubernetesSecret(t *testing.T, name string) {
	if err := KubernetesSecretClient(t).Delete(name, &meta.DeleteOptions{}); err != nil {
		if errors.IsNotFound(err) {
			return
		}

		require.NoError(t, err)
	}

	err := Timeout(250*time.Millisecond, 5*time.Minute, func() error {
		if _, err := KubernetesSecretClient(t).Get(name, meta.GetOptions{}); err!= nil {
			if errors.IsNotFound(err) {
				return Interrupt{}
			}

			return err
		}

		return nil
	})
	require.NoError(t, err)
}

type KubernetesSecretConditional func(t *testing.T, secret *core.Secret, err error) bool

func WaitUntilKubernetesSecretIsPresent(t *testing.T, name string) {
	WaitUntilKubernetesSecret(t, name, KubernetesSecretConditionalIsPresent,
		250*time.Millisecond, 1*time.Minute,
		"Secret %s did not appear in requested time", name)
}

func WaitUntilKubernetesSecretIsMissing(t *testing.T, name string) {
	WaitUntilKubernetesSecret(t, name, KubernetesSecretConditionalIsMissing,
		250*time.Millisecond, 1*time.Minute,
		"Secret %s did not disappear in requested time", name)
}

func WaitUntilKubernetesSecret(t *testing.T,
	name string, conditional KubernetesSecretConditional,
	interval, timeout time.Duration,
	message string, args ... interface{}) {
	err := Timeout(interval, timeout, func() error {
		secretObject, err := KubernetesSecretClient(t).Get(name, meta.GetOptions{})

		if conditional(t, secretObject, err) {
			return Interrupt{}
		}

		return nil
	})
	require.NoErrorf(t, err, message, args...)
}

func KubernetesSecretConditionalIsPresent(t *testing.T, secret *core.Secret, err error) bool {
	if err == nil {
		return true
	}

	if errors.IsNotFound(err) {
		return false
	}

	require.Error(t, err)
	return false
}

func KubernetesSecretConditionalIsMissing(t *testing.T, secret *core.Secret, err error) bool {
	if err == nil {
		return false
	}

	if errors.IsNotFound(err) {
		return true
	}

	require.Error(t, err)
	return false
}