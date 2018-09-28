## Run Locally

You can run the datacol in your local maniche and connect to a local (minikube) or remote 
Kubernetes cluster. 

### Installation

Here are the steps to follow to get up and running -

#### Install minikube

Refer to Minikube installation guide

#### Clone datacol repository into your go path

    mkdir -p $GOPATH/src/github.com/datacol-io && cd $GOPATH/src/github.com/datacol-io
    git clone git@github.com:datacol-io/datacol.git

#### Compile packages
    make -B api
    make -B cmd && ln -nfs $PWD/datacol /usr/local/bin/datacol

#### Create datacol config file

Take backup of `~/.datacol/config.json` if present

    mv ~/.datacol/config.json ~/.datacol/config.json.back

Run

    echo '{
        "context": "dev",
        "auths": [
            {
                "provider": "local",
                "name": "dev",
                "api_server": "localhost"
            }
        ]
    }' >> ~/.datacol/config.json

#### Start Local Kubernetes server

    minikube start
    eval $(minikube docker-env)

#### Start the Datacol API server

    cd $GOPATH/src/github.com/datacol-io/datacol
    DATACOL_STACK=dev DATACOL_PROVIDER=local ./apictl

#### Go to your app

You can now navigate to your app directory or clone a sample node.js application for quick testing.

    cd ~/ && git clone git@github.com:datacol-io/node-demo.git && cd node-demo

Run followings commands to deploy the app

    kubectl create ns dev
    datacol apps create
    datacol deploy




