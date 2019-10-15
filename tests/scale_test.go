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
// Author Ewout Prangsma
//

package tests

import (
	"context"
	"github.com/arangodb/kube-arangodb/tests/helper"
	"github.com/stretchr/testify/assert"
	"testing"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1alpha"
	"github.com/arangodb/kube-arangodb/pkg/util"
)

// TestScaleCluster tests scaling up/down the number of DBServers & coordinators
// of a cluster.
func Test_New_ScaleNonTLS(t *testing.T) {
	helper.MarkSmokeTest(t, true)

	deploymentWrapper := helper.NewArangoDeploymentWrapper("test-scale-non-tls", func(deployment *api.ArangoDeployment) {
		deployment.Spec.Mode = api.NewMode(api.DeploymentModeCluster)
		deployment.Spec.TLS = api.TLSSpec{CASecretName: util.NewString("None")}
	})

	deploymentWrapper.Run(t, func(t *testing.T, deployment *api.ArangoDeployment) {
		helper.WaitUntilArangoDeploymentIsReady(t, deployment.GetName())

		helper.WaitUntilKubernetesSecretIsPresent(t, deployment.Spec.Authentication.GetJWTSecretName())

		ctx := context.Background()
		client := mustNewArangodDatabaseClient(ctx, helper.KubernetesClient(t), deployment, t, nil)

		helper.WaitUntilArangoClientReady(t, client)

		t.Run("ScaleUP", func(t *testing.T) {

			helper.UpdateArangoDeployment(t, deployment.Name, func(deployment *api.ArangoDeployment) {
				deployment.Spec.DBServers.Count = util.NewInt(5)
				deployment.Spec.Coordinators.Count = util.NewInt(4)
			})

			helper.WaitUntilArangoDeploymentIsReady(t, deployment.GetName())

			ctx := context.Background()
			client := mustNewArangodDatabaseClient(ctx, helper.KubernetesClient(t), deployment, t, nil)

			helper.WaitUntilArangoClientReady(t, client)

			newDeployment := helper.RefreshDeployment(t, deployment)

			assert.Len(t, newDeployment.Status.Members.DBServers, 5)
			assert.Len(t, newDeployment.Status.Members.Coordinators, 4)
		})

		t.Run("ScaleDown", func(t *testing.T) {

			helper.UpdateArangoDeployment(t, deployment.Name, func(deployment *api.ArangoDeployment) {
				deployment.Spec.DBServers.Count = util.NewInt(3)
				deployment.Spec.Coordinators.Count = util.NewInt(2)
			})

			helper.WaitUntilArangoDeploymentIsReady(t, deployment.GetName())

			ctx := context.Background()
			client := mustNewArangodDatabaseClient(ctx, helper.KubernetesClient(t), deployment, t, nil)

			helper.WaitUntilArangoClientReady(t, client)

			newDeployment := helper.RefreshDeployment(t, deployment)

			assert.Len(t, newDeployment.Status.Members.DBServers, 3)
			assert.Len(t, newDeployment.Status.Members.Coordinators, 2)
		})
	})
}

func Test_New_ScaleWithSync(t *testing.T) {
	helper.MarkIntegrationTest(t, true)

	deploymentWrapper := helper.NewArangoDeploymentWrapper("test-scale-sync", func(deployment *api.ArangoDeployment) {
		deployment.Spec.Mode = api.NewMode(api.DeploymentModeCluster)
		deployment.Spec.Sync.Enabled = util.NewBool(true)
	})

	deploymentWrapper.Run(t, func(t *testing.T, deployment *api.ArangoDeployment) {
		helper.WaitUntilArangoDeploymentIsReady(t, deployment.GetName())

		helper.WaitUntilKubernetesSecretIsPresent(t, deployment.Spec.Authentication.GetJWTSecretName())

		ctx := context.Background()
		client := mustNewArangodDatabaseClient(ctx, helper.KubernetesClient(t), deployment, t, nil)

		helper.WaitUntilArangoClientReady(t, client)

		helper.SkipIfNotEnterprise(t, client, ctx)

		t.Run("ScaleUP", func(t *testing.T) {

			helper.UpdateArangoDeployment(t, deployment.Name, func(deployment *api.ArangoDeployment) {
				deployment.Spec.DBServers.Count = util.NewInt(4)
				deployment.Spec.SyncWorkers.Count = util.NewInt(4)
				deployment.Spec.SyncMasters.Count = util.NewInt(5)
			})

			helper.WaitUntilArangoDeploymentIsReady(t, deployment.GetName())

			ctx := context.Background()
			client := mustNewArangodDatabaseClient(ctx, helper.KubernetesClient(t), deployment, t, nil)

			helper.WaitUntilArangoClientReady(t, client)

			newDeployment := helper.RefreshDeployment(t, deployment)

			assert.Len(t, newDeployment.Status.Members.DBServers, 4)
			assert.Len(t, newDeployment.Status.Members.SyncWorkers, 4)
			assert.Len(t, newDeployment.Status.Members.SyncMasters, 5)

			syncClient := mustNewArangoSyncClient(ctx, helper.KubernetesClient(t), deployment, t)
			helper.WaitUntilArangoSyncClientReady(t, syncClient, 5, 4)
		})

		t.Run("ScaleDown", func(t *testing.T) {

			helper.UpdateArangoDeployment(t, deployment.Name, func(deployment *api.ArangoDeployment) {
				deployment.Spec.DBServers.Count = util.NewInt(2)
				deployment.Spec.SyncWorkers.Count = util.NewInt(2)
				deployment.Spec.SyncMasters.Count = util.NewInt(3)
			})

			helper.WaitUntilArangoDeploymentIsReady(t, deployment.GetName())

			ctx := context.Background()
			client := mustNewArangodDatabaseClient(ctx, helper.KubernetesClient(t), deployment, t, nil)

			helper.WaitUntilArangoClientReady(t, client)

			newDeployment := helper.RefreshDeployment(t, deployment)

			assert.Len(t, newDeployment.Status.Members.DBServers, 2)
			assert.Len(t, newDeployment.Status.Members.SyncWorkers, 2)
			assert.Len(t, newDeployment.Status.Members.SyncMasters, 3)

			syncClient := mustNewArangoSyncClient(ctx, helper.KubernetesClient(t), deployment, t)
			helper.WaitUntilArangoSyncClientReady(t, syncClient, 3, 2)
		})
	})
}