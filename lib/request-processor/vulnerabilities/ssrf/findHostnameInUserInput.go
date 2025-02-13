package ssrf

import (
	"main/helpers"
)

func findHostnameInUserInput(userInput string, hostname string, port uint32) bool {

	if len(userInput) <= 1 {
		return false
	}

	hostnameURL := helpers.TryParseURL("http://" + hostname)
	if hostnameURL == nil {
		return false
	}

	variants := []string{userInput, "http://" + userInput, "https://" + userInput}

	for _, variant := range variants {
		userInputURL := helpers.TryParseURL(variant)
		if userInputURL != nil && userInputURL.Hostname() == hostnameURL.Hostname() {
			userPort := helpers.GetPortFromURL(userInputURL)

			if port == 0 {
				/* If we couldn't extract the port from the original URL (maybe the scheme is not http or https, or the port is not present in the URL)
				or the port was not provided with the outgoing request function, we just skip the comparison of the ports. */
				return true
			}

			if userPort == port {
				return true
			}
		}
	}

	return false
}
