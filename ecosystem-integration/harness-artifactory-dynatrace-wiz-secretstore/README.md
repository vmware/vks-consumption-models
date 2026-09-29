# Travel Portal (presentation/demo app)

Lightweight Go + SQLite demo app with a static frontend branded as "Travel Portal." This repo contains a Dockerfile and a Helm chart so you can run locally in Docker or deploy to Kubernetes.

Quick notes:
- Default app port: `7540` (configurable via `TODO_PORT`)
- Default password (container/CI): `test12345` (env `TODO_PASSWORD`)
- Health endpoint: `GET /health` (used for probes)

## Requirements
- Docker (build & run)
- Helm (for Kubernetes deploy)
- Kubernetes cluster (optional)

## Build & Run (local)

Build and run with Make (image will be built locally):

```bash
make docker-build REGISTRY=localhost IMAGE=todo TAG=local
docker run --rm -p 7540:7540 -e TODO_PORT=7540 -e TODO_PASSWORD=test12345 localhost/todo:local
```

Open http://localhost:7540 in your browser — unauthenticated users are redirected to `/login.html`. Sign in using the password in `TODO_PASSWORD`.

Notes for CI systems: the Dockerfile is designed to be CI-friendly (sets `GOPROXY`, enables modules, caches `go mod download`, and builds in a multi-stage image). The runtime image runs as a non-root `app` user.

## Deploy to Kubernetes

The Helm chart is in `helm-charts`. By default the chart expects `env.todoPassword` to be provided and will render a Secret named `<release>-auth` with key `todopassword`.

Recommended deploy (example using `helm` flags):

```bash
helm upgrade --install todo-app helm-charts \
	--namespace demo --create-namespace \
	--set image.repository=<registry>/<image> \
	--set image.tag=<tag> \
	--set env.todoPassword="your-secure-password"
```

Alternative: create a Kubernetes Secret manually and skip `env.todoPassword` in values.

Persistence: the app uses SQLite by default (file `scheduler.db`). For production-like runs, mount a PVC at the path you choose using `values.volumeMounts` and `values.volumes` in the chart.

Probes: liveness/readiness probe paths are set to `/health` in `helm-charts/values.yaml`.

Security & runtime recommendations:
- Provide a non-default, strong `TODO_PASSWORD` via the Helm values/Secret.
- Set resource requests/limits in `values.yaml`.
- Consider running the container with a read-only root filesystem and explicit UID/GID via `securityContext` in values.

## CI integration

This repo does not include a specific CI pipeline. Any CI that can run Docker builds and push images will work. Minimal steps for a generic CI job:

```bash
# build
docker build -t <registry>/<image>:<tag> .
# push
docker push <registry>/<image>:<tag>
# optional: helm upgrade --install (requires kubeconfig in CI)
```

If you want, I can add a sample `gitlab-ci.yml`, `drone.yml`, or a minimal Jenkinsfile tailored to your CI provider.

## Where to look
- Dockerfile: `Dockerfile`
- Helm chart: `helm-charts/`
- Web UI: `web/` (login at `login.html`, main app at `/`)

## Updates 
- Enhance login
- Enable Secret Injection