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

package policy

import (
	backupApi "github.com/arangodb/kube-arangodb/pkg/apis/backup/v1alpha"
	database "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1alpha"
	"github.com/arangodb/kube-arangodb/pkg/backup/operator"
	"github.com/arangodb/kube-arangodb/pkg/backup/operator/event"
	arangoClientSet "github.com/arangodb/kube-arangodb/pkg/generated/clientset/versioned"
	arangoInformer "github.com/arangodb/kube-arangodb/pkg/generated/informers/externalversions"
	"k8s.io/client-go/kubernetes"
)

func newEventInstance(eventRecorder event.Recorder) event.RecorderInstance {
	return eventRecorder.NewInstance(database.SchemeGroupVersion.Group,
		database.SchemeGroupVersion.Version,
		backupApi.ArangoBackupPolicyResourceKind)
}

// RegisterInformer in operator
func RegisterInformer(operator operator.Operator, recorder event.Recorder, client arangoClientSet.Interface, kubeClient kubernetes.Interface, informer arangoInformer.SharedInformerFactory) error {
	if err := operator.RegisterInformer(informer.Backup().V1alpha().ArangoBackupPolicies().Informer(),
		database.SchemeGroupVersion.Group,
		database.SchemeGroupVersion.Version,
		backupApi.ArangoBackupPolicyResourceKind); err != nil {
		return err
	}

	h := &handler{
		client:        client,
		kubeClient:    kubeClient,
		eventRecorder: newEventInstance(recorder),

		operator: operator,
	}

	if err := operator.RegisterHandler(h); err != nil {
		return err
	}

	return nil
}
