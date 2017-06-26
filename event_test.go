package main

import "testing"
import "time"
import "github.com/stretchr/testify/assert"

var (
	ipmiSelOutput = `1,Apr-22-2016,09:04:09,Sensor #255,Session Audit,Warning,Invalid Username of Password`
)

func TestParsingIPMISelOutput(t *testing.T) {
	ev, err := newIPMIEvent(ipmiSelOutput)
	assert.Nil(t, err)
	assert.NotNil(t, ev)
	assert.Equal(t, "1", ev.ID)
	assert.Equal(t, "Fri Apr 22 09:04:09 2016", ev.Time.Format(time.ANSIC))
	assert.Equal(t, "Sensor #255", ev.Sensor)
	assert.Equal(t, "Session Audit", ev.Type)
	assert.Equal(t, "Warning", ev.Level)
	assert.Equal(t, "Invalid Username of Password", ev.Message)
}
