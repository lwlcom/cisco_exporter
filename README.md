# cisco_exporter
Exporter for metrics from devices running Cisco (NX-OS/IOS XE/IOS) (via SSH) https://prometheus.io/

The basic structure is based on https://github.com/czerwonk/junos_exporter

# flags
Name     | Description | Default
---------|-------------|---------
version | Print version information. |
web.listen-address | Address on which to expose metrics and web interface. | :9362
web.telemetry-path | Path under which to expose metrics. | /metrics
ssh.targets | Comma seperated list of hosts to scrape |
ssh.user | Username to use for SSH connection | cisco_exporter
ssh.keyfile | Key file to use for SSH connection | cisco_exporter
ssh.timeout | Timeout in seconds to use for SSH connection | 5
debug | Show verbose debug output | false
legacy.ciphers | Allow insecure legacy ciphers: aes128-cbc 3des-cbc aes192-cbc aes256-cbc | false
config.file | Path to config file |

# metrics

All metrics are enabled by default. To disable something pass a flag `--<name>.enabled=false`, where `<name>` is the name of the metric.

Name     | Description | OS
---------|-------------|----
bgp | BGP (message count, prefix counts per peer, session state) | IOS XE/NX-OS
environment | Environment (temperatures, state of power supply) | NX-OS/IOS XE/IOS
facts | System informations (OS Version, memory: total/used/free, cpu: 5s/1m/5m/interrupts) | IOS XE/IOS
interfaces | Interfaces (transmitted/received: bytes/errors/drops, admin/oper state) | NX-OS (*_drops is always 0)/IOS XE/IOS
optics | Optical signals (tx/rx) | NX-OS/IOS XE/IOS

## Install
```bash
go get -u github.com/lwlcom/cisco_exporter
```

## Usage

### Binary
```bash
./cisco_exporter -ssh.targets="host1.example.com,host2.example.com:2233,172.16.0.1" -ssh.keyfile=cisco_exporter
```

```bash
./cisco_exporter -config.file=config.yml
```

## Config file
The exporter can be configured with a YAML based config file:

```yaml
---
debug: false
legacy_ciphers: false
# default values
timeout: 5
batch_size: 10000
username: default-username
password: default-password
key_file: /path/to/key

devices:
  - host: host1.example.com
    key_file: /path/to/key
    timeout: 5
    batch_size: 10000
    features: # enable/disable per host
      bgp: false
  - host: host2.example.com:2233
    username: exporter
    password: secret

features:
  bgp: true
  environment: true
  facts: true
  interfaces: true
  optics: true

```

## Third Party Components
This software uses components of the following projects
* Prometheus Go client library (https://github.com/prometheus/client_golang)

## License
(c) Martin Poppen, 2018. Licensed under [MIT](LICENSE) license.

## Prometheus
see https://prometheus.io/
