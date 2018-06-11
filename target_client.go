package relay

import (
	"bytes"
	"github.com/hashicorp/errwrap"
	"github.com/hashicorp/go-cleanhttp"
	"io/ioutil"
	"net/http"
	"time"
)

func sendRequest(url string, body []byte, followRedirects bool, timeout time.Duration) ([]byte, error) {
	client := cleanhttp.DefaultClient()

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

	resp.Body.Close()

	return responseBody, nil

}
