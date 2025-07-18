package libhelm

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"path"
	"strings"
	"time"
)

func ValidateHelmRepositoryURL(repoUrl string, client *http.Client) error {
	if repoUrl == "" {
		return errors.New("URL is required")
	}

	if strings.HasPrefix(repoUrl, "oci://") {
		return errors.New("OCI repositories are not supported yet")
	}

	url, err := url.ParseRequestURI(repoUrl)
	if err != nil {
		return fmt.Errorf("invalid helm repository URL '%s': %w", repoUrl, err)
	}

	if !strings.EqualFold(url.Scheme, "http") && !strings.EqualFold(url.Scheme, "https") {
		return fmt.Errorf("invalid helm repository URL '%s'", repoUrl)
	}

	url.Path = path.Join(url.Path, "index.yaml")

	if client == nil {
		client = &http.Client{
			Timeout:   120 * time.Second,
			Transport: http.DefaultTransport,
		}
	}

	response, err := client.Head(url.String())
	if err != nil {
		return fmt.Errorf("%s is not a valid chart repository or cannot be reached: %w", repoUrl, err)
	}

	// Success is indicated with 2xx status codes. 3xx status codes indicate a redirect.
	statusOK := response.StatusCode >= 200 && response.StatusCode < 300
	if !statusOK {
		return fmt.Errorf("%s is not a valid chart repository or cannot be reached", repoUrl)
	}

	return nil
}
