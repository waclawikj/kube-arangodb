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
	"testing"

	"github.com/arangodb/kube-arangodb/pkg/backup/operator/operation"

	backupApi "github.com/arangodb/kube-arangodb/pkg/apis/backup/v1alpha"
	"github.com/stretchr/testify/require"
)

func Test_State_Upload_Common(t *testing.T) {
	wrapperUndefinedDeployment(t, backupApi.ArangoBackupStateUpload)
	wrapperConnectionIssues(t, backupApi.ArangoBackupStateUpload)
}

func Test_State_Upload_Success(t *testing.T) {
	// Arrange
	handler, mock := newErrorsFakeHandler(mockErrorsArangoClientBackup{})

	obj, deployment := newObjectSet(backupApi.ArangoBackupStateUpload)

	createResponse, err := mock.Create()
	require.NoError(t, err)

	backupMeta, err := mock.Get(createResponse.ID)
	require.NoError(t, err)

	obj.Status.Backup = createBackupFromMeta(backupMeta, nil)

	// Act
	createArangoDeployment(t, handler, deployment)
	createArangoBackup(t, handler, obj)

	require.NoError(t, handler.Handle(newItemFromBackup(operation.Update, obj)))

	// Assert
	newObj := refreshArangoBackup(t, handler, obj)
	checkBackup(t, newObj, backupApi.ArangoBackupStateUploading, true)

	require.NotNil(t, newObj.Status.Progress)
	progresses := mock.getProgressIDs()
	require.Len(t, progresses, 1)
	require.Equal(t, progresses[0], newObj.Status.Progress.JobID)

	compareBackupMeta(t, backupMeta, obj)
}

func Test_State_Upload_TemporaryGetFailed(t *testing.T) {
	// Arrange
	error := newTemporaryErrorf("error")
	handler, mock := newErrorsFakeHandler(mockErrorsArangoClientBackup{
		getError: error,
	})

	obj, deployment := newObjectSet(backupApi.ArangoBackupStateUpload)

	createResponse, err := mock.Create()
	require.NoError(t, err)

	obj.Status.Backup = createBackupFromMeta(driver.BackupMeta{
		ID: createResponse.ID,
	}, nil)

	// Act
	createArangoDeployment(t, handler, deployment)
	createArangoBackup(t, handler, obj)

	require.EqualError(t, handler.Handle(newItemFromBackup(operation.Update, obj)), error.Error())

	// Assert
	newObj := refreshArangoBackup(t, handler, obj)
	require.Equal(t, obj.Status, newObj.Status)
}

func Test_State_Upload_FatalGetFailed(t *testing.T) {
	// Arrange
	error := newFatalErrorf("error")
	handler, mock := newErrorsFakeHandler(mockErrorsArangoClientBackup{
		getError: error,
	})

	obj, deployment := newObjectSet(backupApi.ArangoBackupStateUpload)

	createResponse, err := mock.Create()
	require.NoError(t, err)

	obj.Status.Backup = createBackupFromMeta(driver.BackupMeta{
		ID: createResponse.ID,
	}, nil)

	// Act
	createArangoDeployment(t, handler, deployment)
	createArangoBackup(t, handler, obj)

	require.EqualError(t, handler.Handle(newItemFromBackup(operation.Update, obj)), error.Error())

	// Assert
	newObj := refreshArangoBackup(t, handler, obj)
	require.Equal(t, obj.Status, newObj.Status)
}

func Test_State_Upload_BackupMissing(t *testing.T) {
	// Arrange
	error := driver.ArangoError{
		Code: 404,
	}
	handler, mock := newErrorsFakeHandler(mockErrorsArangoClientBackup{
		getError: error,
	})

	obj, deployment := newObjectSet(backupApi.ArangoBackupStateUpload)

	createResponse, err := mock.Create()
	require.NoError(t, err)

	obj.Status.Backup = createBackupFromMeta(driver.BackupMeta{
		ID: createResponse.ID,
	}, nil)

	// Act
	createArangoDeployment(t, handler, deployment)
	createArangoBackup(t, handler, obj)

	require.NoError(t, handler.Handle(newItemFromBackup(operation.Update, obj)))

	// Assert
	newObj := refreshArangoBackup(t, handler, obj)
	checkBackup(t, newObj, backupApi.ArangoBackupStateDeleted, false)
}

func Test_State_Upload_TemporaryUploadFailed(t *testing.T) {
	// Arrange
	error := newTemporaryErrorf("error")
	handler, mock := newErrorsFakeHandler(mockErrorsArangoClientBackup{
		uploadError: error,
	})

	obj, deployment := newObjectSet(backupApi.ArangoBackupStateUpload)

	createResponse, err := mock.Create()
	require.NoError(t, err)

	obj.Status.Backup = createBackupFromMeta(driver.BackupMeta{
		ID: createResponse.ID,
	}, nil)

	// Act
	createArangoDeployment(t, handler, deployment)
	createArangoBackup(t, handler, obj)

	require.NoError(t, handler.Handle(newItemFromBackup(operation.Update, obj)))

	// Assert
	newObj := refreshArangoBackup(t, handler, obj)
	checkBackup(t, newObj, backupApi.ArangoBackupStateUploadError, true)
}

func Test_State_Upload_FatalUploadFailed(t *testing.T) {
	// Arrange
	error := newFatalErrorf("error")
	handler, mock := newErrorsFakeHandler(mockErrorsArangoClientBackup{
		uploadError: error,
	})

	obj, deployment := newObjectSet(backupApi.ArangoBackupStateUpload)

	createResponse, err := mock.Create()
	require.NoError(t, err)

	obj.Status.Backup = createBackupFromMeta(driver.BackupMeta{
		ID: createResponse.ID,
	}, nil)

	// Act
	createArangoDeployment(t, handler, deployment)
	createArangoBackup(t, handler, obj)

	require.NoError(t, handler.Handle(newItemFromBackup(operation.Update, obj)))

	// Assert
	newObj := refreshArangoBackup(t, handler, obj)
	checkBackup(t, newObj, backupApi.ArangoBackupStateUploadError, true)
}
