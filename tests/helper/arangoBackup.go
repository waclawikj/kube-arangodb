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
	"fmt"
	"github.com/arangodb/go-driver"
	backupApi "github.com/arangodb/kube-arangodb/pkg/apis/backup/v1alpha"
	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1alpha"
	"github.com/arangodb/kube-arangodb/pkg/backup/state"
	arangoBackupTyped "github.com/arangodb/kube-arangodb/pkg/generated/clientset/versioned/typed/backup/v1alpha"
	"github.com/dchest/uniuri"
	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/api/errors"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"strings"
	"testing"
	"time"
)

func ArangoBackupClient(t *testing.T) arangoBackupTyped.ArangoBackupInterface {
	return ArangoClient(t).BackupV1alpha().ArangoBackups(Namespace(t))
}

func SkipIfNotBackup(t *testing.T, client driver.Client, ctx context.Context) {
	_, err := client.Backup().List(ctx, nil)
	if err == nil {
		return
	}

	if driver.IsNotFound(err) {
		t.Skipf("Backup api is not available")
		return
	}

	require.NoError(t, err)
}

func GetBackupByName(t *testing.T, backupName string) *backupApi.ArangoBackup {
	backup, err := ArangoBackupClient(t).Get(backupName, meta.GetOptions{})
	require.NoError(t, err)
	return backup
}

func GetMetaByBackup(t *testing.T, client driver.Client, ctx context.Context, backup *backupApi.ArangoBackup) driver.BackupMeta {
	require.NotNil(t, backup.Status.Backup)

	backups, err := client.Backup().List(ctx, nil)
	require.NoError(t, err)

	meta, ok := backups[driver.BackupID(backup.Status.Backup.ID)]
	require.True(t, ok)

	return meta
}

func RefreshArangoBackup(t *testing.T, backup *backupApi.ArangoBackup) *backupApi.ArangoBackup {
	newBackup, err := ArangoBackupClient(t).Get(backup.GetName(), meta.GetOptions{})
	require.NoError(t, err)
	return newBackup
}

func CreateArangoBackup(t *testing.T, backup *backupApi.ArangoBackup) *backupApi.ArangoBackup {
	_, err := ArangoBackupClient(t).Create(backup)
	require.NoError(t, err)
	return RefreshArangoBackup(t, backup)
}

func CompareBackup(t *testing.T, meta driver.BackupMeta, backup *backupApi.ArangoBackup) {
	require.NotNil(t, backup.Status.Backup)
	require.Equal(t, meta.Version, backup.Status.Backup.Version)
	require.True(t, meta.SizeInBytes > 0)
	require.True(t, meta.SizeInBytes == backup.Status.Backup.SizeInBytes)
	require.True(t, meta.NumberOfDBServers == backup.Status.Backup.NumberOfDBServers)
}

func CleanArangoBackup(t *testing.T, name string) {
	if err := ArangoBackupClient(t).Delete(name, &meta.DeleteOptions{}); err != nil {
		if errors.IsNotFound(err) {
			return
		}

		require.NoError(t, err)
	}

	err := Timeout(250*time.Millisecond, 5*time.Minute, func() error {
		if _, err := ArangoBackupClient(t).Get(name, meta.GetOptions{}); err!= nil {
			if errors.IsNotFound(err) {
				return Interrupt{}
			}

			return err
		}

		return nil
	})
	require.NoError(t, err)
}

func NewArangoBackup(t *testing.T, prefix string, deployment *api.ArangoDeployment) *backupApi.ArangoBackup {
	name := fmt.Sprintf("%s-%s", prefix, uniuri.NewLen(4))

	depl := &backupApi.ArangoBackup{
		ObjectMeta: meta.ObjectMeta{
			Name: strings.ToLower(name),
		},
		Spec: backupApi.ArangoBackupSpec{
			Deployment:backupApi.ArangoBackupSpecDeployment{
				Name: deployment.GetName(),
			},
		},
	}

	return depl
}

type ArangoBackupConditional func(t *testing.T, backup *backupApi.ArangoBackup, err error) bool

func WaitUntilArangoBackupsAreReady(t *testing.T, names ... string) {
	iterateOverString(t, WaitUntilArangoBackupIsReady, names...)
}

func WaitUntilArangoBackupIsReady(t *testing.T, name string) {
	wrapStateChange(t, name, backupApi.ArangoBackupStateReady)
}

func wrapStateChange(t *testing.T, name string, state state.State) {
	WaitUntilArangoBackup(t, name, ArangoBackupConditionalState(state),
		250*time.Millisecond, 5*time.Minute,
		"ArangoBackup %s did not go into state %s", name, state)
}

func WaitUntilArangoBackup(t *testing.T,
	name string, conditional ArangoBackupConditional,
	interval, timeout time.Duration,
	message string, args ... interface{}) {
	err := Timeout(interval, timeout, func() error {
		backupObject, err := ArangoBackupClient(t).Get(name, meta.GetOptions{})

		if conditional(t, backupObject, err) {
			return Interrupt{}
		}

		return nil
	})
	require.NoErrorf(t, err, message, args...)
}

func ArangoBackupConditionalState(state state.State) func(t *testing.T, backup *backupApi.ArangoBackup, err error) bool {
	return func(t *testing.T, backup *backupApi.ArangoBackup, err error) bool {
		require.NoError(t, err)

		require.NotEqual(t, backupApi.ArangoBackupStateFailed, backup.Status.ArangoBackupState.State)

		return backup.Status.ArangoBackupState.State == state
	}
}