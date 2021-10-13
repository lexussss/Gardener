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
	"k8s.io/apimachinery/pkg/util/validation/field"
)

var _ = Describe("BackupEntry validation tests", func() {
	var be *extensionsv1alpha1.BackupEntry

	BeforeEach(func() {
		be = &extensionsv1alpha1.BackupEntry{
			ObjectMeta: metav1.ObjectMeta{
				Name: "test-be",
			},
			Spec: extensionsv1alpha1.BackupEntrySpec{
				DefaultSpec: extensionsv1alpha1.DefaultSpec{
					Type: "provider",
				},
				BucketName: "bucket-name",
				Region:     "region",
				SecretRef: corev1.SecretReference{
					Name: "test",
				},
			},
		}
	})

	Describe("#ValidBackupEntry", func() {
		It("should forbid empty BackupEntry resources", func() {
			errorList := ValidateBackupEntry(&extensionsv1alpha1.BackupEntry{})

			Expect(errorList).To(ConsistOf(PointTo(MatchFields(IgnoreExtras, Fields{
				"Type":  Equal(field.ErrorTypeRequired),
				"Field": Equal("metadata.name"),
			})), PointTo(MatchFields(IgnoreExtras, Fields{
				"Type":  Equal(field.ErrorTypeRequired),
				"Field": Equal("spec.type"),
			})), PointTo(MatchFields(IgnoreExtras, Fields{
				"Type":  Equal(field.ErrorTypeRequired),
				"Field": Equal("spec.region"),
			})), PointTo(MatchFields(IgnoreExtras, Fields{
				"Type":  Equal(field.ErrorTypeRequired),
				"Field": Equal("spec.bucketName"),
			})), PointTo(MatchFields(IgnoreExtras, Fields{
				"Type":  Equal(field.ErrorTypeRequired),
				"Field": Equal("spec.secretRef.name"),
			}))))
		})

		It("should allow valid be resources", func() {
			errorList := ValidateBackupEntry(be)

			Expect(errorList).To(BeEmpty())
		})
	})

	Describe("#ValidBackupEntryUpdate", func() {
		It("should prevent updating anything if deletion time stamp is set", func() {
			now := metav1.Now()
			be.DeletionTimestamp = &now

			newBackupEntry := prepareBackupEntryForUpdate(be)
			newBackupEntry.DeletionTimestamp = &now
			newBackupEntry.Spec.SecretRef.Name = "changed-secretref-name"

			errorList := ValidateBackupEntryUpdate(newBackupEntry, be)

			Expect(errorList).To(ConsistOf(PointTo(MatchFields(IgnoreExtras, Fields{
				"Type":  Equal(field.ErrorTypeInvalid),
				"Field": Equal("spec"),
			}))))
		})

		It("should prevent updating the type, region or bucketName", func() {
			newBackupEntry := prepareBackupEntryForUpdate(be)
			newBackupEntry.Spec.Type = "changed-type"
			newBackupEntry.Spec.Region = "changed-region"
			newBackupEntry.Spec.BucketName = "changed-bucket-name"

			errorList := ValidateBackupEntryUpdate(newBackupEntry, be)

			Expect(errorList).To(ConsistOf(PointTo(MatchFields(IgnoreExtras, Fields{
				"Type":  Equal(field.ErrorTypeInvalid),
				"Field": Equal("spec.type"),
			})), PointTo(MatchFields(IgnoreExtras, Fields{
				"Type":  Equal(field.ErrorTypeInvalid),
				"Field": Equal("spec.region"),
			})), PointTo(MatchFields(IgnoreExtras, Fields{
				"Type":  Equal(field.ErrorTypeInvalid),
				"Field": Equal("spec.bucketName"),
			}))))
		})

		It("should allow updating the name of the referenced secret", func() {
			newBackupEntry := prepareBackupEntryForUpdate(be)
			newBackupEntry.Spec.SecretRef.Name = "changed-secretref-name"

			errorList := ValidateBackupEntryUpdate(newBackupEntry, be)

			Expect(errorList).To(BeEmpty())
		})
	})
})

func prepareBackupEntryForUpdate(obj *extensionsv1alpha1.BackupEntry) *extensionsv1alpha1.BackupEntry {
	newObj := obj.DeepCopy()
	newObj.ResourceVersion = "1"
	return newObj
}
