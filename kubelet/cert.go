// Originally inspired by OpenTelemetry Collector kubeletstats receiver
// https://github.com/open-telemetry/opentelemetry-collector

package kubelet

import (
	"crypto/x509"
	"io/ioutil"

	"github.com/pkg/errors"
)

func systemCertPoolPlusPath(certPath string) (*x509.CertPool, error) {
	sysCerts, err := x509.SystemCertPool()
	if err != nil {
		return nil, errors.WithMessage(err, "Could not load system x509 cert pool")
	}
	return certPoolPlusPath(sysCerts, certPath)
}

func certPoolPlusPath(certPool *x509.CertPool, certPath string) (*x509.CertPool, error) {
	certBytes, err := ioutil.ReadFile(certPath)
	if err != nil {
		return nil, errors.Wrapf(err, "Cert path %s could not be read", certPath)
	}
	ok := certPool.AppendCertsFromPEM(certBytes)
	if !ok {
		return nil, errors.New("AppendCertsFromPEM failed")
	}
	return certPool, nil
}
