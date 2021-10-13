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

package predicate

import (
	resourcesv1alpha1 "github.com/gardener/gardener/pkg/apis/resources/v1alpha1"

	"k8s.io/apimachinery/pkg/api/equality"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

var classChangedPredicate = predicate.Funcs{
	UpdateFunc: func(e event.UpdateEvent) bool {
		if e.ObjectOld == nil {
			log.Error(nil, "Update event has no old runtime object to update", "event", e)
			return false
		}
		if e.ObjectNew == nil {
			log.Error(nil, "Update event has no new runtime object for update", "event", e)
			return false
		}

		old, ok := e.ObjectOld.(*resourcesv1alpha1.ManagedResource)
		if !ok {
			log.Error(nil, "Update event old runtime object cannot be converted to ManagedResource", "event", e)
			return false
		}
		new, ok := e.ObjectNew.(*resourcesv1alpha1.ManagedResource)
		if !ok {
			log.Error(nil, "Update event new runtime object cannot be converted to ManagedResource", "event", e)
			return false
		}

		return !equality.Semantic.DeepEqual(old.Spec.Class, new.Spec.Class)
	},
}

// ClassChangedPredicate is a predicate for changes in `.spec.class`.
func ClassChangedPredicate() predicate.Predicate {
	return classChangedPredicate
}
