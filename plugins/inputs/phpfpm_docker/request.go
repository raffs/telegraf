package phpfpm_docker

import (
	"bufio"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

const (
	PfPool               = "pool"
	PfProcessManager     = "process manager"
	PfStartSince         = "start since"
	PfAcceptedConn       = "accepted conn"
	PfListenQueue        = "listen queue"
	PfMaxListenQueue     = "max listen queue"
	PfListenQueueLen     = "listen queue len"
	PfIdleProcesses      = "idle processes"
	PfActiveProcesses    = "active processes"
	PfTotalProcesses     = "total processes"
	PfMaxActiveProcesses = "max active processes"
	PfMaxChildrenReached = "max children reached"
	PfSlowRequests       = "slow requests"
)

type metric map[string]int64
type poolStat map[string]metric

func (p *PhpFpm) requestHttp(addr string) (r io.ReadCloser, err error) {
	u, err := url.Parse(addr)
	if err != nil {
		return nil, fmt.Errorf("unable to parse server address '%s': %v", addr, err)
	}

	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("unable to create a request to '%s': '%v'", addr, err)
	}

	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("unable to perform a request to '%s': '%v'", addr, err)
	}

	if res.StatusCode != 200 {
		return nil, fmt.Errorf("http request invalid error to '%d': '%v'", res.StatusCode, err)
	}

	return res.Body, nil
}

func (p *PhpFpm) parsePhpfpmStats(r *io.Reader) (poolStat, error) {
	var currentPool string
	stats := make(poolStat)

	scanner := bufio.NewScanner(*r)
	for scanner.Scan() {
		line := strings.Split(scanner.Text(), ":")
		if len(line) > 2 {
			continue
		}

		key, value := strings.Trim(line[0], " "), strings.Trim(line[1], " ")

		if key == PfPool {
			currentPool = value
			stats[currentPool] = make(metric)
			continue
		}

		if stats[currentPool] != nil {
			switch key {
			case PfStartSince,
				PfAcceptedConn,
				PfListenQueue,
				PfMaxListenQueue,
				PfListenQueueLen,
				PfIdleProcesses,
				PfActiveProcesses,
				PfTotalProcesses,
				PfMaxActiveProcesses,
				PfMaxChildrenReached,
				PfSlowRequests:
				valueInt, err := strconv.ParseInt(value, 10, 64)
				if err == nil {
					stats[currentPool][key] = valueInt
				}
			}
		}
	}

	return stats, nil
}
