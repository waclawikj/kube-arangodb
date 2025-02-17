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

// Code generated by client-gen. DO NOT EDIT.

package fake

import (
	clientset "github.com/arangodb/kube-arangodb/pkg/generated/clientset/versioned"
	backupv1alpha "github.com/arangodb/kube-arangodb/pkg/generated/clientset/versioned/typed/backup/v1alpha"
	fakebackupv1alpha "github.com/arangodb/kube-arangodb/pkg/generated/clientset/versioned/typed/backup/v1alpha/fake"
	databasev1alpha "github.com/arangodb/kube-arangodb/pkg/generated/clientset/versioned/typed/deployment/v1alpha"
	fakedatabasev1alpha "github.com/arangodb/kube-arangodb/pkg/generated/clientset/versioned/typed/deployment/v1alpha/fake"
	replicationv1alpha "github.com/arangodb/kube-arangodb/pkg/generated/clientset/versioned/typed/replication/v1alpha"
	fakereplicationv1alpha "github.com/arangodb/kube-arangodb/pkg/generated/clientset/versioned/typed/replication/v1alpha/fake"
	storagev1alpha "github.com/arangodb/kube-arangodb/pkg/generated/clientset/versioned/typed/storage/v1alpha"
	fakestoragev1alpha "github.com/arangodb/kube-arangodb/pkg/generated/clientset/versioned/typed/storage/v1alpha/fake"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/discovery"
	fakediscovery "k8s.io/client-go/discovery/fake"
	"k8s.io/client-go/testing"
)

// NewSimpleClientset returns a clientset that will respond with the provided objects.
// It's backed by a very simple object tracker that processes creates, updates and deletions as-is,
// without applying any validations and/or defaults. It shouldn't be considered a replacement
// for a real clientset and is mostly useful in simple unit tests.
func NewSimpleClientset(objects ...runtime.Object) *Clientset {
	o := testing.NewObjectTracker(scheme, codecs.UniversalDecoder())
	for _, obj := range objects {
		if err := o.Add(obj); err != nil {
			panic(err)
		}
	}

	cs := &Clientset{}
	cs.discovery = &fakediscovery.FakeDiscovery{Fake: &cs.Fake}
	cs.AddReactor("*", "*", testing.ObjectReaction(o))
	cs.AddWatchReactor("*", func(action testing.Action) (handled bool, ret watch.Interface, err error) {
		gvr := action.GetResource()
		ns := action.GetNamespace()
		watch, err := o.Watch(gvr, ns)
		if err != nil {
			return false, nil, err
		}
		return true, watch, nil
	})

	return cs
}

// Clientset implements clientset.Interface. Meant to be embedded into a
// struct to get a default implementation. This makes faking out just the method
// you want to test easier.
type Clientset struct {
	testing.Fake
	discovery *fakediscovery.FakeDiscovery
}

func (c *Clientset) Discovery() discovery.DiscoveryInterface {
	return c.discovery
}

var _ clientset.Interface = &Clientset{}

// BackupV1alpha retrieves the BackupV1alphaClient
func (c *Clientset) BackupV1alpha() backupv1alpha.BackupV1alphaInterface {
	return &fakebackupv1alpha.FakeBackupV1alpha{Fake: &c.Fake}
}

// DatabaseV1alpha retrieves the DatabaseV1alphaClient
func (c *Clientset) DatabaseV1alpha() databasev1alpha.DatabaseV1alphaInterface {
	return &fakedatabasev1alpha.FakeDatabaseV1alpha{Fake: &c.Fake}
}

// ReplicationV1alpha retrieves the ReplicationV1alphaClient
func (c *Clientset) ReplicationV1alpha() replicationv1alpha.ReplicationV1alphaInterface {
	return &fakereplicationv1alpha.FakeReplicationV1alpha{Fake: &c.Fake}
}

// StorageV1alpha retrieves the StorageV1alphaClient
func (c *Clientset) StorageV1alpha() storagev1alpha.StorageV1alphaInterface {
	return &fakestoragev1alpha.FakeStorageV1alpha{Fake: &c.Fake}
}
