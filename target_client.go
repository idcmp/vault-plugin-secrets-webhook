package relay

import (
	"github.com/hashicorp/errwrap"
	"github.com/hashicorp/go-cleanhttp"
	"net/http"
	"strings"
)

func sendRequest(url string, body string, followRedirects bool) error {
	client := cleanhttp.DefaultClient()

	if !followRedirects {
		client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		}
	}

	req, err := http.NewRequest("POST", url, strings.NewReader(body))
	if err != nil {
		return errwrap.Wrapf("error making request: {{err}}", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		return errwrap.Wrapf("error making request: {{err}}", err)
	}
	resp.Body.Close()

	return nil
}
