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

package validation_test

import (
	extensionsv1alpha1 "github.com/gardener/gardener/pkg/apis/extensions/v1alpha1"
	. "github.com/gardener/gardener/pkg/apis/extensions/validation"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/validation/field"
)

var _ = Describe("Worker validation tests", func() {
	var worker *extensionsv1alpha1.Worker

	BeforeEach(func() {
		worker = &extensionsv1alpha1.Worker{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-worker",
				Namespace: "test-namespace",
			},
			Spec: extensionsv1alpha1.WorkerSpec{
				DefaultSpec: extensionsv1alpha1.DefaultSpec{
					Type: "provider",
				},
				Region: "region",
				SecretRef: corev1.SecretReference{
					Name: "test",
				},
				InfrastructureProviderStatus: &runtime.RawExtension{},
				SSHPublicKey:                 []byte("key"),
				Pools: []extensionsv1alpha1.WorkerPool{
					{
						MachineType: "large",
						MachineImage: extensionsv1alpha1.MachineImage{
							Name:    "image1",
							Version: "version1",
						},
						Name:     "pool1",
						UserData: []byte("bootstrap data"),
					},
				},
			},
		}
	})

	Describe("#ValidWorker", func() {
		It("should forbid empty Worker resources", func() {
			errorList := ValidateWorker(&extensionsv1alpha1.Worker{})

			Expect(errorList).To(ConsistOf(PointTo(MatchFields(IgnoreExtras, Fields{
				"Type":  Equal(field.ErrorTypeRequired),
				"Field": Equal("metadata.name"),
			})), PointTo(MatchFields(IgnoreExtras, Fields{
				"Type":  Equal(field.ErrorTypeRequired),
				"Field": Equal("metadata.namespace"),
			})), PointTo(MatchFields(IgnoreExtras, Fields{
				"Type":  Equal(field.ErrorTypeRequired),
				"Field": Equal("spec.type"),
			})), PointTo(MatchFields(IgnoreExtras, Fields{
				"Type":  Equal(field.ErrorTypeRequired),
				"Field": Equal("spec.region"),
			})), PointTo(MatchFields(IgnoreExtras, Fields{
				"Type":  Equal(field.ErrorTypeRequired),
				"Field": Equal("spec.secretRef.name"),
			}))))
		})

		It("should forbid Worker resources with invalid pools", func() {
			workerCopy := worker.DeepCopy()

			workerCopy.Spec.Pools[0] = extensionsv1alpha1.WorkerPool{}

			errorList := ValidateWorker(workerCopy)

			Expect(errorList).To(ConsistOf(PointTo(MatchFields(IgnoreExtras, Fields{
				"Type":  Equal(field.ErrorTypeRequired),
				"Field": Equal("spec.pools[0].machineType"),
			})), PointTo(MatchFields(IgnoreExtras, Fields{
				"Type":  Equal(field.ErrorTypeRequired),
				"Field": Equal("spec.pools[0].machineImage.name"),
			})), PointTo(MatchFields(IgnoreExtras, Fields{
				"Type":  Equal(field.ErrorTypeRequired),
				"Field": Equal("spec.pools[0].machineImage.version"),
			})), PointTo(MatchFields(IgnoreExtras, Fields{
				"Type":  Equal(field.ErrorTypeRequired),
				"Field": Equal("spec.pools[0].name"),
			})), PointTo(MatchFields(IgnoreExtras, Fields{
				"Type":  Equal(field.ErrorTypeRequired),
				"Field": Equal("spec.pools[0].userData"),
			}))))
		})

		It("should allow valid worker resources", func() {
			errorList := ValidateWorker(worker)

			Expect(errorList).To(BeEmpty())
		})
	})

	Describe("#ValidWorkerUpdate", func() {
		It("should prevent updating anything if deletion time stamp is set", func() {
			now := metav1.Now()
			worker.DeletionTimestamp = &now

			newWorker := prepareWorkerForUpdate(worker)
			newWorker.DeletionTimestamp = &now
			newWorker.Spec.SecretRef.Name = "changed-secretref-name"

			errorList := ValidateWorkerUpdate(newWorker, worker)

			Expect(errorList).To(ConsistOf(PointTo(MatchFields(IgnoreExtras, Fields{
				"Type":  Equal(field.ErrorTypeInvalid),
				"Field": Equal("spec"),
			}))))
		})

		It("should prevent updating the type and region", func() {
			newWorker := prepareWorkerForUpdate(worker)
			newWorker.Spec.Type = "changed-type"
			newWorker.Spec.Region = "changed-region"

			errorList := ValidateWorkerUpdate(newWorker, worker)

			Expect(errorList).To(ConsistOf(PointTo(MatchFields(IgnoreExtras, Fields{
				"Type":  Equal(field.ErrorTypeInvalid),
				"Field": Equal("spec.type"),
			})), PointTo(MatchFields(IgnoreExtras, Fields{
				"Type":  Equal(field.ErrorTypeInvalid),
				"Field": Equal("spec.region"),
			}))))
		})

		It("should allow updating the name of the referenced secret, the infrastructure provider status, the ssh public key, or the worker pools", func() {
			newWorker := prepareWorkerForUpdate(worker)
			newWorker.Spec.SecretRef.Name = "changed-secretref-name"
			newWorker.Spec.InfrastructureProviderStatus = nil
			newWorker.Spec.SSHPublicKey = []byte("other-key")
			newWorker.Spec.Pools = []extensionsv1alpha1.WorkerPool{
				{
					MachineType: "ultra-large",
					MachineImage: extensionsv1alpha1.MachineImage{
						Name:    "image2",
						Version: "version2",
					},
					Name:     "pool2",
					UserData: []byte("new bootstrap data"),
				},
			}

			errorList := ValidateWorkerUpdate(newWorker, worker)

			Expect(errorList).To(BeEmpty())
		})
	})
})

func prepareWorkerForUpdate(obj *extensionsv1alpha1.Worker) *extensionsv1alpha1.Worker {
	newObj := obj.DeepCopy()
	newObj.ResourceVersion = "1"
	return newObj
}
