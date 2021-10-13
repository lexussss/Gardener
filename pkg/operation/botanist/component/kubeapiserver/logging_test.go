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

package kubeapiserver_test

import (
	. "github.com/gardener/gardener/pkg/operation/botanist/component/kubeapiserver"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Logging", func() {
	Describe("#CentralLoggingConfiguration", func() {
		It("should return the expected logging parser and filter", func() {
			loggingConfig, err := CentralLoggingConfiguration()

			Expect(err).NotTo(HaveOccurred())
			Expect(loggingConfig.Parsers).To(Equal(`[PARSER]
    Name        kubeAPIServerParser
    Format      regex
    Regex       ^(?<severity>\w)(?<time>\d{4} [^\s]*)\s+(?<pid>\d+)\s+(?<source>[^ \]]+)\] (?<log>.*)$
    Time_Key    time
    Time_Format %m%d %H:%M:%S.%L

[PARSER]
    Name        vpnSeedParser
    Format      regex
    Regex       ^(?<time>[^0-9]*\d{1,2}\s+[^\s]+\s+\d{4})\s+(?<log>.*)
    Time_Key    time
    Time_Format %a %b%t%d %H:%M:%S %Y
`))

			Expect(loggingConfig.Filters).To(Equal(`[FILTER]
    Name                parser
    Match               kubernetes.*kube-apiserver*kube-apiserver*
    Key_Name            log
    Parser              kubeAPIServerParser
    Reserve_Data        True

[FILTER]
    Name                parser
    Match               kubernetes.*kube-apiserver*vpn-seed*
    Key_Name            log
    Parser              vpnSeedParser
    Reserve_Data        True
`))
			Expect(loggingConfig.PodPrefix).To(BeEmpty())
			Expect(loggingConfig.UserExposed).To(BeFalse())
		})
	})
})
