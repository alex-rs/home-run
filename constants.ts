import { Service, ServiceStatus, ConfigType } from './types';

export const MOCK_SERVICES: Service[] = [
  {
    id: '1',
    name: 'Plex Media Server',
    status: ServiceStatus.RUNNING,
    port: 32400,
    url: 'http://192.168.1.10:32400',
    uptime: '14d 2h 12m',
    cpuUsage: 12,
    memoryUsage: 2048,
    configs: [
      {
        type: ConfigType.YAML,
        path: '/opt/plex/docker-compose.yml',
        lastEdited: '2023-10-25 14:30',
        content: `version: '3'
services:
  plex:
    image: plexinc/pms-docker
    container_name: plex
    network_mode: host
    environment:
      - TZ=America/New_York
      - PLEX_CLAIM=claim-xxxxxxx
    volumes:
      - /opt/plex/config:/config
      - /opt/plex/transcode:/transcode
      - /media/movies:/data/movies
      - /media/tv:/data/tvshows
    restart: unless-stopped`
      }
    ]
  },
  {
    id: '2',
    name: 'Home Assistant',
    status: ServiceStatus.RUNNING,
    port: 8123,
    url: 'http://192.168.1.10:8123',
    uptime: '3d 5h 0m',
    cpuUsage: 5,
    memoryUsage: 512,
    configs: [
      {
        type: ConfigType.YAML,
        path: '/config/configuration.yaml',
        lastEdited: '2023-10-26 09:15',
        content: `default_config:

frontend:
  themes: !include_dir_merge_named themes

automation: !include automations.yaml
script: !include scripts.yaml
scene: !include scenes.yaml

http:
  server_port: 8123
  use_x_forwarded_for: true
  trusted_proxies:
    - 172.30.33.0/24`
      },
      {
        type: ConfigType.YAML,
        path: '/config/automations.yaml',
        lastEdited: '2023-10-27 18:45',
        content: `- id: '1632345678912'
  alias: Turn on lights at sunset
  description: ''
  trigger:
  - platform: sun
    event: sunset
    offset: 0
  condition: []
  action:
  - service: light.turn_on
    target:
      entity_id: light.living_room
  mode: single`
      }
    ]
  },
  {
    id: '3',
    name: 'Pi-hole',
    status: ServiceStatus.ERROR,
    port: 80,
    url: 'http://192.168.1.11/admin',
    uptime: '0m',
    cpuUsage: 0,
    memoryUsage: 0,
    configs: [
      {
        type: ConfigType.DOCKERFILE,
        path: '/opt/pihole/Dockerfile',
        lastEdited: '2023-09-15 11:20',
        content: `FROM pihole/pihole:latest

ENV TZ 'America/Chicago'
ENV WEBPASSWORD 'secretpassword'

# Custom blocklists
COPY adlists.list /etc/pihole/adlists.list

EXPOSE 53 53/udp
EXPOSE 80

HEALTHCHECK CMD dig +short @localhost pi.hole || exit 1`
      }
    ]
  },
  {
    id: '4',
    name: 'Grafana',
    status: ServiceStatus.RUNNING,
    port: 3000,
    url: 'http://192.168.1.10:3000',
    uptime: '45d 1h 30m',
    cpuUsage: 2,
    memoryUsage: 256,
    configs: [
      {
        type: ConfigType.INI,
        path: '/etc/grafana/grafana.ini',
        lastEdited: '2023-08-30 16:00',
        content: `[server]
http_port = 3000
domain = monitor.homelan.local

[security]
admin_user = admin
admin_password = admin

[auth.anonymous]
enabled = true
org_name = Main Org.
org_role = Viewer`
      }
    ]
  },
  {
    id: '5',
    name: 'Nginx Proxy Manager',
    status: ServiceStatus.MAINTENANCE,
    port: 81,
    url: 'http://192.168.1.10:81',
    uptime: '2d 1h',
    cpuUsage: 1,
    memoryUsage: 128,
    configs: [
      {
        type: ConfigType.JSON,
        path: '/app/config/production.json',
        lastEdited: '2023-10-28 10:10',
        content: `{
  "database": {
    "engine": "mysql",
    "host": "db",
    "name": "npm",
    "user": "npm",
    "password": "npm",
    "port": 3306
  },
  "manager": {
    "use_push_state": true
  }
}`
      }
    ]
  }
];