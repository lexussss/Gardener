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

package kubernetes_test

import (
	"context"
	"path/filepath"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/version"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	"github.com/gardener/gardener/pkg/chartrenderer"
	"github.com/gardener/gardener/pkg/client/kubernetes"
	mockclient "github.com/gardener/gardener/pkg/mock/controller-runtime/client"
	. "github.com/gardener/gardener/pkg/utils/test/matchers"
)

var _ = Describe("chart applier", func() {
	const (
		cn     = "test-chart-name"
		cns    = "test-chart-namespace"
		cmName = "test-configmap-name"
	)
	var (
		cp         = filepath.Join("testdata", "render-test")
		ctrl       *gomock.Controller
		ca         kubernetes.ChartApplier
		ctx        context.Context
		c          client.Client
		expectedCM *corev1.ConfigMap
		mapper     *meta.DefaultRESTMapper
		renderer   chartrenderer.Interface
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		ctx = context.TODO()

		c = fake.NewClientBuilder().WithScheme(scheme.Scheme).Build()

		mapper = meta.NewDefaultRESTMapper([]schema.GroupVersion{corev1.SchemeGroupVersion})
		mapper.Add(corev1.SchemeGroupVersion.WithKind("ConfigMap"), meta.RESTScopeNamespace)

		expectedCM = &corev1.ConfigMap{
			TypeMeta: configMapTypeMeta,
			ObjectMeta: metav1.ObjectMeta{
				Name:            cmName,
				Namespace:       cns,
				ResourceVersion: "1",
				Labels:          map[string]string{"baz": "goo"},
			},
			Data: map[string]string{"key": "valz"},
		}

		renderer = chartrenderer.NewWithServerVersion(&version.Info{})
	})

	JustBeforeEach(func() {
		ca = kubernetes.NewChartApplier(renderer, kubernetes.NewApplier(c, mapper))
		Expect(ca).NotTo(BeNil(), "should return chart applier")
	})

	AfterEach(func() {
		ctrl.Finish()
	})

	Describe("#Apply", func() {

		It("renders the chart with default values", func() {
			Expect(ca.Apply(ctx, cp, cns, cn)).ToNot(HaveOccurred())

			actual := &corev1.ConfigMap{}
			err := c.Get(context.TODO(), client.ObjectKey{Name: cmName, Namespace: cns}, actual)

			Expect(err).NotTo(HaveOccurred())
			Expect(actual).To(DeepDerivativeEqual(expectedCM))
		})

		It("renders the chart with custom values", func() {
			const (
				newNS = "other-namespace"
			)

			existingCM := &corev1.ConfigMap{
				TypeMeta: configMapTypeMeta,
				ObjectMeta: metav1.ObjectMeta{
					Name:      cmName,
					Namespace: newNS,
				},
			}

			Expect(c.Create(ctx, existingCM), "dummy configmap creation should succeed")

			expectedCM.Namespace = newNS
			expectedCM.Data = map[string]string{"key": "new-value"}
			expectedCM.Labels = map[string]string{"baz": "new-foo"}
			expectedCM.Annotations = map[string]string{"new": "new-value"}
			expectedCM.ResourceVersion = "2"

			val := &renderTestValues{
				Service: renderTestValuesService{
					Data:   map[string]string{"key": "new-value"},
					Labels: map[string]string{"baz": "new-foo"},
				},
			}

			m := kubernetes.MergeFuncs{
				corev1.SchemeGroupVersion.WithKind("ConfigMap").GroupKind(): func(newObj, oldObj *unstructured.Unstructured) {
					newObj.SetAnnotations(map[string]string{"new": "new-value"})
				},
			}

			Expect(ca.Apply(
				ctx,
				cp,
				newNS,
				cn,
				kubernetes.Values(val),
				kubernetes.ForceNamespace,
				nil, // simulate nil entry
				m,
				nil, // simulate nil entry
			)).ToNot(HaveOccurred())

			actual := &corev1.ConfigMap{}
			err := c.Get(context.TODO(), client.ObjectKey{Name: cmName, Namespace: newNS}, actual)

			Expect(err).NotTo(HaveOccurred())
			Expect(actual).To(DeepDerivativeEqual(expectedCM))
		})
	})

	Describe("#Delete", func() {

		It("deletes the chart with default values", func() {
			existingCM := &corev1.ConfigMap{
				TypeMeta: configMapTypeMeta,
				ObjectMeta: metav1.ObjectMeta{
					Name:      cmName,
					Namespace: cmName,
				},
			}
			Expect(c.Create(ctx, existingCM), "dummy configmap creation should succeed")

			Expect(ca.Delete(ctx, cp, cns, cn)).ToNot(HaveOccurred())

			actual := &corev1.ConfigMap{}
			err := c.Get(context.TODO(), client.ObjectKey{Name: cmName, Namespace: cns}, actual)

			Expect(err).To(BeNotFoundError())
		})

		It("deletes the chart with custom values", func() {
			const (
				newNS = "other-namespace"
			)

			existingCM := &corev1.ConfigMap{
				TypeMeta: configMapTypeMeta,
				ObjectMeta: metav1.ObjectMeta{
					Name:      cmName,
					Namespace: newNS,
				},
			}

			Expect(c.Create(ctx, existingCM), "dummy configmap creation should succeed")

			val := &renderTestValues{
				Service: renderTestValuesService{
					Data:   map[string]string{"key": "new-value"},
					Labels: map[string]string{"baz": "new-foo"},
				},
			}

			Expect(ca.Delete(
				ctx,
				cp,
				newNS,
				cn,
				kubernetes.Values(val),
				nil, // simulate nil entry
				kubernetes.ForceNamespace,
				nil, // simulate nil entry
			)).ToNot(HaveOccurred())

			actual := &corev1.ConfigMap{}
			err := c.Get(context.TODO(), client.ObjectKey{Name: cmName, Namespace: cns}, actual)

			Expect(err).To(BeNotFoundError())
		})

		Context("when object is not mapped", func() {

			var (
				ctrl *gomock.Controller
				mc   *mockclient.MockClient
			)

			BeforeEach(func() {
				ctrl = gomock.NewController(GinkgoT())
				mc = mockclient.NewMockClient(ctrl)

				c = mc
				mapper = meta.NewDefaultRESTMapper([]schema.GroupVersion{})

				mc.EXPECT().Delete(gomock.Any(), gomock.Any()).AnyTimes().Return(
					&meta.NoResourceMatchError{PartialResource: schema.GroupVersionResource{}},
				)
			})

			AfterEach(func() {
				ctrl.Finish()
			})

			It("no error when IgnoreNoMatch is set", func() {
				Expect(ca.Delete(ctx, cp, cns, cn, kubernetes.TolerateErrorFunc(meta.IsNoMatchError))).ToNot(HaveOccurred())
			})

			It("error when IgnoreNoMatch is not set", func() {
				Expect(ca.Delete(ctx, cp, cns, cn)).To(HaveOccurred())
			})
		})
	})
})

type renderTestValues struct {
	Service renderTestValuesService `json:"service,omitempty"`
}

type renderTestValuesService struct {
	Labels map[string]string `json:"labels,omitempty"`
	Data   map[string]string `json:"data,omitempty"`
}
