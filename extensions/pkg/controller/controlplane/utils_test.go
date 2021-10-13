// Copyright (c) 2019 SAP SE or an SAP affiliate company. All rights reserved. This file is licensed under the Apache Software License, v. 2 except as noted otherwise in the LICENSE file
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package controlplane

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("Utils", func() {
	Describe("#MergeSecretMaps", func() {
		var (
			test0 = getSecret("test0", "default", nil)
			test1 = getSecret("test1", "default", nil)
			a     = map[string]*corev1.Secret{
				"test0": test0,
			}
			b = map[string]*corev1.Secret{
				"test1": test1,
			}
		)

		It("should return an empty map if both given maps are empty", func() {
			Expect(MergeSecretMaps(nil, nil)).To(BeEmpty())
		})
		It("should return the other map of one of the given maps is empty", func() {
			Expect(MergeSecretMaps(a, nil)).To(Equal(a))
		})
		It("should properly merge the given non-empty maps", func() {
			result := MergeSecretMaps(a, b)
			Expect(result).To(HaveKeyWithValue("test0", test0))
			Expect(result).To(HaveKeyWithValue("test1", test1))
		})
	})

	Describe("#ComputeCheckums", func() {
		var (
			secrets = map[string]*corev1.Secret{
				"test-secret": getSecret("test-secret", "default", map[string][]byte{"foo": []byte("bar")}),
			}
			cms = map[string]*corev1.ConfigMap{
				"test-config": getConfigMap("test-config", "default", map[string]string{"abc": "xyz"}),
			}
		)
		It("should compute all checksums for the given secrets and configmpas", func() {
			checksums := ComputeChecksums(secrets, cms)
			Expect(checksums).To(HaveKeyWithValue("test-secret", "8bafb35ff1ac60275d62e1cbd495aceb511fb354f74a20f7d06ecb48b3a68432"))
			Expect(checksums).To(HaveKeyWithValue("test-config", "08a7bc7fe8f59b055f173145e211760a83f02cf89635cef26ebb351378635606"))
		})
	})
})

func getSecret(name, namespace string, data map[string][]byte) *corev1.Secret {
	return &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Data: data,
	}
}

func getConfigMap(name, namespace string, data map[string]string) *corev1.ConfigMap {
	return &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Data: data,
	}
}
