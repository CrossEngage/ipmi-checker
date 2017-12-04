//go:generate bash ./g_version.sh
package main

import (
	"fmt"
	"log"
	"log/syslog"
	"os"
	"os/exec"
	"path"
	"strings"
	"syscall"

	"github.com/sirupsen/logrus"
	logSys "github.com/sirupsen/logrus/hooks/syslog"
	"gopkg.in/alecthomas/kingpin.v2"
)

var (
	appName    = path.Base(os.Args[0])
	app        = kingpin.New(appName, "A command-line checker for IPMI checks using ipmi-sel, by CrossEngage")
	checkName  = app.Flag("name", "check name").Default(appName).String()
	debug      = app.Flag("debug", "if set, enables debug logging").Default("false").Bool()
	syslogHook = app.Flag("syslog", "if set, enables logging to syslog").Default("false").Bool()
	deadman    = app.Flag("deadman", "if set, this program will always print something").Default("false").Bool()
	useSudo    = app.Flag("sudo", "if set, this program will use sudo to run ipmi-sel").Default("true").Bool()
	eventTime  = app.Flag("event-time", "if set, the event time will be used instead of current time (Use --no-event-time to set to false)").Default("true").Bool()
	ipmiSel    = app.Flag("ipmi-sel", "Path of ipmi-sel").Default("/usr/sbin/ipmi-sel").String()
)

func main() {
	app.Version(version)
	kingpin.MustParse(app.Parse(os.Args[1:]))

	if *debug {
		logrus.SetLevel(logrus.DebugLevel)
	}

	log.SetOutput(os.Stderr)
	logrus.SetFormatter(&logrus.TextFormatter{})

	if *syslogHook {
		hook, err := logSys.NewSyslogHook("", "", syslog.LOG_INFO, appName)
		if err != nil {
			logrus.Error("Unable to connect to local syslog daemon")
		} else {
			logrus.AddHook(hook)
		}
		log.SetFlags(log.Lshortfile)
	}

	if uid := syscall.Getuid(); !*useSudo && uid != 0 {
		logrus.Fatal("Needs to be executed with UID 0, but it is running with UID ", uid)
	}

	hostname, err := os.Hostname()
	if err != nil {
		logrus.Fatal(err)
	}

	args := []string{"--debug", "--output-event-state", "--comma-separated-output", "--no-header-output"}
	cmd := exec.Command(*ipmiSel, args...)
	if *useSudo {
		args = append([]string{*ipmiSel}, args...)
		cmd = exec.Command("sudo", args...)
	}

	out, err := cmd.Output()
	outStr := string(out)
	if err != nil {
		logrus.Errorf("Got error running `%s`: `%s` : `%s`", *ipmiSel, err, outStr) // purposely do not quit
	}

	lines := strings.Split(strings.TrimSpace(outStr), "\n")
	logrus.Debugf("%#v", lines)

	if len(lines) >= 0 && lines[0] != "" {
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if len(line) == 0 {
				logrus.Warnf("Got empty line from output of ipmi-sel.")
				continue
			}
			logrus.Debugf("Line: `%s`", line)
			ev, err := newIPMIEvent(line)
			logrus.Debugf("Event: %#v", ev)
			if err != nil {
				logrus.Errorf("Could not parse line `%s`, err: `%s`", line, err)
				continue
			}
			fmt.Println(ev.InfluxDB(*checkName, hostname, *eventTime))
		}
	} else if *deadman {
		ev := newEmptyIPMIEvent()
		fmt.Println(ev.InfluxDB(*checkName, hostname, *eventTime))
	}
}
