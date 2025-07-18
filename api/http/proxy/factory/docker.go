package factory

import (
	"io"
	"net/http"
	"strings"

	portainer "github.com/portainer/portainer/api"
	"github.com/portainer/portainer/api/crypto"
	"github.com/portainer/portainer/api/http/proxy/factory/docker"
	"github.com/portainer/portainer/api/url"
	httperror "github.com/portainer/portainer/pkg/libhttp/error"

	"github.com/rs/zerolog/log"
)

func (factory *ProxyFactory) newDockerProxy(endpoint *portainer.Endpoint) (http.Handler, error) {
	if strings.HasPrefix(endpoint.URL, "unix://") || strings.HasPrefix(endpoint.URL, "npipe://") {
		return factory.newDockerLocalProxy(endpoint)
	}

	return factory.newDockerHTTPProxy(endpoint)
}

func (factory *ProxyFactory) newDockerLocalProxy(endpoint *portainer.Endpoint) (http.Handler, error) {
	endpointURL, err := url.ParseURL(endpoint.URL)
	if err != nil {
		return nil, err
	}

	return factory.newOSBasedLocalProxy(endpointURL.Path, endpoint)
}

func (factory *ProxyFactory) newDockerHTTPProxy(endpoint *portainer.Endpoint) (http.Handler, error) {
	rawURL := endpoint.URL
	if endpoint.Type == portainer.EdgeAgentOnDockerEnvironment {
		tunnelAddr, err := factory.reverseTunnelService.TunnelAddr(endpoint)
		if err != nil {
			return nil, err
		}
		rawURL = "http://" + tunnelAddr
	}

	endpointURL, err := url.ParseURL(rawURL)
	if err != nil {
		return nil, err
	}

	endpointURL.Scheme = "http"
	httpTransport := &http.Transport{}

	if endpoint.TLSConfig.TLS || endpoint.TLSConfig.TLSSkipVerify {
		config, err := crypto.CreateTLSConfigurationFromDisk(endpoint.TLSConfig.TLSCACertPath, endpoint.TLSConfig.TLSCertPath, endpoint.TLSConfig.TLSKeyPath, endpoint.TLSConfig.TLSSkipVerify)
		if err != nil {
			return nil, err
		}

		httpTransport.TLSClientConfig = config
		endpointURL.Scheme = "https"
	}

	transportParameters := &docker.TransportParameters{
		Endpoint:             endpoint,
		DataStore:            factory.dataStore,
		ReverseTunnelService: factory.reverseTunnelService,
		SignatureService:     factory.signatureService,
		DockerClientFactory:  factory.dockerClientFactory,
	}

	dockerTransport, err := docker.NewTransport(transportParameters, httpTransport, factory.gitService, factory.snapshotService)
	if err != nil {
		return nil, err
	}

	proxy := NewSingleHostReverseProxyWithHostHeader(endpointURL)
	proxy.Transport = dockerTransport
	return proxy, nil
}

type dockerLocalProxy struct {
	transport *docker.Transport
}

// ServeHTTP is the http.Handler interface implementation
// for a local (Unix socket or Windows named pipe) Docker proxy.
func (proxy *dockerLocalProxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Force URL/domain to http/unixsocket to be able to
	// use http.transport RoundTrip to do the requests via the socket
	r.URL.Scheme = "http"
	r.URL.Host = "unixsocket"

	res, err := proxy.transport.ProxyDockerRequest(r)
	if err != nil {
		code := http.StatusInternalServerError
		if res != nil && res.StatusCode != 0 {
			code = res.StatusCode
		}

		httperror.WriteError(w, code, "Unable to proxy the request via the Docker socket", err)
		return
	}
	defer res.Body.Close()

	for k, vv := range res.Header {
		for _, v := range vv {
			w.Header().Add(k, v)
		}
	}

	w.WriteHeader(res.StatusCode)

	if _, err := io.Copy(w, res.Body); err != nil {
		log.Debug().Err(err).Msg("proxy error")
	}
}
