package phpfpm_docker

import (
	"context"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/plugins/inputs"

	docker "github.com/docker/docker/client"
)

type PhpFpm struct {
	DockerEndpoint            string
	DockerTimeout             config.Duration
	ContainerLabelEnable      string
	ContainerLabelEndpoint    []string
	ContainerEnvTag           []string
	ContainerLabelTag         []string
	ContainerLabelExposedPort string
	ContainerLabelExposedPath string
	ContainerLabelExposedAddr string

	docker              *docker.Client
	dockerCtx           context.Context
	telegraf            telegraf.Accumulator
	metricsEnabledLabel string
	metricsEnabledValue string
}

func (p *PhpFpm) Description() string {
	return "PhpFpm with docker auto discovery"
}

var sampleConfig = `
  ## Docker Endpoint
  ##   To use TCP, set endpoint = "tcp://[ip]:[port]"
  ##   To use environment variables (ie, docker-machine), set endpoint = "ENV"
  docker_endpoint = "unix:///var/run/docker.sock"

  ## Docker Timeout
  ##   Timeout for connecting to the docker specified endpoint
  docker_timeout = "10s"

  ## Container Label Enable
  ##   Define the label name and value that will be used to filter
  ##   containers to monitor.
  ##
  ##   Usage: METRICS_ENABLED=yes
  ##
  ##   Then on the docker labels: {
  ##      name: "METRICS_ENABLED",
  ##      value: "yes",
  ##   }
  container_label_enable = "METRICS_ENABLED=yes"

  ## Container Label Endpoint
  ##   Define the label variable that will expose the php endpoint
  ##   Currently only HTTP is supported.
  ##
  container_label_exposed_port = "METRICS_EXPOSED_PORT"

  ## Container Label Exposed Path
  ##   Define the container label that will configure the
  ##   the http path of phpfpm exposed status.
  container_label_exposed_path = "METRICS_EXPOSED_PATH"

  ## Container Label Exposed Addr
  ##   Define which container label will configure the container's endpoint
  ##   for example: METRICS_EXPOSED_ADDR
  ##
  ##   Leave it empty to fetch the container's ipv4 address
  container_label_exposed_addr = ""

  ## Container Tag Environment Variable
  ##   Pushing environment variable as tags
  ##
  ##   Usage: ["SERVICE_NAME"]
  container_env_tag = []

  ## Container Label Tag
  ##   Pushing labels as tags
  ##
  ##   Usage: ["Environment"]
  container_label_tag = []
`

func (p *PhpFpm) SampleConfig() string {
	return sampleConfig
}

func (p *PhpFpm) Init() error {
	client, err := docker.NewClientWithOpts(docker.FromEnv)
	if err != nil {
		return err
	}

	metricsEnabledLabel := strings.Split(p.ContainerLabelEnable, "=")
	if len(metricsEnabledLabel) != 2 {
		return fmt.Errorf("container_label_enable is not properly undefined: %s", p.ContainerLabelEnable)
	}

	p.docker = client
	p.metricsEnabledLabel = metricsEnabledLabel[0]
	p.metricsEnabledValue = metricsEnabledLabel[1]
	return nil
}

// outputMetric parses the response text from the phpfpm and
// adds the tags and fields on the `telegraf.Accumulator` buffer.
// function impired by the ../phpfpm plugin
func (p *PhpFpm) outputMetric(r io.Reader, addr string, addTags map[string]string) {
	stats, err := p.parsePhpfpmStats(&r)
	if err != nil {
		fmt.Errorf("not able to parse stats")
	}

	// We push the pool metric to telegraf
	for pool := range stats {
		tags := map[string]string{
			"pool": pool,
			"url":  addr,
		}
		for k, v := range addTags {
			tags[k] = v
		}

		fields := make(map[string]interface{})
		for k, v := range stats[pool] {
			fields[strings.Replace(k, " ", "_", -1)] = v
		}

		p.telegraf.AddFields("phpfpm_docker", fields, tags)
	}
}

func (p *PhpFpm) Gather(acc telegraf.Accumulator) error {
	// ensure our instruct points to the telegraf metrics buffer
	p.telegraf = acc

	containers, err := p.listContainersEndpoints()
	if err != nil {
		return err
	}

	for _, c := range containers {
		response, err := p.requestHttp(c.endpoint)
		if err == nil {
			defer response.Close()
			p.outputMetric(response, c.endpoint, c.tags)
		}
	}

	return nil
}

func init() {
	inputs.Add("phpfpm_docker", func() telegraf.Input {
		return &PhpFpm{
			DockerEndpoint:            "unix:///var/run/docker.sock",
			DockerTimeout:             config.Duration(time.Second * 5),
			ContainerLabelEnable:      "METRICS_ENABLED=yes",
			ContainerLabelExposedPort: "METRICS_EXPOSED_PORT",
			ContainerLabelExposedPath: "METRICS_EXPOSED_PATH",
			ContainerLabelExposedAddr: "METRICS_EXPOSED_ADDRESS",
			ContainerEnvTag:           []string{},
			ContainerLabelTag:         []string{},
		}
	})
}
