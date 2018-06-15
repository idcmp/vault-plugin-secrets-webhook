package webhook

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"time"

	"crypto/tls"
	"crypto/x509"
	"fmt"
	"github.com/hashicorp/errwrap"
)

func sendRequest(url string, body []byte, followRedirects bool, timeout time.Duration, cert []byte) ([]byte, error) {

	var tlsConfig *tls.Config

	if len(cert) != 0 {
		rootCAs := x509.NewCertPool()

		if !rootCAs.AppendCertsFromPEM(cert) {
			return nil, fmt.Errorf("couldn't add target specific CA cert when trying to reach %q", url)
		}

		tlsConfig = &tls.Config{
			InsecureSkipVerify: false,
			RootCAs:            rootCAs,
		}
	} else {
		rootCAs, _ := x509.SystemCertPool()

		tlsConfig = &tls.Config{
			InsecureSkipVerify: false,
			RootCAs:            rootCAs,
		}
	}

	tr := &http.Transport{TLSClientConfig: tlsConfig}
	client := &http.Client{Transport: tr}

	client.Timeout = timeout

	if !followRedirects {
		client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		}
	}

	req, err := http.NewRequest("POST", url, bytes.NewReader(body))
	if err != nil {
		return nil, errwrap.Wrapf("error making request: {{err}}", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, errwrap.Wrapf("error making request: {{err}}", err)
	}

	responseBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, errwrap.Wrapf("error reading response: {{err}}", err)
	}
	resp.Body.Close()

	return responseBody, nil

}
