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
	"github.com/arangodb/kube-arangodb/pkg/util"
	"testing"
	"time"

	"github.com/arangodb/kube-arangodb/pkg/backup/operator/operation"

	meta "k8s.io/apimachinery/pkg/apis/meta/v1"

	backupApi "github.com/arangodb/kube-arangodb/pkg/apis/backup/v1alpha"
	"github.com/stretchr/testify/require"
)

func Test_State_UploadError_Reschedule(t *testing.T) {
	// Arrange
	handler, mock := newErrorsFakeHandler(mockErrorsArangoClientBackup{})

	obj, deployment := newObjectSet(backupApi.ArangoBackupStateUploadError)

	obj.Spec.Upload = &backupApi.ArangoBackupSpecOperation{
		RepositoryURL: "S3 URL",
	}

	backupMeta, err := mock.Create()
	require.NoError(t, err)

	obj.Status.Backup = &backupApi.ArangoBackupDetails{
		ID:                string(backupMeta.ID),
		Version:           backupMeta.Version,
		CreationTimestamp: meta.Now(),
		Uploaded:          util.NewBool(true),
	}

	obj.Status.Time.Time = time.Now().Add(-2 * downloadDelay)

	// Act
	createArangoDeployment(t, handler, deployment)
	createArangoBackup(t, handler, obj)

	require.NoError(t, handler.Handle(newItemFromBackup(operation.Update, obj)))

	// Assert
	newObj := refreshArangoBackup(t, handler, obj)
	require.Equal(t, newObj.Status.State, backupApi.ArangoBackupStateReady)

	require.True(t, newObj.Status.Available)

	require.NotNil(t, newObj.Status.Backup)
	require.Equal(t, obj.Status.Backup, newObj.Status.Backup)
}

func Test_State_UploadError_Wait(t *testing.T) {
	handler, mock := newErrorsFakeHandler(mockErrorsArangoClientBackup{})

	obj, deployment := newObjectSet(backupApi.ArangoBackupStateUploadError)

	obj.Spec.Upload = &backupApi.ArangoBackupSpecOperation{
		RepositoryURL: "S3 URL",
	}

	backupMeta, err := mock.Create()
	require.NoError(t, err)

	obj.Status.Backup = &backupApi.ArangoBackupDetails{
		ID:                string(backupMeta.ID),
		Version:           backupMeta.Version,
		CreationTimestamp: meta.Now(),
		Uploaded:          util.NewBool(true),
	}

	obj.Status.Time.Time = time.Now().Add(2 * downloadDelay)
	obj.Status.Message = "message"

	// Act
	createArangoDeployment(t, handler, deployment)
	createArangoBackup(t, handler, obj)

	require.NoError(t, handler.Handle(newItemFromBackup(operation.Update, obj)))

	// Assert
	newObj := refreshArangoBackup(t, handler, obj)
	checkBackup(t, newObj, backupApi.ArangoBackupStateUploadError, true)

	require.Equal(t, "message", newObj.Status.Message)

	require.NotNil(t, newObj.Status.Backup)
	require.Equal(t, obj.Status.Backup, newObj.Status.Backup)
}

func Test_State_UploadError_BackToReady(t *testing.T) {
	handler, mock := newErrorsFakeHandler(mockErrorsArangoClientBackup{})

	obj, deployment := newObjectSet(backupApi.ArangoBackupStateUploadError)

	backupMeta, err := mock.Create()
	require.NoError(t, err)

	obj.Status.Backup = &backupApi.ArangoBackupDetails{
		ID:                string(backupMeta.ID),
		Version:           backupMeta.Version,
		CreationTimestamp: meta.Now(),
		Uploaded:          util.NewBool(true),
	}

	obj.Status.Time.Time = time.Now().Add(2 * downloadDelay)

	// Act
	createArangoDeployment(t, handler, deployment)
	createArangoBackup(t, handler, obj)

	require.NoError(t, handler.Handle(newItemFromBackup(operation.Update, obj)))

	// Assert
	newObj := refreshArangoBackup(t, handler, obj)
	checkBackup(t, newObj, backupApi.ArangoBackupStateReady, true)

	require.NotNil(t, newObj.Status.Backup)
	require.Equal(t, obj.Status.Backup, newObj.Status.Backup)
}
