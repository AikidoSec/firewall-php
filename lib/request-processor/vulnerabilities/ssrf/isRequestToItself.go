package ssrf

import (
	"main/context"
	"main/helpers"
	"main/instance"
	"net/url"
)

// isRequestToItself checks if an outbound request is to the same server that's handling the incoming request.
// This includes a special case for HTTP/HTTPS: if the server is running on HTTP (port 80) and makes a request
// to HTTPS (port 443) of the same hostname, or vice versa, it's considered a request to itself.
// This prevents false positives when a server makes requests to itself via different protocols.
func IsRequestToItself(instance *instance.RequestProcessorInstance, outboundHostname string, outboundPort uint32) bool {
	if instance == nil {
		return false
	}

	server := instance.GetCurrentServer()

	// Check if trust proxy is enabled
	// If not enabled, we don't consider requests to itself as safe
	if server != nil && !server.AikidoConfig.TrustProxy {
		return false
	}

	// Get the current server URL from the incoming request
	serverURL := context.GetUrl(instance)
	if serverURL == "" {
		return false
	}

	// Parse the server URL to extract hostname and port
	parsedURL, err := url.Parse(serverURL)
	if err != nil {
		return false
	}

	serverHostname := parsedURL.Hostname()
	serverPort := helpers.GetPortFromURL(parsedURL)

	// If hostnames don't match, it's not a request to itself
	if serverHostname != outboundHostname {
		return false
	}

	// If ports match exactly, it's a request to itself
	if serverPort == outboundPort {
		return true
	}

	// Special case for HTTP/HTTPS ports
	// In production, the app will be served on port 80 and 443
	if serverPort == 80 && outboundPort == 443 {
		return true
	}
	if serverPort == 443 && outboundPort == 80 {
		return true
	}

	return false
}
