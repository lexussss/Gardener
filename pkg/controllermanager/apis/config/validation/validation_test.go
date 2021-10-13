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

package validation_test

import (
	"github.com/gardener/gardener/pkg/controllermanager/apis/config"
	. "github.com/gardener/gardener/pkg/controllermanager/apis/config/validation"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/validation/field"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"
)

var _ = Describe("#ValidateControllerManagerConfiguration", func() {
	var conf *config.ControllerManagerConfiguration

	BeforeEach(func() {
		conf = &config.ControllerManagerConfiguration{
			Controllers: config.ControllerManagerControllerConfiguration{},
		}
	})

	Context("ProjectControllerConfiguration", func() {
		Context("ProjectQuotaConfiguration", func() {
			BeforeEach(func() {
				conf.Controllers.Project = &config.ProjectControllerConfiguration{}
			})

			It("should pass because no quota configuration is specified", func() {
				errorList := ValidateControllerManagerConfiguration(conf)
				Expect(errorList).To(BeEmpty())
			})
			It("should pass because quota configuration has correct label selector", func() {
				conf.Controllers.Project.Quotas = []config.QuotaConfiguration{
					{
						ProjectSelector: &metav1.LabelSelector{
							MatchExpressions: []metav1.LabelSelectorRequirement{
								{Key: "role", Operator: "In", Values: []string{"user"}},
							},
						},
						Config: &corev1.ResourceQuota{},
					},
				}
				errorList := ValidateControllerManagerConfiguration(conf)
				Expect(errorList).To(BeEmpty())
			})
			It("should fail because quota config is not specified", func() {
				conf.Controllers.Project.Quotas = []config.QuotaConfiguration{
					{
						ProjectSelector: &metav1.LabelSelector{
							MatchExpressions: []metav1.LabelSelectorRequirement{
								{Key: "role", Operator: "In", Values: []string{"user"}},
							},
						},
						Config: nil,
					},
				}
				errorList := ValidateControllerManagerConfiguration(conf)
				Expect(errorList).To(ConsistOf(
					PointTo(MatchFields(IgnoreExtras, Fields{
						"Type":  Equal(field.ErrorTypeRequired),
						"Field": Equal("controllers.project.quotas[0].config"),
					})),
				))
			})
			It("should fail because quota configuration contains invalid label selector", func() {
				conf.Controllers.Project.Quotas = []config.QuotaConfiguration{
					{
						ProjectSelector: &metav1.LabelSelector{
							MatchExpressions: []metav1.LabelSelectorRequirement{
								{Key: "role", Operator: "In", Values: []string{"user"}},
							},
						},
						Config: &corev1.ResourceQuota{},
					},
					{
						ProjectSelector: &metav1.LabelSelector{
							MatchExpressions: []metav1.LabelSelectorRequirement{
								{},
							},
						},
						Config: &corev1.ResourceQuota{},
					},
				}
				errorList := ValidateControllerManagerConfiguration(conf)
				Expect(errorList).To(ConsistOf(
					PointTo(MatchFields(IgnoreExtras, Fields{
						"Type":  Equal(field.ErrorTypeInvalid),
						"Field": Equal("controllers.project.quotas[1].projectSelector.matchExpressions[0].operator"),
					})),
					PointTo(MatchFields(IgnoreExtras, Fields{
						"Type":  Equal(field.ErrorTypeInvalid),
						"Field": Equal("controllers.project.quotas[1].projectSelector.matchExpressions[0].key"),
					})),
					PointTo(MatchFields(IgnoreExtras, Fields{
						"Type":  Equal(field.ErrorTypeInvalid),
						"Field": Equal("controllers.project.quotas[1].projectSelector.matchExpressions[0].key"),
					})),
				))
			})
		})
	})
})
