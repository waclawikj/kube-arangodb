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
	"github.com/arangodb/arangosync-client/client"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

type ArangoSyncClientConditional func(t *testing.T, client client.API, ctx context.Context) bool

func WaitUntilArangoSyncClientReady(t *testing.T, client client.API, syncMaster, syncWorker int) {
	WaitUntilArangoSyncClient(t, client, ArangoDeploymentConditionalSyncApiVersionReady,
		50*time.Millisecond, 45*time.Second,
		"Unable to get version from driver")
	WaitUntilArangoSyncClient(t, client, ArangoDeploymentConditionalSyncMasterReached(syncMaster),
		50*time.Millisecond, time.Second,
		"Masters did not appear")
	WaitUntilArangoSyncClient(t, client, ArangoDeploymentConditionalSyncWorkerReached(syncWorker),
		50*time.Millisecond, time.Second,
		"Workers did not appear")
}

func WaitUntilArangoSyncClient(t *testing.T,
	client client.API, conditional ArangoSyncClientConditional,
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
	require.NoErrorf(t, err, message, args...)
}

func ArangoDeploymentConditionalSyncApiVersionReady(t *testing.T, client client.API, ctx context.Context) bool {
	_, err := client.Version(ctx)
	if err != nil {
		return false
	}

	return true
}

func ArangoDeploymentConditionalSyncMasterReached(syncMasters int) func(t *testing.T, client client.API, ctx context.Context) bool {
	return func(t *testing.T, client client.API, ctx context.Context) bool {
		masters, err := client.Master().Masters(ctx)
		require.NoError(t, err)

		return len(masters) == syncMasters
	}
}

func ArangoDeploymentConditionalSyncWorkerReached(syncWorkers int) func(t *testing.T, client client.API, ctx context.Context) bool {
	return func(t *testing.T, client client.API, ctx context.Context) bool {
		workers, err := client.Master().RegisteredWorkers(ctx)
		require.NoError(t, err)

		return len(workers) == syncWorkers
	}
}