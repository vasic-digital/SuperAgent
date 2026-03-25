# Bare Metal Installation

Install HelixAgent directly on a host machine without containers.

## Prerequisites

- Go 1.25.3 or later
- PostgreSQL 15+
- Redis 7+
- Git with SSH access configured

## Installation Steps

1. Clone the repository:

```bash
git clone git@github.com:user/HelixAgent.git
cd HelixAgent
```

2. Install dependencies and build:

```bash
make install-deps
make build
```

3. Configure environment variables by copying `.env.example` to `.env` and filling in your values.

4. Run database migrations:

```bash
psql -h localhost -U helixagent -d helixagent_db -f sql/schema/*.sql
```

5. Start the server:

```bash
./bin/helixagent
```

## Notes

- Bare metal installation requires manually managing PostgreSQL, Redis, and other dependencies
- For production deployments, container-based installation is recommended
- See [Docker installation](./docker.md) or [Podman installation](./podman.md) for containerized alternatives

## Related Documentation

- [Docker Installation](./docker.md)
- [Deployment Guide](../deployment/DEPLOYMENT_GUIDE.md)
