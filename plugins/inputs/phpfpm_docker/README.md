# PHP-FPM Input Plugin With Docker Auto Discovery

Get phpfpm stats using HTTP status page for all fpm containers.

Autodiscovery is maded by looking for configured labels in the containers.

Largely expired by the `phpfpm` and `docker` plugin.

NOTE: Currently only supports `http`.

### Configuration:

```toml
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
```

### Metrics:

- phpfpm
  - tags:
    - pool
    - url
    - container_id
  - fields:
    - accepted_conn
    - listen_queue
    - max_listen_queue
    - listen_queue_len
    - idle_processes
    - active_processes
    - total_processes
    - max_active_processes
    - max_children_reached
    - slow_requests

Additional tags can be added via `container_env_tag` and `container_label_tag`
variables.

# Example Output

```
phpfpm_docker,container_id=3f9934367a28b5d2c4e744089a0546c6324ea9071b9f5f8e39741682e7a0f63b,pool=www,url=:9001/status max_active_processes=12i,start_since=10i,listen_queue=0i,active_processes=2i,slow_requests=12i,listen_queue_len=511i,idle_processes=3i,total_processes=4i,accepted_conn=32i,max_children_reached=0i,max_listen_queue=0i 1620479105000000000
phpfpm_docker,container_id=3f9934367a28b5d2c4e744089a0546c6324ea9071b9f5f8e39741682e7a0f63b,pool=www,url=:9001/status max_active_processes=12i,start_since=10i,listen_queue=0i,active_processes=2i,slow_requests=12i,listen_queue_len=511i,idle_processes=3i,total_processes=4i,accepted_conn=32i,max_children_reached=0i,max_listen_queue=0i 1620479110000000000
phpfpm_docker,container_id=3f9934367a28b5d2c4e744089a0546c6324ea9071b9f5f8e39741682e7a0f63b,pool=www,url=:9001/status max_active_processes=12i,start_since=10i,listen_queue=0i,active_processes=2i,slow_requests=12i,listen_queue_len=511i,idle_processes=3i,total_processes=4i,accepted_conn=32i,max_children_reached=0i,max_listen_queue=0i 1620479130000000000
```
