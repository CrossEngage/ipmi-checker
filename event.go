package main

import (
	"fmt"
	"strings"
	"time"
)

var (
	errCouldNotParseCmdOut = "Unknown command output: %s"
	errCouldNotParseTime   = "Could not parse time: %s"
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
	timeFmt = "Jan-02-2006T03:04:05"
)

func newIPMIEvent(stdout string) (*ipmiEvent, error) {
	parts := strings.Split(stdout, ",")
	if len(parts) != 7 {
		return nil, fmt.Errorf(errCouldNotParseCmdOut, stdout)
	}

	inputTime := parts[1] + "T" + parts[2]
	eventTime, err := time.Parse(timeFmt, inputTime)
	if err != nil {
		return nil, fmt.Errorf(errCouldNotParseTime, inputTime)
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

func (ev ipmiEvent) InfluxDB(checkName, hostname string) string {
	return fmt.Sprintf(
		`%s,host=%s,event_id=%s,error_level=%s,event_type=%s event_date="%s",event_time="%s",error_level="%s",sensor_name="%s",event_type="%s",error_message="%s",state=%d\n`,
		checkName, hostname, ev.ID, ev.Level, ev.Type, ev.Time.Format("Jan-02-2006"), ev.Time.Format("03:04:05"), ev.Level, ev.Sensor, ev.Type, ev.Message, ev.State,
	)
}

func newEmptyIPMIEvent() *ipmiEvent {
	return &ipmiEvent{
		ID:      "0",
		Time:    time.Now(),
		Sensor:  "OK",
		Type:    "OK",
		Level:   "OK",
		Message: "No errors",
		State:   1,
	}
}
