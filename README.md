## DATACOL

[Datacol](https://www.datacol.io) provides a powerful PaaS control layer on top of AWS/GCP. 

Datacol helps you create Heroku like infrastructure for deploying container-native applications on cloud. It is a deployment platform that simplifies the process developers use to build, deploy and manage applications in the cloud. It aims to make it trivially easy to deploy a container based microservices architecture.

Datacol can be installed into your own cloud account and uses managed cloud services ( like for GCP we use [Container engine](https://cloud.google.com/container-engine/), [GCR](https://cloud.google.com/container-registry/), [ContainerBuilder](https://cloud.google.com/container-builder/)) under the hood but automates it all away to give you a better deployment experience. It uses Docker under the hood so if you want to customize anything (languages, dependencies, etc) you can simply add a Dockerfile to your project.

[![asciicast](https://asciinema.org/a/114966.png)](https://asciinema.org/a/114966)

#### Getting Started

Please follow this [guide](https://www.datacol.io/docs/getting-started/) to get up and running.

#### Community

You can [join](https://slackpass.io/datacol) our Slack team for discussion and support.

#### Development

To generate code from protobuf and bindata from static files, run

    make gen

To build the CLI
  
    make -B cli

To build the api

    make -B api

We are currently adding tests, run tests locally using 
  
    go test $(glide nv)

### License

Datacol is available under [Apache License, Version 2.0](https://www.apache.org/licenses/LICENSE-2.0).