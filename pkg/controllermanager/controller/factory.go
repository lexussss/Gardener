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

package controller

import (
	"context"
	"fmt"

	gardencore "github.com/gardener/gardener/pkg/apis/core"
	gardencorev1beta1 "github.com/gardener/gardener/pkg/apis/core/v1beta1"
	"github.com/gardener/gardener/pkg/client/kubernetes/clientmap"
	"github.com/gardener/gardener/pkg/client/kubernetes/clientmap/keys"
	"github.com/gardener/gardener/pkg/controllermanager"
	"github.com/gardener/gardener/pkg/controllermanager/apis/config"
	bastioncontroller "github.com/gardener/gardener/pkg/controllermanager/controller/bastion"
	csrcontroller "github.com/gardener/gardener/pkg/controllermanager/controller/certificatesigningrequest"
	cloudprofilecontroller "github.com/gardener/gardener/pkg/controllermanager/controller/cloudprofile"
	controllerdeploymentcontroller "github.com/gardener/gardener/pkg/controllermanager/controller/controllerdeployment"
	controllerregistrationcontroller "github.com/gardener/gardener/pkg/controllermanager/controller/controllerregistration"
	eventcontroller "github.com/gardener/gardener/pkg/controllermanager/controller/event"
	exposureclasscontroller "github.com/gardener/gardener/pkg/controllermanager/controller/exposureclass"
	managedseedsetcontroller "github.com/gardener/gardener/pkg/controllermanager/controller/managedseedset"
	plantcontroller "github.com/gardener/gardener/pkg/controllermanager/controller/plant"
	projectcontroller "github.com/gardener/gardener/pkg/controllermanager/controller/project"
	quotacontroller "github.com/gardener/gardener/pkg/controllermanager/controller/quota"
	secretbindingcontroller "github.com/gardener/gardener/pkg/controllermanager/controller/secretbinding"
	seedcontroller "github.com/gardener/gardener/pkg/controllermanager/controller/seed"
	shootcontroller "github.com/gardener/gardener/pkg/controllermanager/controller/shoot"
	"github.com/gardener/gardener/pkg/controllerutils/metrics"
	gardenmetrics "github.com/gardener/gardener/pkg/controllerutils/metrics"
	"github.com/gardener/gardener/pkg/logger"
	"github.com/gardener/gardener/pkg/operation/garden"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/tools/record"
	"k8s.io/component-base/version"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// GardenControllerFactory contains information relevant to controllers for the Garden API group.
type GardenControllerFactory struct {
	cfg       *config.ControllerManagerConfiguration
	clientMap clientmap.ClientMap
	recorder  record.EventRecorder
}

// NewGardenControllerFactory creates a new factory for controllers for the Garden API group.
func NewGardenControllerFactory(clientMap clientmap.ClientMap, cfg *config.ControllerManagerConfiguration, recorder record.EventRecorder) *GardenControllerFactory {
	return &GardenControllerFactory{
		cfg:       cfg,
		clientMap: clientMap,
		recorder:  recorder,
	}
}

// Run starts all the controllers for the Garden API group. It also performs bootstrapping tasks.
func (f *GardenControllerFactory) Run(ctx context.Context) error {
	gardenClientSet, err := f.clientMap.GetClient(ctx, keys.ForGarden())
	if err != nil {
		return fmt.Errorf("failed to get garden client: %+v", err)
	}

	if err := addAllFieldIndexes(ctx, gardenClientSet.Cache()); err != nil {
		return err
	}

	if err := f.clientMap.Start(ctx.Done()); err != nil {
		return fmt.Errorf("failed to start ClientMap: %+v", err)
	}

	runtime.Must(garden.BootstrapCluster(ctx, gardenClientSet))
	logger.Logger.Info("Successfully bootstrapped the Garden cluster.")

	// Initialize the workqueue metrics collection.
	var metricsCollectors []metrics.ControllerMetricsCollector
	gardenmetrics.RegisterWorkqueMetrics()

	// Create controllers.
	bastionController, err := bastioncontroller.NewBastionController(ctx, f.clientMap, f.cfg.Controllers.Bastion.MaxLifetime.Duration)
	if err != nil {
		return fmt.Errorf("failed initializing Bastion controller: %w", err)
	}
	metricsCollectors = append(metricsCollectors, bastionController)

	cloudProfileController, err := cloudprofilecontroller.NewCloudProfileController(ctx, f.clientMap, f.recorder, logger.Logger)
	if err != nil {
		return fmt.Errorf("failed initializing CloudProfile controller: %w", err)
	}
	metricsCollectors = append(metricsCollectors, cloudProfileController)

	controllerRegistrationController, err := controllerregistrationcontroller.NewController(ctx, f.clientMap)
	if err != nil {
		return fmt.Errorf("failed initializing ControllerRegistration controller: %w", err)
	}
	metricsCollectors = append(metricsCollectors, controllerRegistrationController)

	csrController, err := csrcontroller.NewCSRController(ctx, f.clientMap)
	if err != nil {
		return fmt.Errorf("failed initializing CSR controller: %w", err)
	}
	metricsCollectors = append(metricsCollectors, csrController)

	exposureClassController, err := exposureclasscontroller.NewExposureClassController(ctx, f.clientMap, f.recorder)
	if err != nil {
		return fmt.Errorf("failed initializing ExposureClass controller: %w", err)
	}
	metricsCollectors = append(metricsCollectors, exposureClassController)

	plantController, err := plantcontroller.NewController(ctx, f.clientMap, f.cfg)
	if err != nil {
		return fmt.Errorf("failed initializing Plant controller: %w", err)
	}
	metricsCollectors = append(metricsCollectors, plantController)

	projectController, err := projectcontroller.NewProjectController(ctx, f.clientMap, f.cfg, f.recorder)
	if err != nil {
		return fmt.Errorf("failed initializing Project controller: %w", err)
	}
	metricsCollectors = append(metricsCollectors, projectController)

	quotaController, err := quotacontroller.NewQuotaController(ctx, f.clientMap, f.recorder)
	if err != nil {
		return fmt.Errorf("failed initializing Quota controller: %w", err)
	}
	metricsCollectors = append(metricsCollectors, quotaController)

	secretBindingController, err := secretbindingcontroller.NewSecretBindingController(ctx, f.clientMap, f.recorder)
	if err != nil {
		return fmt.Errorf("failed initializing SecretBinding controller: %w", err)
	}
	metricsCollectors = append(metricsCollectors, secretBindingController)

	seedController, err := seedcontroller.NewSeedController(ctx, f.clientMap, f.cfg)
	if err != nil {
		return fmt.Errorf("failed initializing Seed controller: %w", err)
	}
	metricsCollectors = append(metricsCollectors, seedController)

	controllerDeploymentController, err := controllerdeploymentcontroller.New(ctx, f.clientMap, logger.Logger)
	if err != nil {
		return fmt.Errorf("failed initializing ControllerDeployment controller: %w", err)
	}
	metricsCollectors = append(metricsCollectors, controllerDeploymentController)

	shootController, err := shootcontroller.NewShootController(ctx, f.clientMap, f.cfg, f.recorder)
	if err != nil {
		return fmt.Errorf("failed initializing Shoot controller: %w", err)
	}
	metricsCollectors = append(metricsCollectors, shootController)

	managedSeedSetController, err := managedseedsetcontroller.NewManagedSeedSetController(ctx, f.clientMap, f.cfg, f.recorder, logger.Logger)
	if err != nil {
		return fmt.Errorf("failed initializing ManagedSeedSet controller: %w", err)
	}
	metricsCollectors = append(metricsCollectors, managedSeedSetController)

	go bastionController.Run(ctx, f.cfg.Controllers.Bastion.ConcurrentSyncs)
	go cloudProfileController.Run(ctx, f.cfg.Controllers.CloudProfile.ConcurrentSyncs)
	go controllerDeploymentController.Run(ctx, f.cfg.Controllers.ControllerDeployment.ConcurrentSyncs)
	go controllerRegistrationController.Run(ctx, f.cfg.Controllers.ControllerRegistration.ConcurrentSyncs)
	go csrController.Run(ctx, 1)
	go plantController.Run(ctx, f.cfg.Controllers.Plant.ConcurrentSyncs)
	go projectController.Run(ctx, f.cfg.Controllers.Project.ConcurrentSyncs)
	go quotaController.Run(ctx, f.cfg.Controllers.Quota.ConcurrentSyncs)
	go secretBindingController.Run(ctx, f.cfg.Controllers.SecretBinding.ConcurrentSyncs)
	go seedController.Run(ctx, f.cfg.Controllers.Seed.ConcurrentSyncs)
	go shootController.Run(ctx, f.cfg.Controllers.ShootMaintenance.ConcurrentSyncs, f.cfg.Controllers.ShootQuota.ConcurrentSyncs, f.cfg.Controllers.ShootHibernation.ConcurrentSyncs, f.cfg.Controllers.ShootReference.ConcurrentSyncs, f.cfg.Controllers.ShootRetry.ConcurrentSyncs)
	go exposureClassController.Run(ctx, f.cfg.Controllers.ExposureClass.ConcurrentSyncs)
	go managedSeedSetController.Run(ctx, f.cfg.Controllers.ManagedSeedSet.ConcurrentSyncs)

	if eventControllerConfig := f.cfg.Controllers.Event; eventControllerConfig != nil {
		eventController, err := eventcontroller.NewController(ctx, f.clientMap, eventControllerConfig)
		if err != nil {
			return fmt.Errorf("failed initializing Event controller: %w", err)
		}
		metricsCollectors = append(metricsCollectors, eventController)

		go eventController.Run(ctx)
	}

	// Initialize the Controller metrics collection.
	gardenmetrics.RegisterControllerMetrics(controllermanager.ControllerWorkerSum, controllermanager.ScrapeFailures, metricsCollectors...)

	logger.Logger.Infof("Gardener controller manager (version %s) initialized.", version.Get().GitVersion)

	// Shutdown handling
	<-ctx.Done()

	logger.Logger.Infof("I have received a stop signal and will no longer watch resources.")
	logger.Logger.Infof("Bye Bye!")

	return nil
}

// addAllFieldIndexes adds all field indexes used by gardener-controller-manager to the given FieldIndexer (i.e. cache).
// field indexes have to be added before the cache is started (i.e. before the clientmap is started)
func addAllFieldIndexes(ctx context.Context, indexer client.FieldIndexer) error {
	if err := indexer.IndexField(ctx, &gardencorev1beta1.Project{}, gardencore.ProjectNamespace, func(obj client.Object) []string {
		project, ok := obj.(*gardencorev1beta1.Project)
		if !ok {
			return []string{""}
		}
		if project.Spec.Namespace == nil {
			return []string{""}
		}
		return []string{*project.Spec.Namespace}
	}); err != nil {
		return fmt.Errorf("failed to add indexer to Project Informer: %w", err)
	}

	if err := indexer.IndexField(ctx, &gardencorev1beta1.Shoot{}, gardencore.ShootSeedName, func(obj client.Object) []string {
		shoot, ok := obj.(*gardencorev1beta1.Shoot)
		if !ok {
			return []string{""}
		}
		if shoot.Spec.SeedName == nil {
			return []string{""}
		}
		return []string{*shoot.Spec.SeedName}
	}); err != nil {
		return fmt.Errorf("failed to add indexer to Shoot Informer: %w", err)
	}

	return nil
}
