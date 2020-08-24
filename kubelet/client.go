// Originally inspired by OpenTelemetry Collector kubeletstats receiver
// https://github.com/open-telemetry/opentelemetry-collector

package kubelet

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/pkg/errors"
)

const svcAcctCACertPath = "/var/run/secrets/kubernetes.io/serviceaccount/ca.crt"
const svcAcctTokenPath = "/var/run/secrets/kubernetes.io/serviceaccount/token"

type Client interface {
	Get(path string) ([]byte, error)
}

func NewClientProvider(endpoint string) (ClientProvider, error) {
	return &saClientProvider{
		endpoint:   endpoint,
		caCertPath: svcAcctCACertPath,
		tokenPath:  svcAcctTokenPath,
	}, nil
}

type ClientProvider interface {
	BuildClient() (Client, error)
}

type saClientProvider struct {
	endpoint   string
	caCertPath string
	tokenPath  string
}

func (p *saClientProvider) BuildClient() (Client, error) {
	rootCAs, err := systemCertPoolPlusPath(p.caCertPath)
	if err != nil {
		return nil, err
	}
	tok, err := ioutil.ReadFile(p.tokenPath)
	if err != nil {
		return nil, errors.WithMessagef(err, "Unable to read token file %s", p.tokenPath)
	}
	tr := defaultTransport()
	tr.TLSClientConfig = &tls.Config{
		RootCAs: rootCAs,
	}
	return defaultTLSClient(p.endpoint, true, rootCAs, nil, tok)
}

func defaultTLSClient(
	endpoint string,
	insecureSkipVerify bool,
	rootCAs *x509.CertPool,
	certificates []tls.Certificate,
	tok []byte,
) (*clientImpl, error) {
	tr := defaultTransport()
	tr.TLSClientConfig = &tls.Config{
		RootCAs:            rootCAs,
		Certificates:       certificates,
		InsecureSkipVerify: insecureSkipVerify,
	}
	if endpoint == "" {
		var err error
		endpoint, err = defaultEndpoint()
		if err != nil {
			return nil, err
		}
	}
	return &clientImpl{
		baseURL:    "https://" + endpoint,
		httpClient: http.Client{Transport: tr},
		tok:        tok,
		logger:     logrus.WithFields(logrus.Fields{}),
	}, nil
}

// This will work if hostNetwork is turned on, in which case the pod has access
// to the node's loopback device.
// https://kubernetes.io/docs/concepts/policy/pod-security-policy/#host-namespaces
func defaultEndpoint() (string, error) {
	hostname, err := os.Hostname()
	if err != nil {
		return "", errors.WithMessage(err, "Unable to get hostname for default endpoint")
	}
	const kubeletPort = "10250"
	return hostname + ":" + kubeletPort, nil
}

func defaultTransport() *http.Transport {
	return http.DefaultTransport.(*http.Transport).Clone()
}

// clientImpl

var _ Client = (*clientImpl)(nil)

type clientImpl struct {
	baseURL    string
	httpClient http.Client
	logger     *logrus.Entry
	tok        []byte
}

func (c *clientImpl) Get(path string) ([]byte, error) {
	req, err := c.buildReq(path)
	if err != nil {
		return nil, err
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() {
		closeErr := resp.Body.Close()
		if closeErr != nil {
			c.logger.Warn("failed to close response body: ", closeErr)
		}
	}()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.WithMessage(err, "failed to read Kubelet response body")
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("kubelet request GET %s failed - %q, response: %q",
			req.URL.String(), resp.Status, string(body))
	}

	return body, nil
}

func (c *clientImpl) buildReq(path string) (*http.Request, error) {
	url := c.baseURL + path
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	if c.tok != nil {
		req.Header.Set("Authorization", fmt.Sprintf("bearer %s", c.tok))
	}
	return req, nil
}
