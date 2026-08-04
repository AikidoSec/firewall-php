package cloud

import (
	"strings"
)

/*
	Parser for the Server-Sent Events wire format, used to read the cloud config
	stream. Lines are fed in as they are read from the connection and a complete
	event is returned once the blank line that terminates it arrives.
*/

type sseEvent struct {
	name string
	data string
}

type sseParser struct {
	eventName string
	dataLines []string
}

// Returns the event and true once the current event is terminated by a blank line
func (p *sseParser) feedLine(line string) (sseEvent, bool) {
	switch {
	case strings.HasPrefix(line, "event:"):
		p.eventName = strings.TrimSpace(strings.TrimPrefix(line, "event:"))
	case strings.HasPrefix(line, "data:"):
		p.dataLines = append(p.dataLines, strings.TrimSpace(strings.TrimPrefix(line, "data:")))
	case line == "":
		event := sseEvent{name: p.eventName, data: strings.Join(p.dataLines, "\n")}
		p.eventName = ""
		p.dataLines = nil
		return event, true
	}
	// Comments, which servers send as keep-alives, and any other field are ignored
	return sseEvent{}, false
}
