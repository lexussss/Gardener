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

package flow

import (
	"context"
	"sync/atomic"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"go.uber.org/goleak"
	"k8s.io/apimachinery/pkg/util/clock"
)

var _ = Describe("ProgressReporterDelaying", func() {
	It("should behave correctly", func() {
		defer goleak.VerifyNone(GinkgoT(), goleak.IgnoreCurrent())

		var (
			ctx           = context.TODO()
			fakeClock     = clock.NewFakeClock(time.Now())
			period        = 50 * time.Second
			reportedStats atomic.Value
			reporterFn    = func(_ context.Context, stats *Stats) { reportedStats.Store(stats) }
			p             = NewDelayingProgressReporter(fakeClock, reporterFn, period)
		)

		Expect(p.Start(ctx)).To(Succeed())
		Expect(reportedStats.Load()).To(BeNil())

		stats1 := &Stats{FlowName: "1"}
		p.Report(ctx, stats1)
		Expect(reportedStats.Load()).To(Equal(stats1))

		stats2 := &Stats{FlowName: "2"}
		p.Report(ctx, stats2)
		Consistently(reportedStats.Load).Should(Equal(stats1))
		fakeClock.Step(period)
		Eventually(reportedStats.Load).Should(Equal(stats2))

		stats3 := &Stats{FlowName: "3"}
		p.Report(ctx, stats3)
		Consistently(reportedStats.Load).Should(Equal(stats2))
		fakeClock.Step(period)
		Eventually(reportedStats.Load).Should(Equal(stats3))

		stats4 := &Stats{FlowName: "4"}
		p.Report(ctx, stats4)
		stats5 := &Stats{FlowName: "5"}
		p.Report(ctx, stats5)
		Consistently(reportedStats.Load).Should(Equal(stats3))
		fakeClock.Step(period)
		Eventually(reportedStats.Load).Should(Equal(stats5))

		stats6 := &Stats{FlowName: "6"}
		p.Report(ctx, stats6)
		p.Stop()
		Expect(reportedStats.Load()).To(Equal(stats6))
	})
})
