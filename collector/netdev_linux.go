// Copyright 2015 The Prometheus Authors
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// +build !nonetdev

package collector

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"
	"path"

	"github.com/prometheus/common/log"
	"github.com/yookoala/realpath"
	"gopkg.in/alecthomas/kingpin.v2"
)

var (
	procNetDevFieldSep = regexp.MustCompile("[ :] *")

	// The PID to collect net/dev from
	netdevPid = kingpin.Flag("collector.netdev.pid",
		"PID to collect the net/dev from, defaults to 'self'. Set it to 1 to get host net/dev.").Default("self").String()
)


func getNetDevStats(ignore *regexp.Regexp) (map[string]map[string]string, error) {
   netdevPath := path.Join(*netdevPid, "net/dev")
   myRealpath, err := realpath.Realpath(procFilePath(netdevPath))
   log.Infoln("net/dev path is - ", myRealpath)
	file, err := os.Open(myRealpath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	return parseNetDevStats(file, ignore)
}

func parseNetDevStats(r io.Reader, ignore *regexp.Regexp) (map[string]map[string]string, error) {
	scanner := bufio.NewScanner(r)
	scanner.Scan() // skip first header
	scanner.Scan()
	parts := strings.Split(scanner.Text(), "|")
	if len(parts) != 3 { // interface + receive + transmit
		return nil, fmt.Errorf("invalid header line in net/dev: %s",
			scanner.Text())
	}

	header := strings.Fields(parts[1])
	netDev := map[string]map[string]string{}
	for scanner.Scan() {
		line := strings.TrimLeft(scanner.Text(), " ")
		parts := procNetDevFieldSep.Split(line, -1)
		if len(parts) != 2*len(header)+1 {
			return nil, fmt.Errorf("invalid line in net/dev: %s", scanner.Text())
		}

		dev := parts[0][:len(parts[0])]
		if ignore.MatchString(dev) {
			log.Debugf("Ignoring device: %s", dev)
			continue
		}
		netDev[dev] = map[string]string{}
		for i, v := range header {
			netDev[dev]["receive_"+v] = parts[i+1]
			netDev[dev]["transmit_"+v] = parts[i+1+len(header)]
		}
	}
	return netDev, scanner.Err()
}
