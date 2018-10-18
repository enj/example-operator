# Example Operator

```shell
cd "${GOPATH}/src/github.com/enj/example-operator"

oc new-project example-operator

oc create -f manifests/00-crd.yaml
oc create -f install/cr.yaml

make
export PATH="_output/local/bin/linux/amd64:${PATH}"
example operator --kubeconfig admin.kubeconfig --config install/config.yaml --v=4

oc get exampleoperator example-operator-resource -o yaml
oc get secret example-operator-resource -o yaml
```
