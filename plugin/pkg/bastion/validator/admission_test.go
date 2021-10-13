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

package validator_test

import (
	"context"

	gardencore "github.com/gardener/gardener/pkg/apis/core"
	gardencorev1beta1 "github.com/gardener/gardener/pkg/apis/core/v1beta1"
	v1beta1constants "github.com/gardener/gardener/pkg/apis/core/v1beta1/constants"
	"github.com/gardener/gardener/pkg/apis/operations"
	operationsv1alpha1 "github.com/gardener/gardener/pkg/apis/operations/v1alpha1"
	corefake "github.com/gardener/gardener/pkg/client/core/clientset/internalversion/fake"
	. "github.com/gardener/gardener/pkg/utils/test/matchers"
	. "github.com/gardener/gardener/plugin/pkg/bastion/validator"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/validation/field"
	"k8s.io/apiserver/pkg/admission"
	"k8s.io/apiserver/pkg/authentication/user"
	"k8s.io/client-go/testing"
	"k8s.io/utils/pointer"
)

const (
	bastionName = "foo"
	shootName   = "foo"
	seedName    = "foo"
	namespace   = "garden"
	provider    = "foo-provider"
	region      = "foo-region"
	userName    = "ginkgo"
)

var _ = Describe("Bastion", func() {
	Describe("#Admit", func() {
		var (
			bastion          *operations.Bastion
			shoot            *gardencore.Shoot
			coreClient       *corefake.Clientset
			dummyOwnerRef    *metav1.OwnerReference
			admissionHandler *Bastion
		)

		BeforeEach(func() {
			shoot = &gardencore.Shoot{
				ObjectMeta: metav1.ObjectMeta{
					Name:      shootName,
					Namespace: namespace,
					UID:       "shoot-uid",
				},
				Spec: gardencore.ShootSpec{
					SeedName: pointer.String(seedName),
					Provider: gardencore.Provider{
						Type: provider,
					},
					Region: region,
				},
			}

			dummyOwnerRef = &metav1.OwnerReference{
				APIVersion: "v1",
				Kind:       "ConfigMap",
				Name:       "dummy-object",
			}

			bastion = &operations.Bastion{
				ObjectMeta: metav1.ObjectMeta{
					Name:            bastionName,
					Namespace:       namespace,
					OwnerReferences: []metav1.OwnerReference{*dummyOwnerRef},
				},
				Spec: operations.BastionSpec{
					ShootRef: corev1.LocalObjectReference{
						Name: shootName,
					},
				},
			}

			var err error
			admissionHandler, err = New()
			Expect(err).ToNot(HaveOccurred())
			admissionHandler.AssignReadyFunc(func() bool { return true })

			coreClient = &corefake.Clientset{}
			admissionHandler.SetInternalCoreClientset(coreClient)
		})

		It("should do nothing if the resource is not a Bastion", func() {
			attrs := admission.NewAttributesRecord(nil, nil, operations.Kind(bastionName).WithVersion("version"), bastion.Namespace, bastion.Name, operations.Resource("foos").WithVersion("version"), "", admission.Create, &metav1.CreateOptions{}, false, nil)

			err := admissionHandler.Admit(context.TODO(), attrs, nil)
			Expect(err).To(Succeed())
		})

		It("should allow Bastion creation if all fields are set correctly", func() {
			coreClient.AddReactor("get", "shoots", func(action testing.Action) (bool, runtime.Object, error) {
				return true, shoot, nil
			})

			err := admissionHandler.Admit(context.TODO(), getBastionAttributes(bastion), nil)
			Expect(err).To(Succeed())
		})

		It("should mutate Bastion with Shoot information", func() {
			coreClient.AddReactor("get", "shoots", func(action testing.Action) (bool, runtime.Object, error) {
				return true, shoot, nil
			})

			err := admissionHandler.Admit(context.TODO(), getBastionAttributes(bastion), nil)
			Expect(err).To(Succeed())
			Expect(bastion.Spec.SeedName).To(PointTo(Equal(seedName)))
			Expect(bastion.Spec.ProviderType).To(PointTo(Equal(provider)))
		})

		It("should ensure an owner reference from the Bastion to the Shoot", func() {
			coreClient.AddReactor("get", "shoots", func(action testing.Action) (bool, runtime.Object, error) {
				return true, shoot, nil
			})

			shootOwnerRef := metav1.NewControllerRef(shoot, gardencorev1beta1.SchemeGroupVersion.WithKind("Shoot"))

			err := admissionHandler.Admit(context.TODO(), getBastionAttributes(bastion), nil)
			Expect(err).To(Succeed())
			Expect(bastion.Spec.SeedName).To(PointTo(Equal(seedName)))
			Expect(bastion.Spec.ProviderType).To(PointTo(Equal(provider)))
			Expect(bastion.OwnerReferences).To(ConsistOf(*dummyOwnerRef, *shootOwnerRef))
		})

		It("should mutate Bastion with creator information", func() {
			coreClient.AddReactor("get", "shoots", func(action testing.Action) (bool, runtime.Object, error) {
				return true, shoot, nil
			})

			err := admissionHandler.Admit(context.TODO(), getBastionAttributes(bastion), nil)
			Expect(err).To(Succeed())
			Expect(bastion.Annotations[v1beta1constants.GardenCreatedBy]).To(Equal(userName))
		})

		It("should always keep the creator annotation", func() {
			coreClient.AddReactor("get", "shoots", func(action testing.Action) (bool, runtime.Object, error) {
				return true, shoot, nil
			})

			bastion.Annotations = map[string]string{
				v1beta1constants.GardenCreatedBy: "not-" + userName,
			}

			err := admissionHandler.Admit(context.TODO(), getBastionAttributes(bastion), nil)
			Expect(err).To(Succeed())
			Expect(bastion.Annotations[v1beta1constants.GardenCreatedBy]).To(Equal(userName))
		})

		It("should forbid the Bastion creation if a Shoot name is not specified", func() {
			bastion.Spec.ShootRef.Name = ""

			err := admissionHandler.Admit(context.TODO(), getBastionAttributes(bastion), nil)
			Expect(err).To(BeInvalidError())
			Expect(getErrorList(err)).To(ConsistOf(
				PointTo(MatchFields(IgnoreExtras, Fields{
					"Type":  Equal(field.ErrorTypeRequired),
					"Field": Equal("spec.shootRef.name"),
				})),
			))
		})

		It("should forbid the Bastion creation if the Shoot does not exist", func() {
			bastion.Spec.ShootRef.Name = "does-not-exist"

			err := admissionHandler.Admit(context.TODO(), getBastionAttributes(bastion), nil)
			Expect(err).To(BeInvalidError())
			Expect(getErrorList(err)).To(ConsistOf(
				PointTo(MatchFields(IgnoreExtras, Fields{
					"Type":  Equal(field.ErrorTypeInvalid),
					"Field": Equal("spec.shootRef.name"),
				})),
			))
		})

		It("should forbid the Bastion creation if the Shoot does not specify a Seed", func() {
			shoot.Spec.SeedName = nil

			coreClient.AddReactor("get", "shoots", func(action testing.Action) (bool, runtime.Object, error) {
				return true, shoot, nil
			})

			err := admissionHandler.Admit(context.TODO(), getBastionAttributes(bastion), nil)
			Expect(err).To(BeInvalidError())
			Expect(getErrorList(err)).To(ConsistOf(
				PointTo(MatchFields(IgnoreExtras, Fields{
					"Type":  Equal(field.ErrorTypeInvalid),
					"Field": Equal("spec.shootRef.name"),
				})),
			))
		})
	})

	Describe("#Register", func() {
		It("should register the plugin", func() {
			plugins := admission.NewPlugins()
			Register(plugins)

			registered := plugins.Registered()
			Expect(registered).To(HaveLen(1))
			Expect(registered).To(ContainElement(PluginName))
		})
	})

	Describe("#New", func() {
		It("should only handle CREATE and UPDATE operations", func() {
			admissionHandler, err := New()
			Expect(err).ToNot(HaveOccurred())
			Expect(admissionHandler.Handles(admission.Create)).To(BeTrue())
			Expect(admissionHandler.Handles(admission.Update)).To(BeTrue())
			Expect(admissionHandler.Handles(admission.Connect)).To(BeFalse())
			Expect(admissionHandler.Handles(admission.Delete)).To(BeFalse())
		})
	})

	Describe("#ValidateInitialization", func() {
		It("should fail if the required clients are not set", func() {
			admissionHandler, _ := New()

			err := admissionHandler.ValidateInitialization()
			Expect(err).To(HaveOccurred())
		})

		It("should not fail if the required clients are set", func() {
			admissionHandler, _ := New()
			admissionHandler.SetInternalCoreClientset(&corefake.Clientset{})

			err := admissionHandler.ValidateInitialization()
			Expect(err).ToNot(HaveOccurred())
		})
	})
})

func getBastionAttributes(bastion *operations.Bastion) admission.Attributes {
	return admission.NewAttributesRecord(bastion,
		nil,
		operationsv1alpha1.Kind("Bastion").WithVersion("v1alpha1"),
		bastion.Namespace,
		bastion.Name,
		operationsv1alpha1.Resource("bastions").WithVersion("v1alpha1"),
		"",
		admission.Create,
		&metav1.CreateOptions{},
		false,
		&user.DefaultInfo{
			Name: userName,
		},
	)
}

func getErrorList(err error) field.ErrorList {
	statusError, ok := err.(*apierrors.StatusError)
	if !ok {
		return field.ErrorList{}
	}
	var errs field.ErrorList
	for _, cause := range statusError.ErrStatus.Details.Causes {
		errs = append(errs, &field.Error{
			Type:   field.ErrorType(cause.Type),
			Field:  cause.Field,
			Detail: cause.Message,
		})
	}
	return errs
}
