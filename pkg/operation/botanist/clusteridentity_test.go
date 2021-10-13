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
	"fmt"

	gardencorev1beta1 "github.com/gardener/gardener/pkg/apis/core/v1beta1"
	extensionsv1alpha1 "github.com/gardener/gardener/pkg/apis/extensions/v1alpha1"
	"github.com/gardener/gardener/pkg/client/kubernetes"
	fakekubernetes "github.com/gardener/gardener/pkg/client/kubernetes/fake"
	"github.com/gardener/gardener/pkg/operation"
	. "github.com/gardener/gardener/pkg/operation/botanist"
	mockclusteridentity "github.com/gardener/gardener/pkg/operation/botanist/component/clusteridentity/mock"
	shootpkg "github.com/gardener/gardener/pkg/operation/shoot"
	kutil "github.com/gardener/gardener/pkg/utils/kubernetes"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/utils/pointer"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

var _ = Describe("ClusterIdentity", func() {
	const (
		shootName             = "shootName"
		shootNamespace        = "shootNamespace"
		shootSeedNamespace    = "shootSeedNamespace"
		shootUID              = "shootUID"
		gardenClusterIdentity = "garden-cluster-identity"
	)

	var (
		ctrl            *gomock.Controller
		clusterIdentity *mockclusteridentity.MockInterface

		ctx     = context.TODO()
		fakeErr = fmt.Errorf("fake")

		gardenInterface kubernetes.Interface
		seedInterface   kubernetes.Interface

		gardenClient client.Client
		seedClient   client.Client

		shoot *gardencorev1beta1.Shoot

		botanist *Botanist

		expectedShootClusterIdentity = fmt.Sprintf("%s-%s-%s", shootSeedNamespace, shootUID, gardenClusterIdentity)
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		clusterIdentity = mockclusteridentity.NewMockInterface(ctrl)

		s := runtime.NewScheme()
		Expect(corev1.AddToScheme(s)).To(Succeed())
		Expect(extensionsv1alpha1.AddToScheme(s)).NotTo(HaveOccurred())
		Expect(gardencorev1beta1.AddToScheme(s))

		shoot = &gardencorev1beta1.Shoot{
			ObjectMeta: metav1.ObjectMeta{
				Name:      shootName,
				Namespace: shootNamespace,
			},
			Status: gardencorev1beta1.ShootStatus{
				UID: shootUID,
			},
		}

		cluster := &extensionsv1alpha1.Cluster{
			ObjectMeta: metav1.ObjectMeta{
				Name: shootSeedNamespace,
			},
			Spec: extensionsv1alpha1.ClusterSpec{
				Shoot: runtime.RawExtension{Object: shoot},
			},
		}

		gardenClient = fake.NewClientBuilder().WithScheme(s).WithRuntimeObjects(shoot).Build()
		seedClient = fake.NewClientBuilder().WithScheme(s).WithRuntimeObjects(cluster).Build()

		gardenInterface = fakekubernetes.NewClientSetBuilder().WithClient(gardenClient).Build()
		seedInterface = fakekubernetes.NewClientSetBuilder().WithClient(seedClient).Build()

		botanist = &Botanist{
			Operation: &operation.Operation{
				K8sGardenClient: gardenInterface,
				K8sSeedClient:   seedInterface,
				Shoot: &shootpkg.Shoot{
					SeedNamespace: shootSeedNamespace,
					Components: &shootpkg.Components{
						SystemComponents: &shootpkg.SystemComponents{
							ClusterIdentity: clusterIdentity,
						},
					},
				},
				GardenClusterIdentity: gardenClusterIdentity,
			},
		}
		botanist.Shoot.SetInfo(shoot)
	})

	AfterEach(func() {
		ctrl.Finish()
	})

	DescribeTable("#EnsureShootClusterIdentity",
		func(mutator func()) {
			mutator()

			Expect(botanist.EnsureShootClusterIdentity(ctx)).NotTo(HaveOccurred())

			Expect(gardenClient.Get(ctx, kutil.Key(shootNamespace, shootName), shoot)).To(Succeed())
			Expect(shoot.Status.ClusterIdentity).NotTo(BeNil())
			Expect(*shoot.Status.ClusterIdentity).To(Equal(expectedShootClusterIdentity))
		},

		Entry("cluster identity is nil", func() {
			shoot.Status.ClusterIdentity = nil
		}),
		Entry("cluster idenitty already exists", func() {
			shoot.Status.ClusterIdentity = pointer.String(expectedShootClusterIdentity)
		}),
	)

	Describe("#DeployClusterIdentity", func() {
		BeforeEach(func() {
			botanist.Shoot.GetInfo().Status.ClusterIdentity = &expectedShootClusterIdentity
			clusterIdentity.EXPECT().SetIdentity(expectedShootClusterIdentity)
		})

		It("should deploy successfully", func() {
			clusterIdentity.EXPECT().Deploy(ctx)
			Expect(botanist.DeployClusterIdentity(ctx)).To(Succeed())
		})

		It("should return the error during deployment", func() {
			clusterIdentity.EXPECT().Deploy(ctx).Return(fakeErr)
			Expect(botanist.DeployClusterIdentity(ctx)).To(MatchError(fakeErr))
		})
	})
})
