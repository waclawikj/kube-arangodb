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

func stateReadyHandler(h *handler, backup *backupApi.ArangoBackup) (*backupApi.ArangoBackupStatus, error) {
	deployment, err := h.getArangoDeploymentObject(backup)
	if err != nil {
		return nil, err
	}

	client, err := h.arangoClientFactory(deployment, backup)
	if err != nil {
		return nil, newTemporaryError(err)
	}

	if backup.Status.Backup == nil {
		return nil, newFatalErrorf("missing field .status.backup")
	}

	backupMeta, err := client.Get(driver.BackupID(backup.Status.Backup.ID))
	if err != nil {
		if driver.IsNotFound(err) {
			return wrapUpdateStatus(backup,
				updateStatusState(backupApi.ArangoBackupStateDeleted, ""),
				updateStatusAvailable(false),
			)
		}

		return wrapUpdateStatus(backup,
			updateStatusAvailable(true),
		)
	}

	var available = backupMeta.Available

	// Temporally check if number of db servers match number of pieces
	available = available && backupMeta.NumberOfDBServers == backupMeta.NumberOfPiecesPresent

	if !available {
		return wrapUpdateStatus(backup,
			updateStatusState(backupApi.ArangoBackupStateUnavailable, ""),
			updateStatusBackup(backupMeta),
			updateStatusAvailable(false),
		)
	}

	// Check if upload flag was specified later in runtime
	if backup.Spec.Upload != nil &&
		(backup.Status.Backup.Uploaded == nil || (backup.Status.Backup.Uploaded != nil && !*backup.Status.Backup.Uploaded)) {
		// Ensure that we can start upload process
		running, err := isBackupRunning(backup, h.client.BackupV1alpha().ArangoBackups(backup.Namespace))
		if err != nil {
			return nil, err
		}

		if running {
			return wrapUpdateStatus(backup,
				updateStatusState(backupApi.ArangoBackupStateReady, "Upload process queued"),
				updateStatusBackup(backupMeta),
				updateStatusAvailable(true),
			)
		}

		return wrapUpdateStatus(backup,
			updateStatusState(backupApi.ArangoBackupStateUpload, ""),
			updateStatusBackup(backupMeta),
			updateStatusAvailable(true),
		)
	}

	if backup.Spec.Upload == nil && backup.Status.Backup.Uploaded != nil {
		return wrapUpdateStatus(backup,
			updateStatusState(backupApi.ArangoBackupStateReady, ""),
			updateStatusBackup(backupMeta),
			updateStatusBackupUpload(nil),
			updateStatusAvailable(true),
		)
	}

	return wrapUpdateStatus(backup,
		updateStatusBackup(backupMeta),
		updateStatusAvailable(true),
	)
}
