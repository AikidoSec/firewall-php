package ssrf

import (
	"main/helpers"
	"main/log"
	"net/url"
	"strconv"
	"strings"
)

func getVariants(userInput string) []string {
	variants := []string{userInput, "http://" + userInput, "https://" + userInput}
	decodedUserInput, err := url.QueryUnescape(userInput)
	if err == nil && decodedUserInput != userInput {
		variants = append(variants, decodedUserInput, "http://"+decodedUserInput, "https://"+decodedUserInput)
	}
	return variants
}

func findHostnameInUserInput(userInput string, hostname string, port uint32) bool {
	log.Debugf("findHostnameInUserInput: userInput: %s, hostname: %s, port: %d", userInput, hostname, port)
	if len(userInput) <= 1 {
		return false
	}
	// if hostname contains : we need to add the [ and ] to the hostname (ipv6)
	if strings.Contains(hostname, ":") {
		hostname = "[" + hostname + "]"
	}

	hostnameURL := helpers.TryParseURL("http://" + hostname + ":" + strconv.Itoa(int(port)))
	if hostnameURL == nil {
		return false
	}

	userInput = helpers.ExtractResourceOrOriginal(userInput)
	userInput = helpers.NormalizeRawUrl(userInput)

	variants := getVariants(userInput)

	for _, variant := range variants {
		userInputURL := helpers.TryParseURL(variant)
		if userInputURL == nil {
			continue
		}

		// https://datatracker.ietf.org/doc/html/rfc3986#section-3.2.2
		// "The host subcomponent is case-insensitive."
		if userInputURL != nil && strings.EqualFold(userInputURL.Hostname(), hostnameURL.Hostname()) {
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
