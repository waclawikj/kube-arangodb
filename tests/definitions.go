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
	"fmt"
	"github.com/arangodb/go-driver"
	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1alpha"
	"testing"
)

var (
	modes = []api.DeploymentMode{api.DeploymentModeSingle, api.DeploymentModeCluster, api.DeploymentModeActiveFailover}
 	storageEngines = []api.StorageEngine{api.StorageEngineMMFiles, api.StorageEngineRocksDB}

 	modeServerRole = map[api.DeploymentMode]driver.ServerRole {
		api.DeploymentModeSingle: driver.ServerRoleSingle,
		api.DeploymentModeCluster: driver.ServerRoleCoordinator,
		api.DeploymentModeActiveFailover: driver.ServerRoleSingleActive,
	}
)

func iterateOverModes(t *testing.T, action func(t *testing.T, mode api.DeploymentMode)) {
	for _, mode := range modes {
		t.Run(fmt.Sprintf("%s", mode), func(t *testing.T) {
			action(t, mode)
		})
	}
}

func iterateOverStorageEngines(t *testing.T, action func(t *testing.T,  storageEngine api.StorageEngine)) {
	for _,  storageEngine := range storageEngines {
		t.Run(fmt.Sprintf("%s",  storageEngine), func(t *testing.T) {
			action(t,  storageEngine)
		})
	}
}