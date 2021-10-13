// Copyright (c) 2018 SAP SE or an SAP affiliate company. All rights reserved. This file is licensed under the Apache Software License, v. 2 except as noted otherwise in the LICENSE file
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

package shootstate

import (
	"context"

	"github.com/gardener/gardener/pkg/api"
	"github.com/gardener/gardener/pkg/apis/core"
	"github.com/gardener/gardener/pkg/apis/core/validation"

	apiequality "k8s.io/apimachinery/pkg/api/equality"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/validation/field"
	"k8s.io/apiserver/pkg/storage/names"
)

type shootStateStrategy struct {
	runtime.ObjectTyper
	names.NameGenerator
}

// Strategy defines the storage strategy for plants.
var Strategy = shootStateStrategy{api.Scheme, names.SimpleNameGenerator}

func (shootStateStrategy) NamespaceScoped() bool {
	return true
}

func (shootStateStrategy) PrepareForCreate(ctx context.Context, obj runtime.Object) {
	shootState := obj.(*core.ShootState)

	shootState.Generation = 1
}

func (shootStateStrategy) PrepareForUpdate(ctx context.Context, obj, old runtime.Object) {
	newShootState := obj.(*core.ShootState)
	oldShootState := old.(*core.ShootState)

	if mustIncreaseGeneration(oldShootState, newShootState) {
		newShootState.Generation = oldShootState.Generation + 1
	}
}

func mustIncreaseGeneration(oldShootState, newShootState *core.ShootState) bool {
	// The ShootState specification changes.
	if !apiequality.Semantic.DeepEqual(oldShootState.Spec, newShootState.Spec) {
		return true
	}

	// The deletion timestamp was set.
	if oldShootState.DeletionTimestamp == nil && newShootState.DeletionTimestamp != nil {
		return true
	}

	return false
}

func (shootStateStrategy) Validate(ctx context.Context, obj runtime.Object) field.ErrorList {
	shootState := obj.(*core.ShootState)
	return validation.ValidateShootState(shootState)
}

func (shootStateStrategy) Canonicalize(obj runtime.Object) {
}

func (shootStateStrategy) AllowCreateOnUpdate() bool {
	return false
}

func (shootStateStrategy) ValidateUpdate(ctx context.Context, newObj, oldObj runtime.Object) field.ErrorList {
	newShootState := newObj.(*core.ShootState)
	oldShootState := oldObj.(*core.ShootState)
	return validation.ValidateShootStateUpdate(newShootState, oldShootState)
}

func (shootStateStrategy) AllowUnconditionalUpdate() bool {
	return false
}
