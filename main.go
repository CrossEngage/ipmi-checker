//go:generate bash ./g_version.sh
package main

import (
	"fmt"
	"io"
	"log/syslog"
	"os"
	"os/exec"
	"path"
	"strings"
	"syscall"

	"log"

	"gopkg.in/alecthomas/kingpin.v1"
)

var (
	appName   = path.Base(os.Args[0])
	app       = kingpin.New(appName, "A command-line checker for IPMI checks using ipmi-sel, by CrossEngage")
	checkName = app.Flag("name", "check name").Default(appName).String()
	debug     = app.Flag("debug", "if set, enables debug log on stderr").Default("false").Bool()
	deadman   = app.Flag("deadman", "if set, this program will always print something").Default("false").Bool()
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

	if *debug {
		log.SetOutput(io.MultiWriter(slog, os.Stderr))
		log.Printf("uid=%d euid=%d gid=%d egid=%d\n", syscall.Getuid(), syscall.Geteuid(), syscall.Getgid(), syscall.Getegid())
	} else {
		log.SetOutput(slog)
	}

	out, err := exec.Command(*ipmiSel, "--debug", "--output-event-state", "--comma-separated-output", "--no-header-output").Output()
	outStr := string(out)
	if err != nil {
		log.Printf("Got error running `%s`: `%s` : `%s`\n", *ipmiSel, err, outStr) // purposely do not quit
	}

	lines := strings.Split(strings.TrimSpace(outStr), "\n")
	log.Printf("%#v\n", lines)

	if len(lines) >= 0 && lines[0] != "" {
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if len(line) == 0 {
				log.Println("Got empty line from output of ipmi-sel.")
				continue
			}
			log.Printf("Line: `%s`\n", line)
			ev, err := newIPMIEvent(line)
			log.Printf("Event: %#v\n", ev)
			if err != nil {
				log.Printf("Could not parse line `%s`, err: `%s`", line, err)
				continue
			}
		}
	} else if *deadman {
		ev := newEmptyIPMIEvent()
		fmt.Println(ev.InfluxDB(*checkName, hostname))
	}
}
