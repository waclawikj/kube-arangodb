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
	"github.com/dchest/uniuri"
	"github.com/stretchr/testify/require"
	"strings"
	"testing"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1alpha"
	"github.com/arangodb/kube-arangodb/pkg/util"
	"github.com/arangodb/kube-arangodb/pkg/util/arangod"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil"
)

func testAuthenticationRunForModes(t *testing.T, action func(t *testing.T, mode api.DeploymentMode)) {
	iterateOverModes(t, func(t *testing.T, mode api.DeploymentMode) {
		if mode == api.DeploymentModeActiveFailover {
			t.Skipf("Test is not valid for mode %s", mode)
		}

		action(t, mode)
	})
}

// TestAuthenticationDefaultSecret creating a server
// with default authentication (on) using a generated JWT secret.
func Test_New_AuthenticationDefaultSecret(t *testing.T) {
	helper.MarkSmokeTest(t, true)

	testAuthenticationRunForModes(t, func(t *testing.T, mode api.DeploymentMode) {
		helper.MarkSmokeTest(t, true)

		deploymentWrapper := helper.NewArangoDeploymentWrapper("test-auth-def", func(deployment *api.ArangoDeployment) {
			deployment.Spec.Mode = api.NewMode(mode)
		})

		deploymentWrapper.Run(t, func(t *testing.T, deployment *api.ArangoDeployment) {
			helper.WaitUntilArangoDeploymentIsReady(t, deployment.GetName())

			helper.WaitUntilKubernetesSecretIsPresent(t, deployment.Spec.Authentication.GetJWTSecretName())

			ctx := arangod.WithRequireAuthentication(context.Background())
			client := mustNewArangodDatabaseClient(ctx, helper.KubernetesClient(t), deployment, t, nil)

			helper.WaitUntilArangoClientReady(t, client)

			helper.CleanArangoDeployment(t, deployment.GetName())

			helper.WaitUntilKubernetesSecretIsMissing(t, deployment.Spec.Authentication.GetJWTSecretName())
		})
	})
}

// TestAuthenticationCustomSecret creating a server
// with default authentication (on) using a user created JWT secret.
func Test_New_AuthenticationCustomSecret(t *testing.T) {
	helper.MarkSmokeTest(t, true)

	testAuthenticationRunForModes(t, func(t *testing.T, mode api.DeploymentMode) {
		helper.MarkSmokeTest(t, true)

		secretName := strings.ToLower(uniuri.New())
		defer helper.CleanKubernetesSecret(t, secretName)
		err := k8sutil.CreateTokenSecret(helper.KubernetesSecretClient(t), secretName, "foo", nil)
		require.NoError(t, err)

		deploymentWrapper := helper.NewArangoDeploymentWrapper("test-auth-cst", func(deployment *api.ArangoDeployment) {
			deployment.Spec.Mode = api.NewMode(mode)
			deployment.Spec.Authentication.JWTSecretName = &secretName
		})

		helper.WaitUntilKubernetesSecretIsPresent(t, secretName)

		deploymentWrapper.Run(t, func(t *testing.T, deployment *api.ArangoDeployment) {
			helper.WaitUntilArangoDeploymentIsReady(t, deployment.GetName())

			ctx := arangod.WithRequireAuthentication(context.Background())
			client := mustNewArangodDatabaseClient(ctx, helper.KubernetesClient(t), deployment, t, nil)

			helper.WaitUntilArangoClientReady(t, client)

			helper.CleanArangoDeployment(t, deployment.GetName())
		})

		// Secret must still exists
		helper.WaitUntilKubernetesSecretIsPresent(t, secretName)
	})
}

// TestAuthenticationNone creating a server
// with authentication set to `None`.
func Test_New_AuthenticationNone(t *testing.T) {
	helper.MarkSmokeTest(t, true)

	testAuthenticationRunForModes(t, func(t *testing.T, mode api.DeploymentMode) {
		helper.MarkSmokeTest(t, true)

		deploymentWrapper := helper.NewArangoDeploymentWrapper("test-auth-def", func(deployment *api.ArangoDeployment) {
			deployment.Spec.Mode = api.NewMode(mode)
			deployment.Spec.Authentication.JWTSecretName = util.NewString(api.JWTSecretNameDisabled)
		})

		deploymentWrapper.Run(t, func(t *testing.T, deployment *api.ArangoDeployment) {
			helper.WaitUntilArangoDeploymentIsReady(t, deployment.GetName())

			ctx := arangod.WithSkipAuthentication(context.Background())
			client := mustNewArangodDatabaseClient(ctx, helper.KubernetesClient(t), deployment, t, nil)

			helper.WaitUntilArangoClientReady(t, client)
		})
	})
}