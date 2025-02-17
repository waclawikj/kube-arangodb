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
	"github.com/arangodb/go-driver"
	backupApi "github.com/arangodb/kube-arangodb/pkg/apis/backup/v1alpha"
)

func stateCreateHandler(h *handler, backup *backupApi.ArangoBackup) (*backupApi.ArangoBackupStatus, error) {
	deployment, err := h.getArangoDeploymentObject(backup)
	if err != nil {
		return nil, err
	}

	client, err := h.arangoClientFactory(deployment, backup)
	if err != nil {
		return nil, newTemporaryError(err)
	}

	response, err := client.Create()
	if err != nil {
		return nil, err
	}

	backupMeta, err := client.Get(response.ID)
	if err != nil {
		if driver.IsNotFound(err) {
			return wrapUpdateStatus(backup,
				updateStatusState(backupApi.ArangoBackupStateFailed,
					"backup is not present after creation"),
				cleanStatusJob(),
			)
		}

		return nil, newFatalError(err)
	}

	return wrapUpdateStatus(backup,
		updateStatusState(backupApi.ArangoBackupStateReady,""),
		updateStatusAvailable(true),
		updateStatusBackup(backupMeta),
	)
}
