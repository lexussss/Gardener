# Shoot Kubernetes and Operating System Versioning in Gardener

## Motivation

On the one hand-side, Gardener is responsible for managing the Kubernetes and the Operating System (OS) versions of its Shoot clusters.
On the other hand-side, Gardener needs to be configured and updated based on the availability and support of the Kubernetes and Operating System version it provides.
For instance, the Kubernetes community releases **minor** versions roughly every three months and usually maintains **three minor** versions (the current and the last two) with bug fixes and security updates.
Patch releases are done more frequently.

When using the term `Machine image` in the following, we refer to the OS version that comes with the machine image of the node/worker pool of a Gardener Shoot cluster.
As such we are not referring to the `CloudProvider` specific machine image like the [`AMI`](https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/AMIs.html) for AWS.
For more information how Gardener maps machine image versions to `CloudProvider` specific machine images, take a look at the individual gardener extension providers
such as the [provider for AWS](https://github.com/gardener/gardener-extension-provider-aws/blob/master/docs/usage-as-operator.md).

Gardener should be configured accordingly to reflect the "logical state" of a version.
It should be possible to define the Kubernetes or Machine image versions that still receive bug fixes and security patches, and also vice-versa to define the version that are out-of-maintenance and are potentially vulnerable.
Moreover, this allows Gardener to "understand" the current state of a version and act upon it (more information in the following sections).

## Overview

**As a Gardener operator**:

- I can classify a version based on it's logical state (`preview`, `supported`, `deprecated` and `expired` see [Version Classification](#version-classifications)).
- I can define which Machine image and Kubernetes versions are eligible for the auto update of clusters during the maintenance time.
- I can disallow the creation of clusters having a certain version (think of severe security issues).

**As an end-user/Shoot owner of Gardener**:

- I can get information about which Kubernetes and Machine image versions exist and their classification.
- I can determine the time when my Shoot clusters Machine image and Kubernetes version will be forcefully updated to the next patch or minor version (in case the cluster is running a deprecated version with an expiration date).
- I can get this information via API from the `CloudProfile`.

## Version Classifications

Administrators can classify versions into four distinct "logical states": `preview`, `supported`, `deprecated` and `expired`.
The version classification serves as a "point-of-reference" for end-users and also has implications during shoot creation and the maintenance time.

If a version is unclassified, Gardener cannot make those decision based on the "logical state".
Nevertheless, Gardener can operate without version classifications and can be added at any time to the Kubernetes and machine image versions in the `CloudProfile`.

As a best practice, versions usually start with the classification `preview`, then are promoted to `supported`, eventually `deprecated` and finally `expired`.
This information is programmatically available in the `CloudProfiles` of the Garden cluster.

- **preview:** A `preview` version is a new version that has not yet undergone thorough testing, possibly a new release, and needs time to be validated.
Due to its short early age, there is a higher probability of undiscovered issues and is therefore not yet recommended for production usage.
A Shoot does not update (neither `auto-update` or `force-update`) to  a `preview` version during the maintenance time.
Also `preview` versions are not considered for the defaulting to the highest available version when deliberately omitting the patch version during Shoot creation.
Typically, after a fresh release of a new Kubernetes (e.g. v1.17.0) or Machine image version (e.g. coreos-2023.5), the operator tags it as `preview` until he has gained sufficient experience and regards this version to be reliable.
After the operator gained sufficient trust, the version can be manually promoted to `supported`.

- **supported:** A `supported` version is the recommended version for new and existing Shoot clusters. New Shoot clusters should use and existing clusters should update to this version.
Typically for Kubernetes versions, the latest Kubernetes patch versions of the actual (if not still in `preview`) and the last 3 minor Kubernetes versions are maintained by the community. An operator could define these versions as being `supported` (e.g. v1.16.1, v1.15.4, v1.14.9 and v1.13.12).

- **deprecated:** A `deprecated` version is a version that approaches the end of its lifecycle and can contain issues which are probably resolved in a supported version.
New Shoots should not use this version any more.
Existing Shoots will be updated to a newer version if `auto-update` is enabled (`.spec.maintenance.autoUpdate.kubernetesVersion` for Kubernetes version `auto-update`, or `.spec.maintenance.autoUpdate.machineImageVersion` for machine machine image version `auto-update`).
Using automatic upgrades, however, does not guarantee that a Shoot runs a non-deprecated version, as the latest version (overall or of the minor version) can be deprecated as well.
Deprecated versions **should** have an expiration date set for eventual expiration.

- **expired:** An `expired` versions has an expiration date (based on the [Golang time package](https://golang.org/src/time/time.go)) in the past.
 New clusters with that version cannot be created and existing clusters are forcefully migrated to a higher version during the maintenance time.

Below is an example how the relevant section of the `CloudProfile` might look like:

``` yaml
apiVersion: core.gardener.cloud/v1beta1
kind: CloudProfile
metadata:
  name: alicloud
spec:
  kubernetes:
    versions:
      - classification: supported
        version: 1.17.1
      - classification: deprecated
        expirationDate: "2020-07-24T16:13:26Z"
        version: 1.17.0
      - classification: preview
        version: 1.16.6
      - classification: supported
        version: 1.16.5
      - classification: deprecated
        expirationDate: "2020-04-25T09:30:40Z"
        version: 1.16.4
      - classification: supported
        version: 1.15.7
      - classification: deprecated
        expirationDate: "2020-06-09T14:01:39Z"
        version: 1.15.6
```

## Version Requirements (Kubernetes and Machine image)

The Gardener API server enforces the following requirements for versions:

### Deletion of a version

- A version that is in use by a Shoot cannot be deleted from the `CloudProfile`.

### Adding a version

- A version must not have an expiration date in the past.
- There can be only one `supported` version per minor version.
- The latest Kubernetes version cannot have an expiration date.
- The latest version for a machine image can have an expiration date. [*]

<sub>[*] Useful for cases in which support for given machine image needs to be deprecated and removed (for example the machine image reaches end of life).</sub>

## Forceful migration of expired versions

If a Shoot is running a version after its expiration date has passed, it will be forcefully migrated during its maintenance time.
This happens **even if the owner has opted out of automatic cluster updates!**

For **Machine images**, the Shoots worker pools will be updated to the latest `non-preview` version of the pools respective image.

For **Kubernetes versions**, the forceful update picks the latest `non-preview` patch version of the current minor version.

If the cluster is already on the latest patch version and the latest patch version is also expired,
it will continue with the latest patch version of the **next consecutive minor Kubernetes version**, so **it will result in an
update of a minor Kubernetes version!**

Please note, that multiple consecutive minor version upgrades are possible.
This can occur if the Shoot is updated to a version that in turn is also `expired`.
In this case, the version is again upgraded in the **next** maintenance time.

**Depending on the circumstances described above, it can happen that the cluster receives multiple consecutive minor Kubernetes version updates!**

Kubernetes "minor version jumps" are not allowed - meaning to skip the update to the consecutive minor version and directly update to any version after that.
For instance, the version `1.10.x` can only update to a version `1.11.x`, not to `1.12.x` or any other version.
This is because Kubernetes does not guarantee upgradeability in this case, leading to possibly broken Shoot clusters.
The administrator has to set up the `CloudProfile` in such a way, that consecutive Kubernetes minor versions are available.
Otherwise, Shoot clusters will fail to upgrade during the maintenance time.

Consider the `CloudProfile` below with a Shoot using the Kubernetes version `1.10.12`.
Even though the version is `expired`, due to missing `1.11.x` versions, the Gardener Controller Manager cannot upgrade the Shoot's Kubernetes version.

```yaml
spec:
  kubernetes:
    versions:
    - version: 1.12.8
    - version: 1.12.7
    - version: 1.10.12
      expirationDate: "<expiration date in the past>"
```

The `CloudProfile` must specify versions `1.11.x` of the **consecutive** minor version.
Configuring the `CloudProfile` in such a way, the Shoot's Kubernetes version will be upgraded to version `1.11.10` in the next maintenance time.

```yaml
spec:
  kubernetes:
    versions:
    - version: 1.12.8
    - version: 1.11.10
    - version: 1.11.09
    - version: 1.10.12
      expirationDate: "<expiration date in the past>"
```

## Related Documentation

You might want to read about the [Shoot Updates and Upgrades](shoot_updates.md) procedures to get to know the effects of such operations.
