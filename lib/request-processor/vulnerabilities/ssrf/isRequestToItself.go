package ssrf

import (
	"main/context"
	"main/globals"
	"main/helpers"
	"net/url"
)

// isRequestToItself checks if an outbound request is to the same server that's handling the incoming request.
// This includes a special case for HTTP/HTTPS: if the server is running on HTTP (port 80) and makes a request
// to HTTPS (port 443) of the same hostname, or vice versa, it's considered a request to itself.
// This prevents false positives when a server makes requests to itself via different protocols.
func IsRequestToItself(outboundHostname string, outboundPort uint32) bool {
	// Check if trust proxy is enabled
	// If not enabled, we don't consider requests to itself as safe
	server := globals.GetCurrentServer()
	if server == nil || !server.AikidoConfig.TrustProxy {
		return false
	}

	// Get the current server URL from the incoming request
	serverURL := context.GetUrl()
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

	// Special case: HTTP/HTTPS cross-protocol requests to the same hostname
	// If server is on port 80 (HTTP) and outbound is port 443 (HTTPS), or vice versa,
	// consider it as a request to itself
	isServerHTTP := serverPort == 80
	isServerHTTPS := serverPort == 443
	isOutboundHTTP := outboundPort == 80
	isOutboundHTTPS := outboundPort == 443

	// Allow cross-protocol requests between standard HTTP/HTTPS ports
	if (isServerHTTP && isOutboundHTTPS) || (isServerHTTPS && isOutboundHTTP) {
		return true
	}

	return false
}
