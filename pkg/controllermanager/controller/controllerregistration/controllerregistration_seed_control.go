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

package controllerregistration

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/gardener/gardener/pkg/apis/core"
	gardencorev1beta1 "github.com/gardener/gardener/pkg/apis/core/v1beta1"
	gardencorev1beta1helper "github.com/gardener/gardener/pkg/apis/core/v1beta1/helper"
	extensionsv1alpha1 "github.com/gardener/gardener/pkg/apis/extensions/v1alpha1"
	"github.com/gardener/gardener/pkg/client/kubernetes"
	controllermanagerfeatures "github.com/gardener/gardener/pkg/controllermanager/features"
	"github.com/gardener/gardener/pkg/controllerutils"
	"github.com/gardener/gardener/pkg/extensions"
	"github.com/gardener/gardener/pkg/features"
	"github.com/gardener/gardener/pkg/logger"
	"github.com/gardener/gardener/pkg/operation/common"
	gardenpkg "github.com/gardener/gardener/pkg/operation/garden"
	shootpkg "github.com/gardener/gardener/pkg/operation/shoot"
	"github.com/gardener/gardener/pkg/utils"
	gutil "github.com/gardener/gardener/pkg/utils/gardener"
	kutil "github.com/gardener/gardener/pkg/utils/kubernetes"

	dnsv1alpha1 "github.com/gardener/external-dns-management/pkg/apis/dns/v1alpha1"
	"github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/util/sets"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

// NewControllerRegistrationSeedReconciler creates a new instance of a reconciler which determines which
// ControllerRegistrations are required for a seed.
func NewControllerRegistrationSeedReconciler(logger logrus.FieldLogger, gardenClient kubernetes.Interface) reconcile.Reconciler {
	return &controllerRegistrationSeedReconciler{
		logger:       logger,
		gardenClient: gardenClient,
	}
}

type controllerRegistrationSeedReconciler struct {
	logger       logrus.FieldLogger
	gardenClient kubernetes.Interface
}

func (r *controllerRegistrationSeedReconciler) Reconcile(ctx context.Context, request reconcile.Request) (reconcile.Result, error) {
	seed := &gardencorev1beta1.Seed{}
	if err := r.gardenClient.Client().Get(ctx, request.NamespacedName, seed); err != nil {
		if apierrors.IsNotFound(err) {
			r.logger.Infof("Object %q is gone, stop reconciling: %v", request.Name, err)
			return reconcile.Result{}, nil
		}
		r.logger.Infof("Unable to retrieve object %q from store: %v", request.Name, err)
		return reconcile.Result{}, err
	}

	logger := logger.NewFieldLogger(r.logger, "controllerregistration-seed", seed.Name)
	logger.Info("[CONTROLLERINSTALLATION SEED] Reconciling")

	controllerRegistrationList := &gardencorev1beta1.ControllerRegistrationList{}
	if err := r.gardenClient.Client().List(ctx, controllerRegistrationList); err != nil {
		return reconcile.Result{}, err
	}

	// Live lookup to prevent working on a stale cache and trying to create multiple installations for the same
	// registration/seed combination.
	controllerInstallationList := &gardencorev1beta1.ControllerInstallationList{}
	if err := r.gardenClient.APIReader().List(ctx, controllerInstallationList); err != nil {
		return reconcile.Result{}, err
	}

	backupBucketList := &gardencorev1beta1.BackupBucketList{}
	if err := r.gardenClient.Client().List(ctx, backupBucketList); err != nil {
		return reconcile.Result{}, err
	}

	backupEntryList := &gardencorev1beta1.BackupEntryList{}
	if err := r.gardenClient.APIReader().List(ctx, backupEntryList, client.MatchingFields{core.BackupEntrySeedName: seed.Name}); err != nil {
		return reconcile.Result{}, err
	}

	shootList, err := getShoots(ctx, r.gardenClient.APIReader(), seed)
	if err != nil {
		return reconcile.Result{}, err
	}

	secrets, err := gardenpkg.ReadGardenSecrets(ctx, r.gardenClient.Client(), gutil.ComputeGardenNamespace(seed.Name), logger)
	if err != nil {
		return reconcile.Result{}, err
	}

	if len(secrets) < 1 {
		return reconcile.Result{}, fmt.Errorf("garden secrets for seed %q have not been synchronized yet", seed.Name)
	}

	internalDomain, err := gardenpkg.GetInternalDomain(secrets)
	if err != nil {
		return reconcile.Result{}, err
	}
	defaultDomains, err := gardenpkg.GetDefaultDomains(secrets)
	if err != nil {
		return reconcile.Result{}, err
	}

	var (
		controllerRegistrations = computeControllerRegistrationMaps(controllerRegistrationList)

		wantedKindTypeCombinationForBackupBuckets, buckets = computeKindTypesForBackupBuckets(backupBucketList, seed.Name)
		wantedKindTypeCombinationForBackupEntries          = computeKindTypesForBackupEntries(logger, backupEntryList, buckets, seed.Name)
		wantedKindTypeCombinationForShoots                 = computeKindTypesForShoots(ctx, logger, r.gardenClient.Client(), shootList, seed, controllerRegistrationList, internalDomain, defaultDomains)
		wantedKindTypeCombinationForSeed                   = computeKindTypesForSeed(seed)

		wantedKindTypeCombinations = sets.
						NewString().
						Union(wantedKindTypeCombinationForBackupBuckets).
						Union(wantedKindTypeCombinationForBackupEntries).
						Union(wantedKindTypeCombinationForShoots).
						Union(wantedKindTypeCombinationForSeed)
	)

	wantedControllerRegistrationNames, err := computeWantedControllerRegistrationNames(wantedKindTypeCombinations, controllerInstallationList, controllerRegistrations, len(shootList), seed.ObjectMeta)
	if err != nil {
		return reconcile.Result{}, err
	}

	registrationNameToInstallation, err := computeRegistrationNameToInstallationMap(controllerInstallationList, controllerRegistrations, seed.Name)
	if err != nil {
		return reconcile.Result{}, err
	}

	if err := deployNeededInstallations(ctx, logger, r.gardenClient.Client(), seed, wantedControllerRegistrationNames, controllerRegistrations, registrationNameToInstallation); err != nil {
		return reconcile.Result{}, err
	}

	if err := deleteUnneededInstallations(ctx, logger, r.gardenClient.Client(), wantedControllerRegistrationNames, registrationNameToInstallation); err != nil {
		return reconcile.Result{}, err
	}

	return reconcile.Result{}, nil
}

// computeKindTypesForBackupBucket computes the list of wanted kind/type combinations for extension resources based on the
// the list of existing BackupBucket resources.
func computeKindTypesForBackupBuckets(
	backupBucketList *gardencorev1beta1.BackupBucketList,
	seedName string,
) (
	sets.String,
	map[string]gardencorev1beta1.BackupBucket,
) {
	var (
		wantedKindTypeCombinations = sets.NewString()
		buckets                    = make(map[string]gardencorev1beta1.BackupBucket)
	)

	for _, backupBucket := range backupBucketList.Items {
		buckets[backupBucket.Name] = backupBucket

		if backupBucket.Spec.SeedName == nil || *backupBucket.Spec.SeedName != seedName {
			continue
		}

		wantedKindTypeCombinations.Insert(extensions.Id(extensionsv1alpha1.BackupBucketResource, backupBucket.Spec.Provider.Type))
	}

	return wantedKindTypeCombinations, buckets
}

// computeKindTypesForBackupEntries computes the list of wanted kind/type combinations for extension resources based on the
// the list of existing BackupEntry resources.
func computeKindTypesForBackupEntries(
	logger *logrus.Entry,
	backupEntryList *gardencorev1beta1.BackupEntryList,
	buckets map[string]gardencorev1beta1.BackupBucket,
	seedName string,
) sets.String {
	wantedKindTypeCombinations := sets.NewString()

	for _, backupEntry := range backupEntryList.Items {
		if backupEntry.Spec.SeedName == nil || *backupEntry.Spec.SeedName != seedName {
			continue
		}

		bucket, ok := buckets[backupEntry.Spec.BucketName]
		if !ok {
			logger.Errorf("couldn't find BackupBucket %q for BackupEntry %q", backupEntry.Spec.BucketName, backupEntry.Name)
			continue
		}

		wantedKindTypeCombinations.Insert(extensions.Id(extensionsv1alpha1.BackupEntryResource, bucket.Spec.Provider.Type))
	}

	return wantedKindTypeCombinations
}

// computeKindTypesForShoots computes the list of wanted kind/type combinations for extension resources based on the
// the list of existing Shoot resources.
func computeKindTypesForShoots(
	ctx context.Context,
	logger *logrus.Entry,
	client client.Client,
	shootList []gardencorev1beta1.Shoot,
	seed *gardencorev1beta1.Seed,
	controllerRegistrationList *gardencorev1beta1.ControllerRegistrationList,
	internalDomain *gardenpkg.Domain,
	defaultDomains []*gardenpkg.Domain,
) sets.String {
	var (
		wantedKindTypeCombinations = sets.NewString()

		wg  sync.WaitGroup
		out = make(chan sets.String)
	)

	for _, shoot := range shootList {
		if (shoot.Spec.SeedName == nil || *shoot.Spec.SeedName != seed.Name) && (shoot.Status.SeedName == nil || *shoot.Status.SeedName != seed.Name) {
			continue
		}

		wg.Add(1)
		go func(shoot *gardencorev1beta1.Shoot) {
			defer wg.Done()

			externalDomain, err := shootpkg.ConstructExternalDomain(ctx, client, shoot, &corev1.Secret{}, defaultDomains)
			if err != nil && !(shootpkg.IsIncompleteDNSConfigError(err) && shoot.DeletionTimestamp != nil && len(shoot.Status.UID) == 0) {
				logger.Warnf("could not determine external domain for shoot %s/%s: %+v", shoot.Namespace, shoot.Name, err)
			}

			out <- shootpkg.ComputeRequiredExtensions(shoot, seed, controllerRegistrationList, internalDomain, externalDomain,
				controllermanagerfeatures.FeatureGate.Enabled(features.UseDNSRecords))
		}(shoot.DeepCopy())
	}

	go func() {
		wg.Wait()
		close(out)
	}()

	for result := range out {
		wantedKindTypeCombinations = wantedKindTypeCombinations.Union(result)
	}

	return wantedKindTypeCombinations
}

// computeKindTypesForSeed computes the list of wanted kind/type combinations for extension resources based on the
// Seed configuration
func computeKindTypesForSeed(
	seed *gardencorev1beta1.Seed,
) sets.String {
	var wantedKindTypeCombinations = sets.NewString()

	// enable clean up of controller installations in case of seed deletion
	if seed.DeletionTimestamp != nil {
		return sets.NewString()
	}

	if seed.Spec.DNS.Provider != nil {
		if controllermanagerfeatures.FeatureGate.Enabled(features.UseDNSRecords) {
			wantedKindTypeCombinations.Insert(extensions.Id(extensionsv1alpha1.DNSRecordResource, seed.Spec.DNS.Provider.Type))
		} else {
			wantedKindTypeCombinations.Insert(extensions.Id(dnsv1alpha1.DNSProviderKind, seed.Spec.DNS.Provider.Type))
		}
	}

	return wantedKindTypeCombinations
}

type controllerRegistration struct {
	obj                        *gardencorev1beta1.ControllerRegistration
	deployAlways               bool
	deployAlwaysExceptNoShoots bool
}

// computeControllerRegistrationMaps computes a map which maps the name of a ControllerRegistration to the
// *gardencorev1beta1.ControllerRegistration object. It also specifies whether the ControllerRegistration shall be
// always deployed.
func computeControllerRegistrationMaps(
	controllerRegistrationList *gardencorev1beta1.ControllerRegistrationList,
) map[string]controllerRegistration {
	var out = make(map[string]controllerRegistration)
	for _, cr := range controllerRegistrationList.Items {
		out[cr.Name] = controllerRegistration{
			obj:                        cr.DeepCopy(),
			deployAlways:               cr.Spec.Deployment != nil && cr.Spec.Deployment.Policy != nil && *cr.Spec.Deployment.Policy == gardencorev1beta1.ControllerDeploymentPolicyAlways,
			deployAlwaysExceptNoShoots: cr.Spec.Deployment != nil && cr.Spec.Deployment.Policy != nil && *cr.Spec.Deployment.Policy == gardencorev1beta1.ControllerDeploymentPolicyAlwaysExceptNoShoots,
		}
	}
	return out
}

// computeWantedControllerRegistrationNames computes the list of names of ControllerRegistration objects that are desired
// to be installed. The computation is performed based on a list of required kind/type combinations and the proper mapping
// to existing ControllerRegistration objects. Additionally, all names in the alwaysPolicyControllerRegistrationNames list
// will be returned and all currently installed and required installations.
func computeWantedControllerRegistrationNames(
	wantedKindTypeCombinations sets.String,
	controllerInstallationList *gardencorev1beta1.ControllerInstallationList,
	controllerRegistrations map[string]controllerRegistration,
	numberOfShoots int,
	seedObjectMeta metav1.ObjectMeta,
) (
	sets.String,
	error,
) {
	var (
		kindTypeToControllerRegistrationNames = make(map[string][]string)
		wantedControllerRegistrationNames     = sets.NewString()
	)

	for name, controllerRegistration := range controllerRegistrations {
		if controllerRegistration.deployAlways && seedObjectMeta.DeletionTimestamp == nil {
			wantedControllerRegistrationNames.Insert(name)
		}

		if controllerRegistration.deployAlwaysExceptNoShoots && numberOfShoots > 0 {
			wantedControllerRegistrationNames.Insert(name)
		}

		for _, resource := range controllerRegistration.obj.Spec.Resources {
			id := extensions.Id(resource.Kind, resource.Type)
			kindTypeToControllerRegistrationNames[id] = append(kindTypeToControllerRegistrationNames[id], name)
		}
	}

	for _, wantedExtension := range wantedKindTypeCombinations.UnsortedList() {
		names, ok := kindTypeToControllerRegistrationNames[wantedExtension]
		if !ok {
			return nil, fmt.Errorf("need to install an extension controller for %q but no appropriate ControllerRegistration found", wantedExtension)
		}
		wantedControllerRegistrationNames.Insert(names...)
	}

	wantedControllerRegistrationNames.Insert(installedAndRequiredRegistrationNames(controllerInstallationList, seedObjectMeta.Name).List()...)

	// filter controller registrations with non-matching seed selector
	return controllerRegistrationNamesWithMatchingSeedLabelSelector(wantedControllerRegistrationNames.UnsortedList(), controllerRegistrations, seedObjectMeta.Labels)
}

func installedAndRequiredRegistrationNames(controllerInstallationList *gardencorev1beta1.ControllerInstallationList, seedName string) sets.String {
	requiredControllerRegistrationNames := sets.NewString()
	for _, controllerInstallation := range controllerInstallationList.Items {
		if controllerInstallation.Spec.SeedRef.Name != seedName {
			continue
		}
		if !gardencorev1beta1helper.IsControllerInstallationRequired(controllerInstallation) {
			continue
		}
		requiredControllerRegistrationNames.Insert(controllerInstallation.Spec.RegistrationRef.Name)
	}
	return requiredControllerRegistrationNames
}

// computeRegistrationNameToInstallationMap computes a map that maps the name of a ControllerRegistration to an
// existing ControllerInstallation object that references this registration.
func computeRegistrationNameToInstallationMap(
	controllerInstallationList *gardencorev1beta1.ControllerInstallationList,
	controllerRegistrations map[string]controllerRegistration,
	seedName string,
) (
	map[string]*gardencorev1beta1.ControllerInstallation,
	error,
) {
	registrationNameToInstallationName := make(map[string]*gardencorev1beta1.ControllerInstallation)

	for _, controllerInstallation := range controllerInstallationList.Items {
		if controllerInstallation.Spec.SeedRef.Name != seedName {
			continue
		}

		if _, ok := controllerRegistrations[controllerInstallation.Spec.RegistrationRef.Name]; !ok {
			return nil, fmt.Errorf("ControllerRegistration %q does not exist", controllerInstallation.Spec.RegistrationRef.Name)
		}

		controllerInstallationObj := controllerInstallation
		registrationNameToInstallationName[controllerInstallation.Spec.RegistrationRef.Name] = &controllerInstallationObj
	}

	return registrationNameToInstallationName, nil
}

// deployNeededInstallations takes the list of required names of ControllerRegistrations, a mapping of ControllerRegistration
// names to their actual objects, and another mapping of ControllerRegistration names to existing ControllerInstallations. It
// creates or update ControllerInstallation objects for that reference the given seed and the various desired ControllerRegistrations.
func deployNeededInstallations(
	ctx context.Context,
	logger *logrus.Entry,
	c client.Client,
	seed *gardencorev1beta1.Seed,
	wantedControllerRegistrations sets.String,
	controllerRegistrations map[string]controllerRegistration,
	registrationNameToInstallation map[string]*gardencorev1beta1.ControllerInstallation,
) error {
	for _, registrationName := range wantedControllerRegistrations.UnsortedList() {
		// Sometimes an operator needs to migrate to a new controller registration that supports the required
		// kind and types, but it is required to offboard the old extension. Thus, the operator marks the old
		// controller registration for deletion and manually delete its controller installation.
		// In parallel, Gardener should not create new controller installations for the deleted controller registation.
		if controllerRegistrations[registrationName].obj.DeletionTimestamp != nil {
			logger.Infof("Do not create or update ControllerInstallation for %q which is in deletion", registrationName)
			continue
		}

		logger.Infof("Deploying wanted ControllerInstallation for %q", registrationName)

		var (
			controllerDeployment   *gardencorev1beta1.ControllerDeployment
			controllerRegistration = controllerRegistrations[registrationName].obj
		)

		if controllerRegistration.Spec.Deployment != nil && len(controllerRegistration.Spec.Deployment.DeploymentRefs) > 0 {
			// Today, only one DeploymentRef element is allowed, which is why can simply pick the first one from the slice.
			controllerDeployment = &gardencorev1beta1.ControllerDeployment{}

			if err := c.Get(ctx, kutil.Key(controllerRegistration.Spec.Deployment.DeploymentRefs[0].Name), controllerDeployment); err != nil {
				return fmt.Errorf("cannot deploy ControllerInstallation because the referenced ControllerDeployment cannot be retrieved: %w", err)
			}
		}

		existingControllerInstallation := registrationNameToInstallation[registrationName]
		if existingControllerInstallation != nil && existingControllerInstallation.DeletionTimestamp != nil {
			return fmt.Errorf("cannot deploy new ControllerInstallation for %q because the deletion of the old ControllerInstallation is still pending", registrationName)
		}

		if err := deployNeededInstallation(ctx, c, seed, controllerDeployment, controllerRegistration, existingControllerInstallation); err != nil {
			return err
		}
	}

	return nil
}

func deployNeededInstallation(
	ctx context.Context,
	c client.Client,
	seed *gardencorev1beta1.Seed,
	controllerDeployment *gardencorev1beta1.ControllerDeployment,
	controllerRegistration *gardencorev1beta1.ControllerRegistration,
	existingControllerInstallation *gardencorev1beta1.ControllerInstallation,
) error {
	installationSpec := gardencorev1beta1.ControllerInstallationSpec{
		SeedRef: corev1.ObjectReference{
			Name:            seed.Name,
			ResourceVersion: seed.ResourceVersion,
		},
		RegistrationRef: corev1.ObjectReference{
			Name:            controllerRegistration.Name,
			ResourceVersion: controllerRegistration.ResourceVersion,
		},
	}

	if controllerDeployment != nil {
		installationSpec.DeploymentRef = &corev1.ObjectReference{
			Name:            controllerDeployment.Name,
			ResourceVersion: controllerDeployment.ResourceVersion,
		}
	}

	controllerInstallation := &gardencorev1beta1.ControllerInstallation{}
	mutate := func() error {
		seedSpecMap, err := convertObjToMap(seed.Spec)
		if err != nil {
			return err
		}
		seedSpecHash := utils.HashForMap(seedSpecMap)[:16]
		kutil.SetMetaDataLabel(&controllerInstallation.ObjectMeta, common.SeedSpecHash, seedSpecHash)

		registrationSpecMap, err := convertObjToMap(controllerRegistration.Spec)
		if err != nil {
			return err
		}
		registrationSpecHash := utils.HashForMap(registrationSpecMap)[:16]
		kutil.SetMetaDataLabel(&controllerInstallation.ObjectMeta, common.RegistrationSpecHash, registrationSpecHash)

		if controllerDeployment != nil {
			// Add all fields that are relevant for the hash calculation as `ControllerDeployment`s don't have a `spec` field.
			hashFields := map[string]interface{}{
				"type":           controllerDeployment.Type,
				"providerConfig": controllerDeployment.ProviderConfig,
			}

			deploymentMap, err := convertObjToMap(hashFields)
			if err != nil {
				return err
			}
			deploymentSpecHash := utils.HashForMap(deploymentMap)[:16]
			kutil.SetMetaDataLabel(&controllerInstallation.ObjectMeta, common.ControllerDeploymentHash, deploymentSpecHash)
		}
		controllerInstallation.Spec = installationSpec
		return nil
	}

	if existingControllerInstallation != nil {
		// The installation already exists, however, we do not have the latest version of the ControllerInstallation object.
		// Hence, we are running the `GetAndCreateOrMergePatch` function as it first GETs the current objects and then runs the
		// mutate() func before sending the PATCH. This way we ensure that we have applied our mutations to the latest version.
		controllerInstallation.Name = existingControllerInstallation.Name
		_, err := controllerutils.GetAndCreateOrMergePatch(ctx, c, controllerInstallation, mutate)
		return err
	}

	// The installation does not exist yet, hence, we set `GenerateName` which will automatically append a random suffix to
	// the name. Unfortunately, the `CreateOrUpdate` function does not support creating an object that does not specify a name
	// but only `GenerateName`, thus, we call `Create` directly.
	controllerInstallation.GenerateName = controllerRegistration.Name + "-"
	_ = mutate()
	return c.Create(ctx, controllerInstallation)
}

// deleteUnneededInstallations takes the list of required names of ControllerRegistrations, and another mapping of
// ControllerRegistration names to existing ControllerInstallations. It deletes every existing ControllerInstallation whose
// referenced ControllerRegistration is not part of the given list of required list.
func deleteUnneededInstallations(
	ctx context.Context,
	logger *logrus.Entry,
	c client.Client,
	wantedControllerRegistrationNames sets.String,
	registrationNameToInstallation map[string]*gardencorev1beta1.ControllerInstallation,
) error {
	for registrationName, installation := range registrationNameToInstallation {
		if !wantedControllerRegistrationNames.Has(registrationName) {
			logger.Infof("Deleting unneeded ControllerInstallation %q", installation.Name)

			if err := c.Delete(ctx, installation); client.IgnoreNotFound(err) != nil {
				return err
			}
		}
	}

	return nil
}

func convertObjToMap(in interface{}) (map[string]interface{}, error) {
	var out map[string]interface{}

	data, err := json.Marshal(in)
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(data, &out); err != nil {
		return nil, err
	}

	return out, nil
}

func seedSelectorMatches(deployment *gardencorev1beta1.ControllerRegistrationDeployment, seedLabels map[string]string) (bool, error) {
	selector := &metav1.LabelSelector{}
	if deployment != nil && deployment.SeedSelector != nil {
		selector = deployment.SeedSelector
	}

	seedSelector, err := metav1.LabelSelectorAsSelector(selector)
	if err != nil {
		return false, fmt.Errorf("label selector conversion failed: %v for seedSelector: %w", *selector, err)
	}

	return seedSelector.Matches(labels.Set(seedLabels)), nil
}

func controllerRegistrationNamesWithMatchingSeedLabelSelector(
	namesInQuestion []string,
	controllerRegistrations map[string]controllerRegistration,
	seedLabels map[string]string,
) (sets.String, error) {
	matchingNames := sets.NewString()

	for _, name := range namesInQuestion {
		controllerRegistration, ok := controllerRegistrations[name]
		if !ok {
			return nil, fmt.Errorf("ControllerRegistration with name %s not found", name)
		}

		matches, err := seedSelectorMatches(controllerRegistration.obj.Spec.Deployment, seedLabels)
		if err != nil {
			return nil, err
		}

		if matches {
			matchingNames.Insert(name)
		}
	}

	return matchingNames, nil
}

func getShoots(ctx context.Context, c client.Reader, seed *gardencorev1beta1.Seed) ([]gardencorev1beta1.Shoot, error) {
	shootList := &gardencorev1beta1.ShootList{}
	if err := c.List(ctx, shootList, client.MatchingFields{core.ShootSeedName: seed.Name}); err != nil {
		return nil, err
	}
	shootListAsItems := gardencorev1beta1helper.ShootItems(*shootList)

	shootList2 := &gardencorev1beta1.ShootList{}
	if err := c.List(ctx, shootList2, client.MatchingFields{core.ShootStatusSeedName: seed.Name}); err != nil {
		return nil, err
	}
	shootListAsItems2 := gardencorev1beta1helper.ShootItems(*shootList2)

	return shootListAsItems.Union(&shootListAsItems2), nil
}
