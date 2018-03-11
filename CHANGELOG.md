# Change Log
All notable changes to this project will be documented in this file.
This project adheres to [Semantic Versioning](http://semver.org/).

## [UNRELEASED] - 0000-00-00
### Added
- AWS elasticsearch support
- Websocket connection for streaming logs and Running one-off commands
- Added `--ref` flag into deploy cmd
- Proxy support through bastion Host 
### Fixed
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