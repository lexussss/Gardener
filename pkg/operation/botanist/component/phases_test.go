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

package component_test

import (
	"github.com/gardener/gardener/pkg/operation/botanist/component"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
)

var _ = Describe("Phases' Done", func() {
	DescribeTable("correct phase for", func(old, expected component.Phase) {
		Expect(old.Done()).To(Equal(expected))
	},
		Entry("unknown is not changed", component.PhaseUnknown, component.PhaseUnknown),
		Entry("not phase a is always unknown", component.Phase(1234), component.PhaseUnknown),
		Entry("enabled is not changed", component.PhaseEnabled, component.PhaseEnabled),
		Entry("disabled is not changed", component.PhaseDisabled, component.PhaseDisabled),
		Entry("enabling is changed to enabled", component.PhaseEnabling, component.PhaseEnabled),
		Entry("disabling is changed to disabled", component.PhaseDisabling, component.PhaseDisabled),
	)
})
