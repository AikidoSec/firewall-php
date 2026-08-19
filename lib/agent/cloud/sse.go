package cloud

import (
	"bufio"
	"context"
	"encoding/json"
	. "main/aikido_types"
	"main/config"
	"main/constants"
	"main/log"
	"main/utils"
	"math/rand"
	"net"
	"net/http"
	"net/url"
	"slices"
	"time"
)

/*
	Listens for cloud config changes pushed over a Server-Sent Events stream, so
	config updates are applied as soon as they happen instead of waiting for the
	next config polling interval. Polling stays active as a fallback.
*/

/*
Dedicated client for the config stream: compression must stay off so the cloud
does not buffer the stream, and no client timeout can be set because a healthy
connection is long lived. Stalled connections are detected via a read timeout.
Every phase before the first byte is bounded as well, so a connection that never
completes cannot hold up the reconnect loop.
*/
var configStreamClient = &http.Client{
	Transport: &http.Transport{
		DisableCompression: true,
		DialContext: (&net.Dialer{
			Timeout: constants.ConfigStreamDialTimeoutInMs * time.Millisecond,
		}).DialContext,
		TLSHandshakeTimeout:   constants.ConfigStreamTLSHandshakeTimeoutInMs * time.Millisecond,
		ResponseHeaderTimeout: constants.ConfigStreamReadTimeoutInMs * time.Millisecond,
	},
}

/*
Opens the config stream and feeds the response to the parser until the connection ends.
Returns the HTTP status code of the response, or 0 if no response was received.
*/
func connectToConfigStream(ctx context.Context, server *ServerData, readTimeout time.Duration, onEvent func(sseEvent)) int {
	token := config.GetToken(server)
	if token == "" {
		return 0
	}

	streamEndpoint, err := url.JoinPath(server.AikidoConfig.ConfigEndpoint, constants.ConfigStreamAPI)
	if err != nil {
		log.Warnf(server.Logger, "Failed to build config stream endpoint: %v", err)
		return 0
	}

	// Own cancellable context, so the read timeout can abort a stalled connection
	readCtx, cancelRead := context.WithCancel(ctx)
	defer cancelRead()

	req, err := http.NewRequestWithContext(readCtx, constants.ConfigStreamMethod, streamEndpoint, nil)
	if err != nil {
		log.Warnf(server.Logger, "Failed to create config stream request: %v", err)
		return 0
	}
	req.Header.Set("Authorization", token)
	req.Header.Set("Accept", "text/event-stream")
	req.Header.Set("Cache-Control", "no-cache")
	req.Header.Set("Accept-Encoding", "identity")
	req.Header.Set("X-Agent-Platform", "php")
	req.Header.Set("X-Agent-Version", constants.Version)

	log.Debugf(server.Logger, "[%s] Connecting to config stream %s", utils.AnonymizeToken(token), streamEndpoint)

	resp, err := configStreamClient.Do(req)
	if err != nil {
		// Only logged on debug: config polling already reports an unreachable cloud
		log.Debugf(server.Logger, "Error in connecting to config stream: %v", err)
		return 0
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Debugf(server.Logger, "Config stream responded with status %s", resp.Status)
		return resp.StatusCode
	}

	log.Debugf(server.Logger, "Connected to config stream!")

	readTimeoutTimer := time.AfterFunc(readTimeout, func() {
		log.Debugf(server.Logger, "Config stream read timeout, closing connection!")
		cancelRead()
	})
	defer readTimeoutTimer.Stop()

	parser := &sseParser{}
	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		readTimeoutTimer.Reset(readTimeout)

		if event, ok := parser.feedLine(scanner.Text()); ok {
			onEvent(event)
		}
	}

	if err := scanner.Err(); err != nil {
		log.Debugf(server.Logger, "Config stream read ended: %v", err)
	}
	return resp.StatusCode
}

func handleConfigStreamEvent(server *ServerData, event sseEvent) {
	log.Debugf(server.Logger, "Got config stream event: %s", event.name)

	if event.name != constants.ConfigUpdatedEvent {
		return
	}

	cloudConfigUpdatedAt := CloudConfigUpdatedAt{}
	if err := json.Unmarshal([]byte(event.data), &cloudConfigUpdatedAt); err != nil {
		log.Debugf(server.Logger, "Config stream event has invalid payload: %s", event.data)
		return
	}

	if !WasConfigUpdated(server, cloudConfigUpdatedAt.ConfigUpdatedAt) {
		return
	}

	if configStreamEventArrivedTooFast(server) {
		log.Debug(server.Logger, "Ignoring config stream event during refresh throttle")
		return
	}

	log.Infof(server.Logger, "Config stream reported a config update, fetching new config!")
	FetchAndStoreCloudConfig(server)
}

func configStreamEventArrivedTooFast(server *ServerData) bool {
	server.ConfigStreamRefreshMutex.Lock()
	defer server.ConfigStreamRefreshMutex.Unlock()

	now := time.Now()
	if now.Sub(server.ConfigStreamLastRefreshStart) < 9*time.Second {
		return true
	}

	server.ConfigStreamLastRefreshStart = now
	return false
}

/*
Keeps the config stream connected, reconnecting with an exponential backoff and
jitter. A connection that stayed up long enough resets the backoff, while an
authentication failure stops the routine for good.
*/
func configStreamRoutine(ctx context.Context, server *ServerData, initialReconnectInMs int, readTimeout time.Duration) {
	reconnectInMs := initialReconnectInMs

	for {
		connectedAt := utils.GetTime()

		statusCode := connectToConfigStream(ctx, server, readTimeout, func(event sseEvent) {
			handleConfigStreamEvent(server, event)
		})

		if ctx.Err() != nil {
			log.Debug(server.Logger, "Config stream routine stopped!")
			return
		}

		if statusCode == http.StatusUnauthorized || statusCode == http.StatusForbidden {
			log.Warnf(server.Logger, "Config stream rejected with status %d, stopping!", statusCode)
			return
		}

		if utils.GetTime()-connectedAt >= constants.ConfigStreamStableConnectionInMs {
			reconnectInMs = initialReconnectInMs
		}

		jitterInMs := rand.Intn(reconnectInMs/2 + 1)
		delay := time.Duration(reconnectInMs+jitterInMs) * time.Millisecond

		log.Debugf(server.Logger, "Scheduling config stream reconnect in %v", delay)

		reconnectInMs = min(reconnectInMs*2, constants.ConfigStreamMaxReconnectInMs)

		select {
		case <-ctx.Done():
			log.Debug(server.Logger, "Config stream routine stopped!")
			return
		case <-time.After(delay):
		}
	}
}

/*
Realtime updates are enabled either locally, via AIKIDO_FEATURE_SSE, or remotely, when
the cloud lists the feature for this service, which is how the feature gets rolled out.
*/
func isRealtimeEnabled(server *ServerData) bool {
	if config.GetSse(server) {
		return true
	}

	server.CloudConfigMutex.Lock()
	defer server.CloudConfigMutex.Unlock()

	return slices.Contains(server.CloudConfig.EnabledFeatures, constants.RealtimeUpdatesFeature)
}

func StartConfigStreamRoutine(server *ServerData) {
	if !isRealtimeEnabled(server) {
		return
	}

	if config.GetToken(server) == "" {
		log.Info(server.Logger, "No token set, not listening for config updates!")
		return
	}

	ctx, cancel := context.WithCancel(context.Background())

	/*
		Bridge the stop channel to the context, so stopping the routine also aborts
		an in-flight read instead of leaving the connection open until the cloud closes it.
	*/
	go func() {
		select {
		case <-server.PollingData.ConfigStreamRoutineChannel:
		case <-ctx.Done():
		}
		cancel()
	}()

	go func() {
		defer cancel()
		configStreamRoutine(ctx, server, constants.ConfigStreamInitialReconnectInMs,
			constants.ConfigStreamReadTimeoutInMs*time.Millisecond)
	}()
}

/*
Stopped unconditionally, like the other polling routines: signalling a routine that
was never started is harmless, while skipping the signal would leave a running
routine behind if the feature flag ever read differently than it did at start.
*/
func StopConfigStreamRoutine(server *ServerData) {
	utils.StopPollingRoutine(server.PollingData.ConfigStreamRoutineChannel)
}
