# simple-ingress-controller

## Generate code:

Without go modules:

go get k8s.io/code-generator
go get k8s.io/apimachinery

from root directory: ./hack/update-codegen.sh

(More info at: https://itnext.io/how-to-generate-client-codes-for-kubernetes-custom-resource-definitions-crd-b4b9907769ba)

## Test on minikube:

### Build docker image inside minikube:
eval $(minikube docker-env)
docker build -t simple-ingress-controller .

### Create controller deployment and expose the proxy as service:
kubectl run simple-ingress --image=simple-ingress-controller:latest --image-pull-policy=Never --port=8080

kubectl apply -f cluster-role.yaml

kubectl apply -f cluster-role-binding.yaml

ubectl expose deployment simple-ingress --type=NodePort --port 8080

### Test (Should have a service and simple ingress rule):
curl <node-ip>:8080 <path>
