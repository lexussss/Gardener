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

package v1alpha1

// GardenerScheduler contains the configuration of the Gardener Scheduler
type GardenerScheduler struct {
	// DeploymentConfiguration contains optional configurations for
	// the deployment of the Gardener Scheduler
	// +optional
	DeploymentConfiguration *CommonDeploymentConfiguration `json:"deploymentConfiguration,omitempty"`
	// ComponentConfiguration contains the component configuration for the Gardener Scheduler
	// +optional
	ComponentConfiguration *SchedulerComponentConfiguration `json:"componentConfiguration,omitempty"`
}

// SchedulerComponentConfiguration contains the component configuration of the Gardener Scheduler
type SchedulerComponentConfiguration struct {
	// Configuration specifies values for the Gardener Scheduler component configuration
	// Please see example/20-componentconfig-gardener-scheduler.yaml for what
	// can be configured here
	// +optional
	*Configuration `json:",inline,omitempty"`
}
