# Change Log
All notable changes to this project will be documented in this file.
This project adheres to [Semantic Versioning](http://semver.org/).

## [UNRELEASED] - 0000-00-00
### Fixed
- add COLUMNS/LINES, TERM for interactive exec

## [1.0.0-alpha.17] - 2019-02-01
### Added
- Support terminal raw mode for process exec/run cli

## [1.0.0-alpha.16] - 2018-12-21
### Added
- [CLI] `docker-login` to push/pull docker images from CI/CD provider
### Fixed
- sorting with metadata.creationTimestamp
- activate gcloud email account

## [1.0.0-alpha.15] - 2018-10-12
### Added
- [CLI] `build` can accept build's id to monitor progress of unfinished build
- Support for CronJobs using extended Procfile
- [CLI] BuildImport API to create a build from docker archive
- [CLI] `datacol login` to support DATACOL_API_HOST and DATACOL_API_KEY env var

## [1.0.0-alpha.14] - 2018-05-15
### Fixed
- Support `<passowrd>` flag into `datacol login` CLI
- Support `--app` flags for setting limits
- Persist limit/memory values across deplopyments
- Validate certificate in `tls set` command
- Sorted keys in `datacol env`
- Showing request/limit for memory, cpu for a process 
- Bump default GKE cluster version to `1.8.8-gke.0`
- Bump GCP image to `debian-8-jessie-v20170523` version
- [CLI] Check build status before deploying an app
- Ingress integration with LocalProvider (minikube) 
- Install `unzip` for datacol controller
- Controller AMI for us-east-1 region
- `datacol destroy <name>` don't need `STACK` env var
- CLI to respect `--app` and `--stack` flag
### Added
- Inject Version, and Rollbar Token during build for CLI and API
- TLS certificates for GCP based stacks
- Index for `status` and `created_at` for datastore Build object
- Colored output for `datacol env`
- [CLI] `datacol env unset` supports multiple parameters
- Default memory limits (256Mi/512Mi) for any container
- [CLI] `limits {set|unset}` API
- [CLI] Output `CPU`, `Memory` in `ps` command
- *Certificate management API*
- `datacol switch` command

## [1.0.0-alpha.13] - 2018-04-21
### Fixed
- [CLI] Remove code and description from GRCP errors
- [API] Skipping ephemeral pods while log streaming
- [CLI] `ps` should list recent pods with tabular format (Added cpu, memory fields)
- [API] Don't stream logs from crashed/failed pods
- Sort environment variables in API response
- AWS nginx ingress controller to respect `Path: /`
- Async support for streaming logs from mutiple processes.
- Making `datacol run` independent of shell
### Added
- [CLI] Added domains:{add, remove} API
- [CLI] Renaming command `ps scale` to `scale`
- [CLI] Renaming command `build list` to `builds`
- [API] version label into k8s deployments
- Paging for `GET /v1/builds` API 
- [CLI] Tabular output for listing apps and builds
- [CLI] `STACK` env var for `datacol env`, `datacol infra`
- [CLI] Number of logs lines for process logs (`--lines 10`)

## [1.0.0-alpha.12] - 2018-03-27
### Added
- [CLI] `ps` to support container status
- -a flag for logs, ps, env and run command
- commit sha in build struct for git based app
- Docker build logs with websocket on AWS
- Adding cluster-instance-type and controller-instance-type in `datacol init`
- AWS elasticsearch support
- Websocket connection for streaming logs and Running one-off commands
- Added `--ref` flag into deploy cmd
- Proxy support through bastion Host
### Fixed
- CLI improvements
- [CLI] Bump default version of GCP cluster to `1.7.14-gke.1`
- [API] Merging app's domain individually 
- [CLI] Respect `STACK` env variable to ditermine stack
- [GCP] Ingress should have Path: '/*' to match sub-resources
- Embedding Provider for `datacol login`
- Procfile support for Codecommit based app

## [1.0.0-alpha.11] - 2018-03-02
### Added
- Supported --build-args while building containers from Dockerfile
- Procfile to manage multiple processes in an app
- Process Management API
- Supporting nginx-ingress controller on AWS based stack
- Supporting interactive command (`datacol run rails c`)
- Support for Local cloud provider based on minikube
- Domains for an App
