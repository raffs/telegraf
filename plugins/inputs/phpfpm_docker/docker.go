package phpfpm_docker

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/docker/docker/api/types"
)

type Container struct {
	endpoint string
	tags     map[string]string
}

func (p *PhpFpm) listContainersEndpoints() ([]Container, error) {
	var containers []Container

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(p.DockerTimeout))
	defer cancel()

	opts := types.ContainerListOptions{}
	dockerContainers, err := p.docker.ContainerList(ctx, opts)
	if err != nil {
		return nil, err
	}

	for _, container := range dockerContainers {
		tags := make(map[string]string)

		// Check whether container has the `container_metrics_enabled` label set
		// and accordinging to the expected configure variable.
		enabled, found := container.Labels[p.metricsEnabledLabel]
		if !found || enabled != p.metricsEnabledValue {
			continue
		}

		exposedPort, found := container.Labels[p.ContainerLabelExposedPort]
		if !found {
			// if we didn't find the exposed port, we should move on
			continue
		}

		exposedPath, found := container.Labels[p.ContainerLabelExposedPath]
		if !found {
			// if we didn't find the exposed path, we should move on
			continue
		}

		info, err := p.docker.ContainerInspect(ctx, container.ID)
		if err != nil {
			continue
		}

		exposedAddr, found := container.Labels[p.ContainerLabelExposedAddr]
		if !found {
			// fallback to container's network settings
			exposedAddr = info.NetworkSettings.DefaultNetworkSettings.IPAddress
		}
		endpoint := fmt.Sprintf("http://%s:%s/%s", exposedAddr, exposedPort, exposedPath)

		// add listed labels as metrics tags
		for _, label := range p.ContainerLabelTag {
			if value, found := container.Labels[label]; found {
				tags[label] = value
			}
		}

		// add listed envvars as metrics tags
		for _, envvar := range p.ContainerEnvTag {
			for _, cEnvName := range info.Config.Env {
				env := strings.Split(cEnvName, "=")
				if len(env) == 2 &&
					len(strings.TrimSpace(env[1])) != 0 &&
					strings.TrimSpace(env[0]) == envvar {
					tags[env[0]] = env[1]
				}
			}
		}

		tags["container_id"] = container.ID
		containers = append(containers, Container{
			endpoint: endpoint,
			tags:     tags,
		})
	}

	return containers, nil
}
