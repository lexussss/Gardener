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

package botanist_test

import (
	"context"
	"errors"
	"fmt"

	"github.com/gardener/gardener/pkg/client/kubernetes"
	"github.com/gardener/gardener/pkg/client/kubernetes/fake"
	"github.com/gardener/gardener/pkg/logger"
	mockclient "github.com/gardener/gardener/pkg/mock/controller-runtime/client"
	"github.com/gardener/gardener/pkg/operation/botanist"
	"github.com/gardener/gardener/pkg/operation/common"
	"github.com/gardener/gardener/pkg/utils/test"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ = Describe("Tunnel", func() {
	Describe("CheckTunnelConnection", func() {
		var (
			ctrl *gomock.Controller

			ctx        context.Context
			cl         *mockclient.MockClient
			clientset  *fake.ClientSet
			logEntry   logrus.FieldLogger
			tunnelName string
			tunnelPod  corev1.Pod
		)

		BeforeEach(func() {
			ctrl = gomock.NewController(GinkgoT())

			ctx = context.Background()
			cl = mockclient.NewMockClient(ctrl)
			logEntry = logger.NewNopLogger()
			tunnelName = common.VPNTunnel
			tunnelPod = corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: metav1.NamespaceSystem,
					Name:      tunnelName,
				},
			}
		})

		AfterEach(func() {
			ctrl.Finish()
		})

		JustBeforeEach(func() {
			clientset = fake.NewClientSetBuilder().
				WithClient(cl).
				Build()
		})

		Context("unavailable tunnel pod", func() {
			It("should fail because pod does not exist", func() {
				cl.EXPECT().List(ctx, gomock.AssignableToTypeOf(&corev1.PodList{}), client.InNamespace(metav1.NamespaceSystem), client.MatchingLabels{"app": tunnelName}).Return(nil)
				done, err := botanist.CheckTunnelConnection(context.Background(), clientset, logEntry, tunnelName)
				Expect(done).To(BeFalse())
				Expect(err).To(HaveOccurred())
			})
			It("should fail because pod list returns error", func() {
				cl.EXPECT().List(ctx, gomock.AssignableToTypeOf(&corev1.PodList{}), client.InNamespace(metav1.NamespaceSystem), client.MatchingLabels{"app": tunnelName}).Return(errors.New("foo"))
				done, err := botanist.CheckTunnelConnection(context.Background(), clientset, logEntry, tunnelName)
				Expect(done).To(BeTrue())
				Expect(err).To(HaveOccurred())
			})
			It("should fail because pod is not running", func() {
				cl.EXPECT().List(ctx, gomock.AssignableToTypeOf(&corev1.PodList{}), client.InNamespace(metav1.NamespaceSystem), client.MatchingLabels{"app": tunnelName}).DoAndReturn(
					func(_ context.Context, podList *corev1.PodList, _ ...client.ListOption) error {
						podList.Items = append(podList.Items, tunnelPod)
						return nil
					})
				done, err := botanist.CheckTunnelConnection(context.Background(), clientset, logEntry, tunnelName)
				Expect(done).To(BeFalse())
				Expect(err).To(HaveOccurred())
			})
		})
		Context("available tunnel pod", func() {
			BeforeEach(func() {
				tunnelPod.Status = corev1.PodStatus{
					Phase: corev1.PodRunning,
				}
				cl.EXPECT().List(ctx, gomock.AssignableToTypeOf(&corev1.PodList{}), client.InNamespace(metav1.NamespaceSystem), client.MatchingLabels{"app": tunnelName}).DoAndReturn(
					func(_ context.Context, podList *corev1.PodList, _ ...client.ListOption) error {
						podList.Items = append(podList.Items, tunnelPod)
						return nil
					})
			})
			Context("established connection", func() {
				It("should succeed because pod is running and connection successful", func() {
					fw := fake.PortForwarder{
						ReadyChan: make(chan struct{}, 1),
					}

					defer test.WithVar(&botanist.SetupPortForwarder, func(context.Context, *rest.Config, string, string, int, int) (kubernetes.PortForwarder, error) {
						return fw, nil
					})()
					close(fw.ReadyChan)

					done, err := botanist.CheckTunnelConnection(context.Background(), clientset, logEntry, tunnelName)
					Expect(done).To(BeTrue())
					Expect(err).ToNot(HaveOccurred())
				})
			})
			Context("broken connection", func() {
				It("should fail because pod is running but connection is not established", func() {
					defer test.WithVar(&botanist.SetupPortForwarder, func(context.Context, *rest.Config, string, string, int, int) (kubernetes.PortForwarder, error) {
						return nil, fmt.Errorf("foo")
					})()

					done, err := botanist.CheckTunnelConnection(context.Background(), clientset, logEntry, tunnelName)
					Expect(done).To(BeFalse())
					Expect(err).To(HaveOccurred())
				})
			})
		})
	})
})
