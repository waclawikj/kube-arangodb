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
	arangoClient "github.com/arangodb/kube-arangodb/pkg/generated/clientset/versioned"
	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"os"
	"testing"
)

const (
	NAMESPACE_ENV = "TEST_NAMESPACE"
	IMAGE_ENV = "ARANGODIMAGE"
	LICENSE_SECRET = "arangodb-jenkins-license-key"
)

func KubernetesConfig(t *testing.T) *rest.Config {
	c, err := rest.InClusterConfig()
	require.NoError(t, err)

	return c
}

func ArangoClient(t *testing.T) arangoClient.Interface {
	c, err := arangoClient.NewForConfig(KubernetesConfig(t))
	require.NoError(t, err)
	return c
}

func KubernetesClient(t *testing.T) kubernetes.Interface {
	c, err := kubernetes.NewForConfig(KubernetesConfig(t))
	require.NoError(t, err)
	return c
}

func Namespace(t *testing.T) string {
	ns := os.Getenv(NAMESPACE_ENV)
	require.NotEqual(t, "", ns, "Namespace cannot be empty")
	return ns
}

func Image(t *testing.T) string {
	image := os.Getenv(IMAGE_ENV)

	require.NotEqual(t, "", image, "Image cannot be empty")
	return image
}

func LicenseSecret(t *testing.T) string {
	_, err := KubernetesSecretClient(t).Get(LICENSE_SECRET, v1.GetOptions{})

	if err != nil {
		if errors.IsNotFound(err) {
			require.Fail(t, "Secret with license is not found!")
			return LICENSE_SECRET
		}

		require.NoError(t, err)
		return LICENSE_SECRET
	}

	return LICENSE_SECRET
}