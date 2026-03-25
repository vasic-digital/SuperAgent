# Kubernetes Installation

Deploy HelixAgent on a Kubernetes cluster.

## Prerequisites

- Kubernetes 1.28+ cluster
- `kubectl` configured with cluster access
- Helm 3+ (optional, for chart-based deployment)
- Container registry access for HelixAgent images

## Quick Start

1. Build and push the container image:

```bash
make docker-build
docker tag helixagent:latest your-registry/helixagent:latest
docker push your-registry/helixagent:latest
```

2. Create the namespace and secrets:

```bash
kubectl create namespace helixagent
kubectl create secret generic helixagent-env --from-env-file=.env -n helixagent
```

3. Apply the deployment manifests (see `docs/deployment/kubernetes-deployment.md` for full manifests).

4. Verify the deployment:

```bash
kubectl get pods -n helixagent
kubectl port-forward svc/helixagent 7061:7061 -n helixagent
curl http://localhost:7061/v1/health
```

## Production Considerations

- Use `PodDisruptionBudget` for high availability
- Configure resource requests and limits
- Enable horizontal pod autoscaling based on request latency
- Use persistent volumes for PostgreSQL and Redis

## Related Documentation

- [Kubernetes Deployment Details](../deployment/kubernetes-deployment.md)
- [Production Deployment](../deployment/production-deployment.md)
- [Docker Installation](./docker.md)
