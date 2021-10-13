// Copyright (c) 2019 SAP SE or an SAP affiliate company. All rights reserved. This file is licensed under the Apache Software License, v. 2 except as noted otherwise in the LICENSE file
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

package managedresource

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var defaultEquivalences = []equivalenceList{
	newEquivalenceList("Deployment", "extensions", "apps"),
	newEquivalenceList("DaemonSet", "extensions", "apps"),
	newEquivalenceList("ReplicaSet", "extensions", "apps"),
	newEquivalenceList("StatefulSet", "extensions", "apps"),
	newEquivalenceList("Ingress", "extensions", "networking.k8s.io"),
	newEquivalenceList("NetworkPolicy", "extensions", "networking.k8s.io"),
	newEquivalenceList("PodSecurityPolicy", "extensions", "policy"),
}

type equivalenceList []metav1.GroupKind

// EquivalenceSet is a set of GroupKinds which should be considered as equivalent representation of an Object Kind.
type EquivalenceSet map[metav1.GroupKind]struct{}

// Insert adds the given GroupKinds to the EquivalenceSet
func (s EquivalenceSet) Insert(gks ...metav1.GroupKind) EquivalenceSet {
	for _, gk := range gks {
		s[gk] = struct{}{}
	}
	return s
}

// Equivalences is a set of EquivalenceSets, which can be used to look up equivalent GroupKinds for a given GroupKind.
type Equivalences map[metav1.GroupKind]EquivalenceSet

// NewEquivalences constructs a new Equivalences object, which can be used to look up equivalent GroupKinds for a given
// GroupKind. It already has some default equivalences predefined (e.g. for Kind `Deployment` in Group `apps` and
// `extensions`). It can optionally take additional lists of GroupKinds which should be considered as equivalent
// representations of the respective Object Kinds.
func NewEquivalences(additionalEquivalences ...[]metav1.GroupKind) Equivalences {
	e := Equivalences{}

	for _, equivalences := range defaultEquivalences {
		e.addEquivalentGroupKinds(equivalences)
	}

	for _, equivalences := range additionalEquivalences {
		e.addEquivalentGroupKinds(equivalences)
	}

	return e
}

func (e Equivalences) addEquivalentGroupKinds(equivalentGroupKinds []metav1.GroupKind) {
	var m EquivalenceSet

	// check if we already have an equivalence set for one of the given GroupKinds
	// if so, add the equivalents to the existing set, otherwise construct a new one
	for _, groupKind := range equivalentGroupKinds {
		if f, ok := (e)[groupKind]; ok {
			m = f
			break
		}
	}

	if m == nil {
		m = EquivalenceSet{}
	}

	// add the equivalence set for each group kind
	for _, groupKind := range equivalentGroupKinds {
		m.Insert(groupKind)
		e[groupKind] = m
	}
}

// GetEquivalencesFor looks up which GroupKinds should be considered as equivalent to a given GroupKind.
func (e Equivalences) GetEquivalencesFor(gk metav1.GroupKind) EquivalenceSet {
	return e[gk]
}

func newEquivalenceList(kind string, groups ...string) equivalenceList {
	var r equivalenceList

	for _, g := range groups {
		r = append(r, metav1.GroupKind{
			Group: g,
			Kind:  kind,
		})
	}

	return r
}
