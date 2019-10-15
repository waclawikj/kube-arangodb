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
	"context"
	"github.com/arangodb/go-driver"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func SkipIfNotEnterprise(t *testing.T, client driver.Client, ctx context.Context) {
	version, err := client.Version(ctx)
	require.NoError(t, err)

	if !version.IsEnterprise() {
		t.Skipf("Provided ArangoDB version is not enterprise!")
	}
}

type ArangoClientConditional func(t *testing.T, client driver.Client, ctx context.Context) bool

func WaitUntilArangoClientReady(t *testing.T, client driver.Client) {
	WaitUntilArangoClient(t, client, ArangoDeploymentConditionalApiVersionReady,
		50*time.Millisecond, 45*time.Second,
		"Unable to get version from driver")
	WaitUntilArangoClient(t, client, ArangoDeploymentConditionalHealthy,
		50*time.Millisecond, 45*time.Second,
		"Cluster did not become healthy in time")
}

func WaitUntilArangoClient(t *testing.T,
	client driver.Client, conditional ArangoClientConditional,
	interval, timeout time.Duration,
	message string, args ... interface{}) {
	err := Timeout(interval, timeout, func() error {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if conditional(t, client, ctx) {
			return Interrupt{}
		}

		return nil
	})
	require.NoErrorf(t, err, message, args)
}

func ArangoDeploymentConditionalApiVersionReady(t *testing.T, client driver.Client, ctx context.Context) bool {
	_, err := client.Version(ctx)
	if err != nil {
		return false
	}

	return true
}

func ArangoDeploymentConditionalHealthy(t *testing.T, client driver.Client, ctx context.Context) bool {
	cluster, err := client.Cluster(ctx)
	if err != nil {
		if !driver.IsPreconditionFailed(err) {
			require.NoError(t, err)
			return false
		}

		// If ve are not in cluster it is enough
		return true
	}

	health, err := cluster.Health(ctx)
	require.NoError(t, err)

	for _, status := range health.Health  {
		if status.Status != driver.ServerStatusGood {
			return false
		}
	}

	return true
}