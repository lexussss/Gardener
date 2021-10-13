// Copyright (c) 2021 SAP SE or an SAP affiliate company. All rights reserved. This file is licensed under the Apache Software License, v. 2 except as noted otherwise in the LICENSE file
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

package gardener_test

import (
	. "github.com/gardener/gardener/pkg/utils/gardener"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/types"
)

var _ = Describe("BackupEntry", func() {
	var (
		shootTechnicalID = "seednamespace"
		shootUID         = types.UID("1234")
		backupEntryName  = shootTechnicalID + "--" + string(shootUID)
	)

	Describe("#GenerateBackupEntryName", func() {
		It("should compute the correct name", func() {
			result, err := GenerateBackupEntryName(shootTechnicalID, shootUID)
			Expect(err).To(Not(HaveOccurred()))
			Expect(result).To(Equal(backupEntryName))
		})

		It("should fail if the shoot technical ID is empty", func() {
			_, err := GenerateBackupEntryName("", shootUID)
			Expect(err).To(HaveOccurred())
		})

		It("should fail if the shoot UID is empty", func() {
			_, err := GenerateBackupEntryName(shootTechnicalID, "")
			Expect(err).To(HaveOccurred())
		})
	})

	Describe("#ExtractShootDetailsFromBackupEntryName", func() {
		It("should return the correct parts of the name", func() {
			technicalID, uid := ExtractShootDetailsFromBackupEntryName(backupEntryName)
			Expect(technicalID).To(Equal(shootTechnicalID))
			Expect(uid).To(Equal(shootUID))
		})
	})
})
