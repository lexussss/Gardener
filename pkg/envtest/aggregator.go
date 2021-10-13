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

package envtest

import (
	"fmt"
	"os"
	"path/filepath"

	"sigs.k8s.io/controller-runtime/pkg/envtest"

	"github.com/gardener/gardener/pkg/utils/secrets"
)

// AggregatorConfig is able to configure the kube-aggregator (kube-apiserver aggregation layer)
// by provisioning the front-proxy certs and setting the corresponding flags on the kube-apiserver.
type AggregatorConfig struct {
	certDir string
}

// ConfigureAPIServerArgs generates the needed certs, writes them to the given directory and configures args
// to point to the generated certs.
func (a AggregatorConfig) ConfigureAPIServerArgs(certDir string, args *envtest.Arguments) error {
	a.certDir = certDir

	if err := a.generateCerts(); err != nil {
		return err
	}

	args.
		Set("requestheader-extra-headers-prefix", "X-Remote-Extra-").
		Set("requestheader-group-headers", "X-Remote-Group").
		Set("requestheader-username-headers", "X-Remote-User").
		Set("requestheader-client-ca-file", a.caCrtPath()).
		Set("proxy-client-cert-file", a.clientCrtPath()).
		Set("proxy-client-key-file", a.clientKeyPath())

	return nil
}

func (a AggregatorConfig) caCrtPath() string {
	return filepath.Join(a.certDir, "proxy-ca.crt")
}

func (a AggregatorConfig) clientCrtPath() string {
	return filepath.Join(a.certDir, "proxy-client.crt")
}

func (a AggregatorConfig) clientKeyPath() string {
	return filepath.Join(a.certDir, "proxy-client.key")
}

func (a AggregatorConfig) generateCerts() error {
	caConfig := &secrets.CertificateSecretConfig{
		Name:       "front-proxy",
		CommonName: "front-proxy",
		CertType:   secrets.CACert,
	}

	ca, err := caConfig.GenerateCertificate()
	if err != nil {
		return err
	}
	if err := os.WriteFile(a.caCrtPath(), ca.CertificatePEM, 0640); err != nil {
		return fmt.Errorf("unable to save the proxy client CA certificate to %s: %w", a.caCrtPath(), err)
	}

	clientConfig := &secrets.CertificateSecretConfig{
		Name:       "front-proxy",
		CommonName: "front-proxy",
		CertType:   secrets.ClientCert,
		SigningCA:  ca,
	}

	clientCert, err := clientConfig.GenerateCertificate()
	if err != nil {
		return err
	}
	if err := os.WriteFile(a.clientCrtPath(), clientCert.CertificatePEM, 0640); err != nil {
		return fmt.Errorf("unable to save the proxy client certificate to %s: %w", a.clientCrtPath(), err)
	}
	if err := os.WriteFile(a.clientKeyPath(), clientCert.PrivateKeyPEM, 0640); err != nil {
		return fmt.Errorf("unable to save the proxy client key to %s: %w", a.clientKeyPath(), err)
	}

	return nil
}
