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

package shoothibernationwakeup_test

import (
	"context"
	"time"

	"github.com/gardener/gardener/test/framework"

	. "github.com/onsi/ginkgo"
)

func init() {
	framework.RegisterShootFrameworkFlags()
}

var _ = Describe("Shoot hibernation wake-up testing", func() {
	f := framework.NewShootFramework(nil)

	framework.CIt("should wake up shoot", func(ctx context.Context) {
		hibernation := f.Shoot.Spec.Hibernation
		if hibernation == nil || hibernation.Enabled == nil || !*hibernation.Enabled {
			Fail("shoot is already woken up")
		}

		err := f.WakeUpShoot(ctx)
		framework.ExpectNoError(err)
	}, 30*time.Minute)
})
