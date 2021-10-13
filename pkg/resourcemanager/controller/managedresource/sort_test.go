// Copyright (c) 2020 SAP SE or an SAP affiliate company. All rights reserved. This file is licensed under the Apache Software License, v. 2 except as noted otherwise in the LICENSE file
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

package managedresource

import (
	resourcesv1alpha1 "github.com/gardener/gardener/pkg/apis/resources/v1alpha1"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
)

var _ = Describe("Reference sorter", func() {
	var refs, refsBase []resourcesv1alpha1.ObjectReference

	BeforeEach(func() {
		refsBase = []resourcesv1alpha1.ObjectReference{
			{
				ObjectReference: corev1.ObjectReference{
					APIVersion: "v1",
					Kind:       "ConfigMap",
					Name:       "foo",
					Namespace:  "bar",
				},
			},
			{
				ObjectReference: corev1.ObjectReference{
					APIVersion: "apps/v1",
					Kind:       "Deployment",
					Name:       "nginx",
					Namespace:  "web",
				},
			},
		}

		// copy refs for assertions, as referenceSorter is sorting in-place
		refs = append(refsBase[:0:0], refsBase...)
	})

	Describe("#sortObjectReferences", func() {
		It("should correctly sort refs", func() {
			sortObjectReferences(refs)
			Expect(refs).To(Equal(refsBase))
		})
		It("should correctly sort refs (inverted order)", func() {
			refs[0], refs[1] = refs[1], refs[0]
			sortObjectReferences(refs)
			Expect(refs).To(Equal(refsBase))
		})
	})

	Describe("#newReferenceSorter", func() {
		var sorter referenceSorter

		BeforeEach(func() {
			sorter = newReferenceSorter(refs).(referenceSorter)
		})

		It("should return the correct length", func() {
			Expect(sorter.Len()).To(BeEquivalentTo(len(refsBase)))
		})

		It("should return the correct length (nil slice)", func() {
			sorter = newReferenceSorter(nil).(referenceSorter)
			Expect(sorter.Len()).To(BeEquivalentTo(0))
		})

		It("should calculate the correct keys for refs", func() {
			Expect(refs).To(Equal([]resourcesv1alpha1.ObjectReference{
				refsBase[0],
				refsBase[1],
			}))
			Expect(sorter.keys).To(Equal([]string{
				"/ConfigMap/bar/foo",
				"apps/Deployment/web/nginx",
			}))
		})

		It("should correctly compare refs", func() {
			Expect(sorter.Less(0, 1)).To(BeTrue())
		})

		It("should correctly swap refs and keys", func() {
			sorter.Swap(0, 1)
			Expect(refs).To(Equal([]resourcesv1alpha1.ObjectReference{
				refsBase[1],
				refsBase[0],
			}))
			Expect(sorter.keys).To(Equal([]string{
				"apps/Deployment/web/nginx",
				"/ConfigMap/bar/foo",
			}))
		})
	})
})
