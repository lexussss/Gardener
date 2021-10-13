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

package etcd_test

import (
	"context"
	"time"

	druidv1alpha1 "github.com/gardener/etcd-druid/api/v1alpha1"
	hvpav1alpha1 "github.com/gardener/hvpa-controller/api/v1alpha1"
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/sirupsen/logrus"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/utils/pointer"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	gardencorev1beta1 "github.com/gardener/gardener/pkg/apis/core/v1beta1"
	v1beta1constants "github.com/gardener/gardener/pkg/apis/core/v1beta1/constants"
	"github.com/gardener/gardener/pkg/logger"
	mocktime "github.com/gardener/gardener/pkg/mock/go/time"
	"github.com/gardener/gardener/pkg/operation/botanist/component"
	. "github.com/gardener/gardener/pkg/operation/botanist/component/etcd"
	"github.com/gardener/gardener/pkg/utils/retry"
	retryfake "github.com/gardener/gardener/pkg/utils/retry/fake"
	"github.com/gardener/gardener/pkg/utils/test"
)

var _ = Describe("#Wait", func() {
	var (
		ctrl    *gomock.Controller
		c       client.Client
		log     logrus.FieldLogger
		mockNow *mocktime.MockNow
		now     time.Time

		waiter      *retryfake.Ops
		cleanupFunc func()

		ctx  = context.TODO()
		name = "etcd-" + testRole

		etcd     Interface
		expected *druidv1alpha1.Etcd
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		mockNow = mocktime.NewMockNow(ctrl)
		now = time.Now()

		s := runtime.NewScheme()
		Expect(appsv1.AddToScheme(s)).To(Succeed())
		Expect(networkingv1.AddToScheme(s)).To(Succeed())
		Expect(hvpav1alpha1.AddToScheme(s)).To(Succeed())
		Expect(druidv1alpha1.AddToScheme(s)).To(Succeed())
		c = fake.NewClientBuilder().WithScheme(s).Build()

		log = logger.NewNopLogger()

		waiter = &retryfake.Ops{MaxAttempts: 1}
		cleanupFunc = test.WithVars(
			&retry.Until, waiter.Until,
			&retry.UntilTimeout, waiter.UntilTimeout,
		)

		etcd = New(c, log, testNamespace, testRole, ClassNormal, false, "12Gi", pointer.String("abcd"))
		etcd.SetSecrets(Secrets{
			CA:     component.Secret{Name: "ca", Checksum: "abcdef"},
			Server: component.Secret{Name: "server", Checksum: "abcdef"},
			Client: component.Secret{Name: "client", Checksum: "abcdef"},
		})
		etcd.SetHVPAConfig(&HVPAConfig{
			Enabled: true,
			MaintenanceTimeWindow: gardencorev1beta1.MaintenanceTimeWindow{
				Begin: "1234",
				End:   "5678",
			},
			ScaleDownUpdateMode: pointer.String(hvpav1alpha1.UpdateModeMaintenanceWindow),
		})

		expected = &druidv1alpha1.Etcd{
			TypeMeta: metav1.TypeMeta{
				APIVersion: druidv1alpha1.GroupVersion.String(),
				Kind:       "Etcd",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      name,
				Namespace: testNamespace,
				Annotations: map[string]string{
					v1beta1constants.GardenerOperation: v1beta1constants.GardenerOperationReconcile,
					v1beta1constants.GardenerTimestamp: now.UTC().String(),
				},
			},
			Spec: druidv1alpha1.EtcdSpec{},
		}
	})

	AfterEach(func() {
		ctrl.Finish()
		cleanupFunc()
	})

	It("should return error when it's not found", func() {
		Expect(etcd.Wait(ctx)).To(MatchError(ContainSubstring("not found")))
	})

	It("should return error when it's not ready", func() {
		defer test.WithVars(
			&TimeNow, mockNow.Do,
		)()
		mockNow.EXPECT().Do().Return(now.UTC()).AnyTimes()
		expected.Status.LastError = pointer.String("some error")

		Expect(c.Create(ctx, expected)).To(Succeed(), "creating etcd succeeds")
		Expect(etcd.Wait(ctx)).To(MatchError(ContainSubstring("some error")))
	})

	It("should return error if we haven't observed the latest timestamp annotation", func() {
		defer test.WithVars(
			&TimeNow, mockNow.Do,
		)()
		mockNow.EXPECT().Do().Return(now.UTC()).AnyTimes()

		By("deploy")
		// Deploy should fill internal state with the added timestamp annotation
		Expect(etcd.Deploy(ctx)).To(Succeed())

		By("patch object")
		patch := client.MergeFrom(expected.DeepCopy())
		expected.Status.LastError = nil
		// remove operation annotation, add old timestamp annotation
		expected.ObjectMeta.Annotations = map[string]string{
			v1beta1constants.GardenerTimestamp: now.Add(-time.Millisecond).UTC().String(),
		}
		expected.Status.Ready = pointer.Bool(true)
		Expect(c.Patch(ctx, expected, patch)).To(Succeed(), "patching etcd succeeds")

		By("wait")
		Expect(etcd.Wait(ctx)).NotTo(Succeed(), "etcd indicates error")
	})

	It("should return no error when is ready", func() {
		defer test.WithVars(
			&TimeNow, mockNow.Do,
		)()
		mockNow.EXPECT().Do().Return(now.UTC()).AnyTimes()

		By("deploy")
		// Deploy should fill internal state with the added timestamp annotation
		Expect(etcd.Deploy(ctx)).To(Succeed())

		By("patch object")
		delete(expected.Annotations, v1beta1constants.GardenerTimestamp)
		patch := client.MergeFrom(expected.DeepCopy())
		expected.Status.ObservedGeneration = pointer.Int64(0)
		expected.Status.LastError = nil
		// remove operation annotation, add up-to-date timestamp annotation
		expected.ObjectMeta.Annotations = map[string]string{
			v1beta1constants.GardenerTimestamp: now.UTC().String(),
		}
		expected.Status.Ready = pointer.Bool(true)
		Expect(c.Patch(ctx, expected, patch)).To(Succeed(), "patching etcd succeeds")

		By("wait")
		Expect(etcd.Wait(ctx)).To(Succeed(), "etcd is ready")
	})
})

var _ = Describe("#CheckEtcdObject", func() {
	var (
		obj *druidv1alpha1.Etcd
	)

	BeforeEach(func() {
		obj = &druidv1alpha1.Etcd{}
	})

	It("should return error for non-dns object", func() {
		Expect(CheckEtcdObject(&corev1.ConfigMap{}))
	})

	It("should return error if reconciliation failed", func() {
		obj.Status.LastError = pointer.String("foo")
		err := CheckEtcdObject(obj)
		Expect(err).To(MatchError("foo"))
		Expect(retry.IsRetriable(err)).To(BeTrue())
	})

	It("should return error if etcd is marked for deletion", func() {
		now := metav1.Now()
		obj.SetDeletionTimestamp(&now)
		Expect(CheckEtcdObject(obj)).To(MatchError("unexpectedly has a deletion timestamp"))
	})

	It("should return error if observedGeneration is not set", func() {
		Expect(CheckEtcdObject(obj)).To(MatchError("observed generation not recorded"))
	})

	It("should return error if observedGeneration is outdated", func() {
		obj.SetGeneration(1)
		obj.Status.ObservedGeneration = pointer.Int64(0)
		Expect(CheckEtcdObject(obj)).To(MatchError("observed generation outdated (0/1)"))
	})

	It("should return error if operation annotation is not removed yet", func() {
		obj.SetGeneration(1)
		obj.Status.ObservedGeneration = pointer.Int64(1)
		metav1.SetMetaDataAnnotation(&obj.ObjectMeta, v1beta1constants.GardenerOperation, "reconcile")
		Expect(CheckEtcdObject(obj)).To(MatchError("gardener operation \"reconcile\" is not yet picked up by etcd-druid"))
	})

	It("should return error if status.ready==nil", func() {
		obj.SetGeneration(1)
		obj.Status.ObservedGeneration = pointer.Int64(1)
		Expect(CheckEtcdObject(obj)).To(MatchError("is not ready yet"))
	})

	It("should return error if status.ready==false", func() {
		obj.SetGeneration(1)
		obj.Status.ObservedGeneration = pointer.Int64(1)
		obj.Status.Ready = pointer.Bool(false)
		Expect(CheckEtcdObject(obj)).To(MatchError("is not ready yet"))
	})

	It("should not return error if object is ready", func() {
		obj.SetGeneration(1)
		obj.Status.ObservedGeneration = pointer.Int64(1)
		obj.Status.Ready = pointer.Bool(true)
		Expect(CheckEtcdObject(obj)).To(Succeed())
	})
})
