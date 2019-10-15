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
	"fmt"
	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1alpha"
	arangoDeploymentTyped "github.com/arangodb/kube-arangodb/pkg/generated/clientset/versioned/typed/deployment/v1alpha"
	"github.com/arangodb/kube-arangodb/pkg/util"
	"github.com/dchest/uniuri"
	"github.com/stretchr/testify/require"
	core "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	typedCore "k8s.io/client-go/kubernetes/typed/core/v1"
	"strings"
	"testing"
	"time"
)

func ArangoDeploymentClient(t *testing.T) arangoDeploymentTyped.ArangoDeploymentInterface {
	return ArangoClient(t).DatabaseV1alpha().ArangoDeployments(Namespace(t))
}

type ArangoDeploymentMod func(deployment *api.ArangoDeployment)

func NewArangoDeployment(t *testing.T, prefix string) *api.ArangoDeployment {
	name := fmt.Sprintf("%s-%s", prefix, uniuri.NewLen(4))

	depl := &api.ArangoDeployment{
		TypeMeta: meta.TypeMeta{
			APIVersion: api.SchemeGroupVersion.String(),
			Kind:       api.ArangoDeploymentResourceKind,
		},
		ObjectMeta: meta.ObjectMeta{
			Name: strings.ToLower(name),
		},
		Spec: api.DeploymentSpec{
			ImagePullPolicy: util.NewPullPolicy(core.PullAlways),
			License: api.LicenseSpec{
				SecretName: util.NewString(LicenseSecret(t)),
			},
			Image: util.NewString(Image(t)),
			Agents: api.ServerGroupSpec{
				Count: util.NewInt(3),
			},
			DBServers: api.ServerGroupSpec{
				Count: util.NewInt(2),
			},
			Coordinators: api.ServerGroupSpec{
				Count: util.NewInt(2),
			},
		},
	}

	return depl
}

func RefreshDeployment(t *testing.T, deployment *api.ArangoDeployment) *api.ArangoDeployment {
	newDeployment, err := ArangoDeploymentClient(t).Get(deployment.GetName(), meta.GetOptions{})
	require.NoError(t, err)
	return newDeployment
}

func CleanArangoDeployment(t *testing.T, name string) {
	if err := ArangoDeploymentClient(t).Delete(name, &meta.DeleteOptions{}); err != nil {
		if errors.IsNotFound(err) {
			return
		}

		require.NoError(t, err)
	}

	err := Timeout(250*time.Millisecond, 5*time.Minute, func() error {
		if _, err := ArangoDeploymentClient(t).Get(name, meta.GetOptions{}); err!= nil {
			if errors.IsNotFound(err) {
				return Interrupt{}
			}

			return err
		}

		return nil
	})
	require.NoError(t, err)
}

func UpdateArangoDeployment(t *testing.T, name string, mod ArangoDeploymentMod) {
	deploymentClient := ArangoDeploymentClient(t)
	podClient := KubernetesPodClient(t)

	err := Retry(100*time.Millisecond, 10, func() error {
		deployment, err := deploymentClient.Get(name, meta.GetOptions{})
		if err != nil {
			return err
		}

		mod(deployment)

		_, err = deploymentClient.Update(deployment)
		return err
	})
	require.NoError(t, err)

	t.Logf("Deployment updated")

	// Wait for deployment to become ready
	err = Timeout(100*time.Millisecond, 5*time.Minute, func() error {
		deployment, err := deploymentClient.Get(name, meta.GetOptions{})
		if err != nil {
			return err
		}

		if deployment.Spec.Mode == nil {
			return fmt.Errorf("Deployment mode cannot be nil")
		}

		switch *deployment.Spec.Mode {
		case api.DeploymentModeSingle:
			if !equalArangoDeploymentMember(podClient, deployment.Spec.Single, deployment.Status.Members.Single) {
				return nil
			}
		case api.DeploymentModeCluster, api.DeploymentModeActiveFailover:
			if !(equalArangoDeploymentMember(podClient, deployment.Spec.DBServers, deployment.Status.Members.DBServers) &&
				equalArangoDeploymentMember(podClient, deployment.Spec.Agents, deployment.Status.Members.Agents) &&
				equalArangoDeploymentMember(podClient, deployment.Spec.Coordinators, deployment.Status.Members.Coordinators)) {
				return nil
			}
		}

		if deployment.Spec.Sync.Enabled != nil && *deployment.Spec.Sync.Enabled {
			if !(equalArangoDeploymentMember(podClient, deployment.Spec.SyncMasters, deployment.Status.Members.SyncMasters) &&
				equalArangoDeploymentMember(podClient, deployment.Spec.SyncWorkers, deployment.Status.Members.SyncWorkers)) {
				return nil
			}
		}

		return Interrupt{}
	})
	require.NoError(t, err)

	t.Logf("Deployment scaled")
}

func equalArangoDeploymentMember(podClient typedCore.PodInterface,
	spec api.ServerGroupSpec,
	status api.MemberStatusList) bool {
	if spec.Count == nil {
		return false
	}

	if *spec.Count != len(status) {
		return false
	}

	for _, s := range status {
		if s.Phase != api.MemberPhaseCreated {
			return false
		}

		if len(s.Conditions) == 0 {
			return false
		}

		if !s.Conditions.IsTrue(api.ConditionTypeReady) {
			return false
		}

		pod, err := podClient.Get(s.PodName, meta.GetOptions{})
		if err != nil {
			return false
		}

		container, ok := getContainerByName("server", pod.Spec.Containers)
		if !ok {
			return false
		}

		if spec.Args != nil && len(spec.Args) > 0 {
			if container.Args == nil {
				return false
			}

			containerStart := len(container.Args) - len(spec.Args)

			for id, arg := range spec.Args {
				if container.Args[containerStart+id] != arg {
					return false
				}
			}
		}
	}

	return true
}

func getContainerByName(name string, containers []core.Container) (core.Container, bool) {
	for _, c := range containers{
		if c.Name == name {
			return c, true
		}
	}

	return core.Container{}, false
}

type arangoDeploymentWrapper struct {
	mods []ArangoDeploymentMod
	prefix string
}

func (d*arangoDeploymentWrapper) Run(t *testing.T, action func(t *testing.T, deployment *api.ArangoDeployment)) {
	deployment := NewArangoDeployment(t, d.prefix)

	for _, a := range d.mods {
		a(deployment)
	}

	deployment.Spec.SetDefaults(deployment.GetName())

	defer CleanArangoDeployment(t, deployment.GetName())
	deploymentObject, err := ArangoDeploymentClient(t).Create(deployment)
	require.NoError(t, err)

	action(t, deploymentObject)
}

type ArangoDeploymentWrapper interface {
	Run(t *testing.T, action func(t *testing.T, deployment *api.ArangoDeployment))
}

func NewArangoDeploymentWrapper(prefix string, mods ...ArangoDeploymentMod) ArangoDeploymentWrapper {
	return &arangoDeploymentWrapper{
		prefix: prefix,
		mods:   mods,
	}
}

type ArangoDeploymentConditional func(t *testing.T, deployment *api.ArangoDeployment, err error) bool

func WaitUntilArangoDeploymentIsReady(t *testing.T, name string) {
	WaitUntilArangoDeployment(t, name, ArangoDeploymentConditionalIsReady,
		250*time.Millisecond, 5*time.Minute,
		"ArangoDeployment %s did not start in expected time", name)
}

func WaitUntilArangoDeployment(t *testing.T,
	name string, conditional ArangoDeploymentConditional,
	interval, timeout time.Duration,
	message string, args ... interface{}) {
	err := Timeout(interval, timeout, func() error {
		deploymentObject, err := ArangoDeploymentClient(t).Get(name, meta.GetOptions{})

		if conditional(t, deploymentObject, err) {
			return Interrupt{}
		}

		return nil
	})
	require.NoErrorf(t, err, message, args...)
}

func ArangoDeploymentConditionalIsReady(t *testing.T, deployment *api.ArangoDeployment, err error) bool {
	require.NoError(t, err)

	if deployment.Status.Phase == api.DeploymentPhaseFailed {
		require.Fail(t, "Expected Running phase, got %s", deployment.Status.Phase)
	}
	if deployment.Status.Phase != api.DeploymentPhaseRunning {
		return false
	}
	if deployment.Status.Conditions.IsTrue(api.ConditionTypeReady) {
		return true
	}
	return false
}