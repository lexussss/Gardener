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

package validation

import (
	apisconfig "github.com/gardener/gardener/pkg/admissioncontroller/apis/config"

	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/apimachinery/pkg/util/validation/field"
)

// ValidateAdmissionControllerConfiguration validates the given `AdmissionControllerConfiguration`.
func ValidateAdmissionControllerConfiguration(config *apisconfig.AdmissionControllerConfiguration) field.ErrorList {
	allErrs := field.ErrorList{}

	serverPath := field.NewPath("server")
	if config.Server.ResourceAdmissionConfiguration != nil {
		allErrs = append(allErrs, validateResourceAdmissionConfiguration(config.Server.ResourceAdmissionConfiguration, serverPath.Child("resourceAdmissionConfiguration"))...)
	}
	return allErrs
}

// ValidateResourceAdmissionConfiguration validates the given `ResourceAdmissionConfiguration`.
func validateResourceAdmissionConfiguration(config *apisconfig.ResourceAdmissionConfiguration, fldPath *field.Path) field.ErrorList {
	allErrs := field.ErrorList{}
	validValues := sets.NewString(string(apisconfig.AdmissionModeBlock), string(apisconfig.AdmissionModeLog))

	if config.OperationMode != nil && !validValues.Has(string(*config.OperationMode)) {
		allErrs = append(allErrs, field.NotSupported(fldPath.Child("mode"), string(*config.OperationMode), validValues.UnsortedList()))
	}

	allowedSubjectKinds := sets.NewString(rbacv1.UserKind, rbacv1.GroupKind, rbacv1.ServiceAccountKind)

	for i, subject := range config.UnrestrictedSubjects {
		fld := fldPath.Child("unrestrictedSubjects").Index(i)

		if !allowedSubjectKinds.Has(subject.Kind) {
			allErrs = append(allErrs, field.NotSupported(fld.Child("kind"), subject.Kind, allowedSubjectKinds.UnsortedList()))
		}
		if subject.Name == "" {
			allErrs = append(allErrs, field.Invalid(fld.Child("name"), subject.Name, "name must not be empty"))
		}

		switch subject.Kind {
		case rbacv1.ServiceAccountKind:
			if subject.Namespace == "" {
				allErrs = append(allErrs, field.Invalid(fld.Child("namespace"), subject.Namespace, "name must not be empty"))
			}
			if subject.APIGroup != "" {
				allErrs = append(allErrs, field.Invalid(fld.Child("apiGroup"), subject.APIGroup, "apiGroup must be empty"))
			}
		case rbacv1.UserKind, rbacv1.GroupKind:
			if subject.Namespace != "" {
				allErrs = append(allErrs, field.Invalid(fld.Child("namespace"), subject.Namespace, "name must be empty"))
			}
			if subject.APIGroup != rbacv1.GroupName {
				allErrs = append(allErrs, field.NotSupported(fld.Child("apiGroup"), subject.APIGroup, []string{rbacv1.GroupName}))
			}
		}
	}

	for i, limit := range config.Limits {
		fld := fldPath.Child("limits").Index(i)
		hasResources := false
		for j, resource := range limit.Resources {
			hasResources = true
			if resource == "" {
				allErrs = append(allErrs, field.Invalid(fld.Child("resources").Index(j), resource, "must not be empty"))
			}
		}
		if !hasResources {
			allErrs = append(allErrs, field.Invalid(fld.Child("resources"), limit.Resources, "must at least have one element"))
		}

		if len(limit.APIGroups) < 1 {
			allErrs = append(allErrs, field.Invalid(fld.Child("apiGroups"), limit.Resources, "must at least have one element"))
		}

		hasVersions := false
		for j, version := range limit.APIVersions {
			hasVersions = true
			if version == "" {
				allErrs = append(allErrs, field.Invalid(fld.Child("versions").Index(j), version, "must not be empty"))
			}
		}
		if !hasVersions {
			allErrs = append(allErrs, field.Invalid(fld.Child("versions"), limit.Resources, "must at least have one element"))
		}

		if limit.Size.Cmp(resource.Quantity{}) < 0 {
			allErrs = append(allErrs, field.Invalid(fld.Child("size"), limit.Size.String(), "value must not be negative"))
		}
	}

	return allErrs
}
