//go:generate bash ./g_version.sh
package main

import (
	"bytes"
	"fmt"
	"log/syslog"
	"os"
	"os/exec"
	"path"
	"strings"

	"log"

	"gopkg.in/alecthomas/kingpin.v1"
)

var (
	appName   = path.Base(os.Args[0])
	app       = kingpin.New(appName, "A command-line checker for IPMI checks using ipmi-sel, by CrossEngage")
	checkName = app.Flag("name", "check name").Default(appName).String()
	ipmiSel   = app.Flag("ipmi-sel", "Path of ipmi-sel").Default("/usr/sbin/ipmi-sel").String()
)

func main() {
	app.Version(version)
	kingpin.MustParse(app.Parse(os.Args[1:]))

	hostname, err := os.Hostname()
	if err != nil {
		log.Fatal(err)
	}

	slog, err := syslog.New(syslog.LOG_NOTICE|syslog.LOG_DAEMON, appName)
	if err != nil {
		log.Fatal(err)
	}
	log.SetOutput(slog)

	cmd := exec.Command(*ipmiSel, "--output-event-state", "--comma-separated-output", "--no-header-output")
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		log.Fatalf("Failed running `%s` with error `%s`\n", *ipmiSel, err)
	}

	outStr, errStr := string(stdout.Bytes()), string(stderr.Bytes())
	slog.Debug(fmt.Sprintf("%s: stdout `%s`, stderr `%s`", *ipmiSel, outStr, errStr))

	outStr = strings.TrimSpace(outStr)
	lines := strings.Split(outStr, "\n")
	if len(lines) >= 0 {
		for _, line := range lines {
			ev, err := newIPMIEvent(line)
			if err != nil {
				slog.Err(fmt.Sprintf("Could not parse `%s` stdout: `%s`", line, outStr))
			}
			fmt.Println(ev.InfluxDB(*checkName, hostname))
		}
	} else {
		ev := newEmptyIPMIEvent()
		fmt.Println(ev.InfluxDB(*checkName, hostname))
	}
}
