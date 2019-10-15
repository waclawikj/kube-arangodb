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

package tests

import (
	"context"
	"fmt"
	"github.com/arangodb/kube-arangodb/pkg/util/arangod"
	"github.com/arangodb/kube-arangodb/tests/helper"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/arangodb/go-driver"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1alpha"
)


func Test_New_Deployment(t *testing.T) {
	helper.MarkSmokeTest(t, true)

	iterateOverStorageEngines(t, func(t *testing.T, storageEngine api.StorageEngine) {
		iterateOverModes(t, func(t *testing.T, mode api.DeploymentMode) {
			deploymentSubTest(t, mode, storageEngine)
		})
	})
}

func deploymentSubTest(t *testing.T, mode api.DeploymentMode, engine api.StorageEngine) {
	helper.MarkSmokeTest(t, true)
	deploymentWrapper := helper.NewArangoDeploymentWrapper(fmt.Sprintf("test-deployment-%s-%s", mode, engine), func(deployment *api.ArangoDeployment) {
		deployment.Spec.Mode = api.NewMode(mode)
		deployment.Spec.StorageEngine = api.NewStorageEngine(engine)
	})

	deploymentWrapper.Run(t, func(t *testing.T, deployment *api.ArangoDeployment) {
		helper.WaitUntilArangoDeploymentIsReady(t, deployment.GetName())

		helper.WaitUntilKubernetesSecretIsPresent(t, deployment.Spec.Authentication.GetJWTSecretName())

		ctx := context.Background()
		client := mustNewArangodDatabaseClient(ctx, helper.KubernetesClient(t), deployment, t, nil)

		helper.WaitUntilArangoClientReady(t, client)
	})
}

// test a setup containing multiple deployments
func Test_New_DeploymentMulti(t *testing.T) {
	helper.MarkSmokeTest(t, false)

	deploymentWrapper1 := helper.NewArangoDeploymentWrapper("test-multidepl-1", func(deployment *api.ArangoDeployment) {
		deployment.Spec.Mode = api.NewMode(api.DeploymentModeCluster)
		deployment.Spec.StorageEngine = api.NewStorageEngine(api.StorageEngineRocksDB)
	})

	deploymentWrapper2 := helper.NewArangoDeploymentWrapper("test-multidepl-2", func(deployment *api.ArangoDeployment) {
		deployment.Spec.Mode = api.NewMode(api.DeploymentModeSingle)
		deployment.Spec.StorageEngine = api.NewStorageEngine(api.StorageEngineMMFiles)
	})

	deploymentWrapper1.Run(t, func(t *testing.T, deployment1 *api.ArangoDeployment) {
		deploymentWrapper2.Run(t, func(t *testing.T, deployment2 *api.ArangoDeployment) {
			helper.WaitUntilArangoDeploymentIsReady(t, deployment1.GetName())
			helper.WaitUntilArangoDeploymentIsReady(t, deployment2.GetName())

			ctx := arangod.WithRequireAuthentication(context.Background())

			client1 := mustNewArangodDatabaseClient(ctx, helper.KubernetesClient(t), deployment1, t, nil)
			client2 := mustNewArangodDatabaseClient(ctx, helper.KubernetesClient(t), deployment2, t, nil)

			helper.WaitUntilArangoClientReady(t, client1)
			helper.WaitUntilArangoClientReady(t, client2)

			// Test if we are able to create a collections in both deployments.
			db1, err := client1.Database(ctx, "_system")
			require.NoError(t, err, "failed to get database")
			_, err = db1.CreateCollection(ctx, "col1", nil)
			require.NoError(t, err, "failed to create collection")

			db2, err := client2.Database(ctx, "_system")
			require.NoError(t, err, "failed to get database")
			_, err = db2.CreateCollection(ctx, "col2", nil)
			require.NoError(t, err, "failed to create collection")

			// The newly created collections must be (only) visible in the deployment
			// that it was created in. The following lines ensure this behavior.
			collections1, err := db1.Collections(ctx)
			require.NoError(t, err, "failed to get collections")
			collections2, err := db2.Collections(ctx)
			require.NoError(t, err, "failed to get collections")

			assert.True(t, containsCollection(collections1, "col1"), "collection missing")
			assert.True(t, containsCollection(collections2, "col2"), "collection missing")
			assert.False(t, containsCollection(collections1, "col2"), "collection must not be in this deployment")
			assert.False(t, containsCollection(collections2, "col1"), "collection must not be in this deployment")
		})
	})
}

func containsCollection(colls []driver.Collection, name string) bool {
	for _, col := range colls {
		if name == col.Name() {
			return true
		}
	}
	return false
}
