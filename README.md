# Monitoring service
This background task checks the status of a given *URL*. If the request takes 
longer than a specific timeout, is cancelled (and it's considered as a 499
status code).

For every request, we store some Prometheus metrics that will be helpful to
create a dashboard to visualize the status and the evolution of the system.
Metrics are exposed in `http://localhost:port/metrics`, *port* can be adjusted 
with `--port=8080` 

![Grafana overview](/images/grafana-overview.png)

## Run commands
Additionally, the server exposes the `/exec` endpoint that allows running 
any command. Once the command has been executed, creates a new annotation
in the Grafana, this way you can track the evolution since that specific 
moment (in this dashboard we use a dashed blue line).

![Grafana annotation](/images/grafana-annotation.png)

For example, we can see that the response time decreases after a restart
of a specific service in our docker-compose.

Example request - POST `http://localhost:port/exec`
```json
{
    "command": "docker compose restart"
}
```

## Configuration

| **Flag**  | **Default**           | **Description**         |
|-----------|-----------------------|-------------------------|
| url       | https://google.com/   | Url to monitor          |
| token     | Non valid token       | Grafana API token       |
| host      | http://localhost:3000 | Grafana host            |
| frequency | 1m                    | Status check frequency  |
| timeout   | 30s                   | Request timeout         |
| port      | 8080                  | Port used by the server |
