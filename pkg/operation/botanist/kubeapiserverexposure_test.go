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

package botanist

import (
	"context"

	gardencorev1beta1 "github.com/gardener/gardener/pkg/apis/core/v1beta1"
	fakeclientset "github.com/gardener/gardener/pkg/client/kubernetes/fake"
	gardenletfeatures "github.com/gardener/gardener/pkg/gardenlet/features"
	"github.com/gardener/gardener/pkg/logger"
	"github.com/gardener/gardener/pkg/operation"
	"github.com/gardener/gardener/pkg/operation/botanist/component"
	"github.com/gardener/gardener/pkg/operation/garden"
	"github.com/gardener/gardener/pkg/operation/shoot"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	"github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/utils/pointer"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

var _ = Describe("KubeAPIServerExposure", func() {
	var (
		ctrl   *gomock.Controller
		scheme *runtime.Scheme
		client client.Client

		botanist *Botanist

		ctx       = context.TODO()
		namespace = "shoot--foo--bar"
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())

		scheme = runtime.NewScheme()
		Expect(corev1.AddToScheme(scheme)).NotTo(HaveOccurred())
		client = fake.NewClientBuilder().WithScheme(scheme).Build()

		fakeClientSet := fakeclientset.NewClientSetBuilder().
			WithAPIReader(client).
			Build()

		botanist = &Botanist{
			Operation: &operation.Operation{
				K8sSeedClient: fakeClientSet,
				Shoot: &shoot.Shoot{
					SeedNamespace: namespace,
				},
				Garden: &garden.Garden{},
				Logger: logrus.NewEntry(logger.NewNopLogger()),
			},
		}
		botanist.Shoot.SetInfo(&gardencorev1beta1.Shoot{})
	})

	AfterEach(func() {
		ctrl.Finish()
	})

	Describe("#SNIPhase", func() {
		var svc *corev1.Service

		BeforeEach(func() {
			gardenletfeatures.RegisterFeatureGates()

			svc = &corev1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "kube-apiserver",
					Namespace: namespace,
				},
			}
		})

		Context("sni enabled", func() {
			BeforeEach(func() {
				Expect(gardenletfeatures.FeatureGate.Set("APIServerSNI=true")).ToNot(HaveOccurred())
				botanist.Garden.InternalDomain = &garden.Domain{Provider: "some-provider"}
				botanist.Shoot.GetInfo().Spec.DNS = &gardencorev1beta1.DNS{Domain: pointer.String("foo")}
				botanist.Shoot.ExternalClusterDomain = pointer.String("baz")
				botanist.Shoot.ExternalDomain = &garden.Domain{Provider: "valid-provider"}
			})

			It("returns Enabled for not existing services", func() {
				phase, err := botanist.SNIPhase(ctx)
				Expect(err).NotTo(HaveOccurred())
				Expect(phase).To(Equal(component.PhaseEnabled))
			})

			It("returns Enabling for service of type LoadBalancer", func() {
				svc.Spec.Type = corev1.ServiceTypeLoadBalancer
				Expect(client.Create(ctx, svc)).NotTo(HaveOccurred())

				phase, err := botanist.SNIPhase(ctx)
				Expect(err).NotTo(HaveOccurred())
				Expect(phase).To(Equal(component.PhaseEnabling))
			})

			It("returns Enabled for service of type ClusterIP", func() {
				svc.Spec.Type = corev1.ServiceTypeClusterIP
				Expect(client.Create(ctx, svc)).NotTo(HaveOccurred())

				phase, err := botanist.SNIPhase(ctx)
				Expect(err).NotTo(HaveOccurred())
				Expect(phase).To(Equal(component.PhaseEnabled))
			})

			DescribeTable(
				"return Enabled for service of type",
				func(svcType corev1.ServiceType) {
					svc.Spec.Type = svcType
					Expect(client.Create(ctx, svc)).NotTo(HaveOccurred())

					phase, err := botanist.SNIPhase(ctx)
					Expect(err).NotTo(HaveOccurred())
					Expect(phase).To(Equal(component.PhaseEnabled))
				},

				Entry("ExternalName", corev1.ServiceTypeExternalName),
				Entry("NodePort", corev1.ServiceTypeNodePort),
			)
		})

		Context("sni disabled", func() {
			BeforeEach(func() {
				Expect(gardenletfeatures.FeatureGate.Set("APIServerSNI=false")).ToNot(HaveOccurred())
			})

			It("returns Disabled for not existing services", func() {
				phase, err := botanist.SNIPhase(ctx)
				Expect(err).NotTo(HaveOccurred())
				Expect(phase).To(Equal(component.PhaseDisabled))
			})

			It("returns Disabling for service of type ClusterIP", func() {
				svc.Spec.Type = corev1.ServiceTypeClusterIP
				Expect(client.Create(ctx, svc)).NotTo(HaveOccurred())

				phase, err := botanist.SNIPhase(ctx)
				Expect(err).NotTo(HaveOccurred())
				Expect(phase).To(Equal(component.PhaseDisabling))
			})

			DescribeTable(
				"return Disabled for service of type",
				func(svcType corev1.ServiceType) {
					svc.Spec.Type = svcType
					Expect(client.Create(ctx, svc)).NotTo(HaveOccurred())

					phase, err := botanist.SNIPhase(ctx)
					Expect(err).NotTo(HaveOccurred())
					Expect(phase).To(Equal(component.PhaseDisabled))
				},

				Entry("ExternalName", corev1.ServiceTypeExternalName),
				Entry("LoadBalancer", corev1.ServiceTypeLoadBalancer),
				Entry("NodePort", corev1.ServiceTypeNodePort),
			)
		})
	})
})
