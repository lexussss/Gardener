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

package botanist_test

import (
	"context"
	"fmt"

	gardencorev1alpha1 "github.com/gardener/gardener/pkg/apis/core/v1alpha1"
	gardencorev1beta1 "github.com/gardener/gardener/pkg/apis/core/v1beta1"
	mockkubernetes "github.com/gardener/gardener/pkg/client/kubernetes/mock"
	mockclient "github.com/gardener/gardener/pkg/mock/controller-runtime/client"
	"github.com/gardener/gardener/pkg/operation"
	. "github.com/gardener/gardener/pkg/operation/botanist"
	mockinfrastructure "github.com/gardener/gardener/pkg/operation/botanist/component/extensions/infrastructure/mock"
	shootpkg "github.com/gardener/gardener/pkg/operation/shoot"
	"github.com/gardener/gardener/pkg/utils/test"
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/utils/pointer"
)

var _ = Describe("Infrastructure", func() {
	var (
		ctrl           *gomock.Controller
		infrastructure *mockinfrastructure.MockInterface
		botanist       *Botanist

		ctx          = context.TODO()
		fakeErr      = fmt.Errorf("fake")
		shootState   = &gardencorev1alpha1.ShootState{}
		sshPublicKey = []byte("key")
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		infrastructure = mockinfrastructure.NewMockInterface(ctrl)
		botanist = &Botanist{Operation: &operation.Operation{
			Shoot: &shootpkg.Shoot{
				Components: &shootpkg.Components{
					Extensions: &shootpkg.Extensions{
						Infrastructure: infrastructure,
					},
				},
			},
		}}
		botanist.SetShootState(shootState)
		botanist.Shoot.SetInfo(&gardencorev1beta1.Shoot{})
		botanist.StoreSecret("ssh-keypair", &corev1.Secret{Data: map[string][]byte{"id_rsa.pub": sshPublicKey}})
	})

	AfterEach(func() {
		ctrl.Finish()
	})

	Describe("#DeployInfrastructure", func() {
		BeforeEach(func() {
			infrastructure.EXPECT().SetSSHPublicKey(sshPublicKey)
		})

		Context("deploy", func() {
			It("should deploy successfully", func() {
				infrastructure.EXPECT().Deploy(ctx)
				Expect(botanist.DeployInfrastructure(ctx)).To(Succeed())
			})

			It("should return the error during deployment", func() {
				infrastructure.EXPECT().Deploy(ctx).Return(fakeErr)
				Expect(botanist.DeployInfrastructure(ctx)).To(MatchError(fakeErr))
			})
		})

		Context("restore", func() {
			BeforeEach(func() {
				botanist.Shoot.SetInfo(&gardencorev1beta1.Shoot{
					Status: gardencorev1beta1.ShootStatus{
						LastOperation: &gardencorev1beta1.LastOperation{
							Type: gardencorev1beta1.LastOperationTypeRestore,
						},
					},
				})
			})

			It("should restore successfully", func() {
				infrastructure.EXPECT().Restore(ctx, shootState)
				Expect(botanist.DeployInfrastructure(ctx)).To(Succeed())
			})

			It("should return the error during restoration", func() {
				infrastructure.EXPECT().Restore(ctx, shootState).Return(fakeErr)
				Expect(botanist.DeployInfrastructure(ctx)).To(MatchError(fakeErr))
			})
		})
	})

	Describe("#WaitForInfrastructure", func() {
		var (
			kubernetesGardenInterface *mockkubernetes.MockInterface
			kubernetesGardenClient    *mockclient.MockClient
			kubernetesSeedInterface   *mockkubernetes.MockInterface
			kubernetesSeedClient      *mockclient.MockClient

			namespace = "namespace"
			name      = "name"
			nodesCIDR = pointer.String("1.2.3.4/5")
			shoot     = &gardencorev1beta1.Shoot{
				ObjectMeta: metav1.ObjectMeta{
					Name:      name,
					Namespace: namespace,
				},
			}
		)

		BeforeEach(func() {
			kubernetesGardenInterface = mockkubernetes.NewMockInterface(ctrl)
			kubernetesGardenClient = mockclient.NewMockClient(ctrl)
			kubernetesSeedInterface = mockkubernetes.NewMockInterface(ctrl)
			kubernetesSeedClient = mockclient.NewMockClient(ctrl)

			botanist.K8sGardenClient = kubernetesGardenInterface
			botanist.K8sSeedClient = kubernetesSeedInterface
			botanist.Shoot.SetInfo(shoot)
		})

		It("should successfully wait (w/ provider status, w/ nodes cidr)", func() {
			infrastructure.EXPECT().Wait(ctx)
			infrastructure.EXPECT().NodesCIDR().Return(nodesCIDR)

			kubernetesGardenInterface.EXPECT().Client().Return(kubernetesGardenClient)
			updatedShoot := shoot.DeepCopy()
			updatedShoot.Spec.Networking.Nodes = nodesCIDR
			test.EXPECTPatch(ctx, kubernetesGardenClient, updatedShoot, shoot, types.StrategicMergePatchType)

			kubernetesSeedInterface.EXPECT().Client().Return(kubernetesSeedClient)

			Expect(botanist.WaitForInfrastructure(ctx)).To(Succeed())
			Expect(botanist.Shoot.GetInfo()).To(Equal(updatedShoot))
		})

		It("should successfully wait (w/o provider status, w/o nodes cidr)", func() {
			infrastructure.EXPECT().Wait(ctx)
			infrastructure.EXPECT().NodesCIDR()

			Expect(botanist.WaitForInfrastructure(ctx)).To(Succeed())
			Expect(botanist.Shoot.GetInfo()).To(Equal(shoot))
		})

		It("should return the error during wait", func() {
			infrastructure.EXPECT().Wait(ctx).Return(fakeErr)

			Expect(botanist.WaitForInfrastructure(ctx)).To(MatchError(fakeErr))
			Expect(botanist.Shoot.GetInfo()).To(Equal(shoot))
		})

		It("should return the error during nodes cidr update", func() {
			infrastructure.EXPECT().Wait(ctx)
			infrastructure.EXPECT().NodesCIDR().Return(nodesCIDR)

			kubernetesGardenInterface.EXPECT().Client().Return(kubernetesGardenClient)
			updatedShoot := shoot.DeepCopy()
			updatedShoot.Spec.Networking.Nodes = nodesCIDR
			test.EXPECTPatch(ctx, kubernetesGardenClient, updatedShoot, shoot, types.StrategicMergePatchType, fakeErr)

			Expect(botanist.WaitForInfrastructure(ctx)).To(MatchError(fakeErr))
			Expect(botanist.Shoot.GetInfo()).To(Equal(shoot))
		})
	})
})
