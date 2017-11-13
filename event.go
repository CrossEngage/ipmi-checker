package main

import (
	"fmt"
	"strings"
	"time"
)

var (
	errCouldNotParseCmdOut = "Unknown command output: %s"
	errCouldNotParseTime   = "Could not parse time: %s: %s"
)

type ipmiEvent struct {
	ID      string
	Time    time.Time
	Sensor  string
	Type    string
	Level   string
	Message string
	State   int
}

const (
	timeFmt = "Jan-02-2006T15:04:05"
)

func newIPMIEvent(stdOutLine string) (*ipmiEvent, error) {
	parts := strings.Split(stdOutLine, ",")
	if len(parts) != 7 {
		return nil, fmt.Errorf(errCouldNotParseCmdOut, stdOutLine)
	}

	inputTime := parts[1] + "T" + parts[2]
	eventTime, err := time.Parse(timeFmt, inputTime)
	if err != nil {
		return nil, fmt.Errorf(errCouldNotParseTime, inputTime, err)
	}

	event := &ipmiEvent{
		ID:      parts[0],
		Time:    eventTime,
		Sensor:  parts[3],
		Type:    parts[4],
		Level:   parts[5],
		Message: parts[6],
		State:   1,
	}

	return event, nil
}

func spacesToUnderscore(str string) string {
	return strings.Replace(str, " ", "_", -1)
}

func (ev ipmiEvent) InfluxDB(checkName, hostname string) string {
	return fmt.Sprintf(
		`%s,host=%s,event_type=%s,error_level=%s,sensor_name=%s event_id=%s,error_message="%s",state=%d %d`,
		checkName, hostname, spacesToUnderscore(ev.Type), spacesToUnderscore(ev.Level), spacesToUnderscore(ev.Sensor),
		ev.ID, ev.Message, ev.State, ev.Time.UnixNano())
}

func newEmptyIPMIEvent() *ipmiEvent {
	return &ipmiEvent{
		ID:      "0",
		Time:    time.Now(),
		Sensor:  "OK",
		Type:    "OK",
		Level:   "OK",
		Message: "No errors",
		State:   0,
	}
}
