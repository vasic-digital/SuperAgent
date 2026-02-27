# User Manual 25: Multi-Region Deployment

## Overview
Deploying HelixAgent across multiple regions.

## Architecture
```
Region 1          Region 2
┌──────┐         ┌──────┐
│  LB  │         │  LB  │
└──┬───┘         └──┬───┘
   │                │
┌──┴───┐         ┌──┴───┐
│App 1 │         │App 2 │
└──┬───┘         └──┬───┘
   │                │
┌──┴───┐         ┌──┴───┐
│ DB 1 │◄───────►│ DB 2 │
└──────┘  Sync   └──────┘
```

## Configuration
```yaml
regions:
  - name: us-east-1
    url: https://us-east.helixagent.io
  - name: eu-west-1
    url: https://eu-west.helixagent.io
```

## DNS Setup
```
helixagent.io
├── @ A record (Geo-routed)
├── us-east CNAME us-east.elb.amazonaws.com
└── eu-west CNAME eu-west.elb.amazonaws.com
```

## Failover
Automatic failover configured with health checks every 10 seconds.
