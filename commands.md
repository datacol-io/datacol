
test deployment for having redis, pg, workers etc
    http://louistiao.me/posts/walkthrough-deploying-a-flask-app-with-redis-queue-rq-workers-and-dashboard-using-kubernetes/

datacol init --nodes 1 --zone asia-south1-a --name staging gcp

datacol init --machine-type t2.medium --key datacol-8 aws

https://github.com/googlefonts/fontbakery-dashboard/issues/3

Yaxis consideration

rails c
pubsub support for RDS
elasticsearch integration
rollback / history

https://github.com/convox/rack/pull/449/files

DATACOL_STACK=dev DATACOL_PROVIDER=local ./apictl

https://gist.github.com/camilb/69008e4dddeb5d663e2e6699b3a93283

curl -Ls /tmp https://storage.googleapis.com/datacol-dev/binaries/1.0.0-alpha.13/apictl.zip > /tmp/apictl.zip &&
              unzip /tmp/apictl.zip -d /opt/datacol &&
              chmod +x /opt/datacol/apictl


kubectl delete pods,svc,ing,deployments,roles,rolebinding --namespace ingress-nginx --all

https://github.com/rubykube/barong/blob/master/Dockerfile

## setup minikube registry locally
minikube start --insecure-registry localhost:5000 && \ 
    eval $(minikube docker-env) && \
    docker run -d -p 5000:5000 --restart=always --name registry registry:2

DATACOL_STACK=dev DATACOL_PROVIDER=local ./apictl

pg_restore --verbose --clean --no-acl --no-owner -h  demo-postgres-46683.c1e8rxlhsmop.ap-southeast-1.rds.amazonaws.com -U app -d app latest.dump

openssl req -x509 -nodes -days 365 -newkey rsa:2048 -keyout tls.key -out tls.crt -subj "/CN=foo.bar.com"

things other people can help
1. Showing the list of things in cmd into a table like builds, releases


https://github.com/ChrisTerBeke/k8s-project-infrastructure - nginx ingress on GCP