# alertmanager-filter
`alertmanager-filter` receives webhooks from [Prometheus Alertmanager](https://prometheus.io/docs/alerting/latest/alertmanager), filters and forwards them by a ruleset.  
These rules can consist of labels and/or the time when the webhook was received.

## Motivation
We wanted time based routing without introducing additional tools which have their own state about alerts.  
There is a proposed time-based muting of alerts in [alertmanager#2393](https://github.com/prometheus/alertmanager/pull/2393), which time interval parser this tool uses.  
We dislike the proposed muting because we more wanted a ruleset of who is oncall and as of writing this the PR is not merged.

## Docker image

Docker image is available on Docker Hub, Quay.io and GitHub

`docker pull swoga/alertmanager-filter`  
`docker pull quay.io/swoga/alertmanager-filter`  
`docker pull ghcr.io/swoga/alertmanager-filter`

You just need to map your config file into the container at `/etc/alertmanager-filter/config.yml`  
`docker run -v config.yml:/etc/alertmanager-filter/config.yml swoga/alertmanager-filter`

## Command line flags
`alertmanager-filter` requires a path to a YAML config file supplied in the command line flag `--config.file=config.yml`.

`--debug` can be used to raise the log level.

## Configuration file
```yaml
listen: <string> | default = :9776
metrics_path: <string> | default = /metrics
alerts_path: <string> | default = /alerts

time_intervals:
  # map key is the name of the time interval
  <string>: [ <time_interval>, ... ]

receivers:
  # map key is the name of the alertmanager receiver
  <string>: <receiver>
```

### `<time_interval>`
```yaml
years: ['2019', '2021:2025', ... ]
months: ['1', '3:5', 'august', 'september:december', ... ]
days_of_month: ['1', '3:5', ... ]
weekdays: ['monday', 'wednesday:friday', ... ]
times:
  - start_time: 'HH:MM'
    end_time: 'HH:MM'
  - ...
```

### `<receiver>`
see [<http_config>](https://prometheus.io/docs/alerting/latest/configuration/#http_config)
```yaml
target:
  url: <string>
  http_config: <http_config>

rules: [ <rule>, ... ]
```

### `<rule>`
Evaluates true if any `match` and no `not_match` evaluated true.  
Please see [example.yml](example.yml) for some pseudo code on how the evaluation is performed.
```yaml
match: [ <match>, ... ]
not_match: [ <match>, ... ]
```

### `<match>`
Evaluates true if all labels and the time match one of their array values.
```yaml
labels:
  <string>: [ <string>, ... ]
# must be defined in time_intervals map
times: [ <string>, ... ]
```

## Building
You can either build with `go build ./cmd/alertmanager-filter/` or to build with `promu` use `make build`