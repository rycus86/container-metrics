# Simple container metrics in Go

Collect [Docker](https://www.docker.com/) container and daemon statistics with [Prometheus](https://prometheus.io/).

## Motivation

I'm running a [Home Lab](https://blog.viktoradam.net/tag/home-lab/) on resource constrained devices, and wanted a simple metrics collector with low CPU, memory and disk requirements as an alternative to (the otherwise fantastic) [Telegraf](https://github.com/influxdata/telegraf) collector.

The end result is a small static binary, packaged as a [Docker image](https://hub.docker.com/r/rycus86/container-metrics/), weighing only 3 MB (compressed).

## Usage

The easiest way to get started is to use the Docker image available.

```shell
$ docker run -d --name container-metrics \
	-p 8080:8080 \
	-v /var/run/docker.sock:/var/run/docker.sock:ro \
	rycus86/container-metrics #[flags]
```

The application accepts the following flags:

- __-p__ or __-port__: HTTP port to listen on *(default: 8080)*
- __-i__ or __-interval__: Interval for reading metrics from the engine *(default: 5s)*
- __-t__ or __-timeout__: Timeout for calling endpoints on the engine *(default: 30s)*
- __-l__ or __-labels__: Labels to keep (comma separated, accepts regex)
- __-d__ or __-debug__: Enable debug messages
- __-v__ or __-verbose__: Enable verbose messages - assumes debug

You can also build the application with Go, currently tested with version 1.10, then simply run it on the host:

```shell
$ ./container-metrics -p 8080 -i 15s -l com.docker.compose.service,com.mycompany.custom
```

## Metrics collected

Currently, the following *Gauge* metrics are exported.

### Engine metrics

- __cntm_engine_num_images__: Number of images
- __cntm_engine_num_containers__: Number of containers
- __cntm_engine_num_containers_running__: Number of running containers
- __cntm_engine_num_containers_stopped__: Number of stopped containers
- __cntm_engine_num_containers_paused__: Number of paused containers

### Container CPU metrics

- __cntm_cpu_usage_total_seconds__: Total CPU usage
- __cntm_cpu_usage_system_seconds__: CPU usage in system mode
- __cntm_cpu_usage_user_seconds__: CPU usage in user mode
- __cntm_cpu_usage_percent__: Total CPU usage in percent

### Container memory metrics

- __cntm_memory_total_bytes__: Total memory available
- __cntm_memory_usage_bytes__: Memory usage
- __cntm_memory_usage_percent__: Memory usage in percent

### Container I/O metrics

- __cntm_io_read_bytes__: I/O bytes read
- __cntm_io_write_bytes__: I/O bytes written

### Container network metrics

- __cntm_net_rx_bytes__: Network receive bytes
- __cntm_net_rx_packets__: Network receive packets
- __cntm_net_rx_dropped__: Network receive packets dropped
- __cntm_net_rx_errors__: Network receive errors
- __cntm_net_tx_bytes__: Network transmit bytes
- __cntm_net_tx_packets__: Network transmit packets
- __cntm_net_tx_dropped__: Network transmit packets dropped
- __cntm_net_tx_errors__: Network transmit errors

## License

MIT
