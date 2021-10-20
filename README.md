Custom Resource : apiserver

ApiGroup : raihankhan.github.io

Version : v1alpha1

Get code-generator in your `go/src` path using `git clone git@github.com:kubernetes/code-generator.git`

Generate clientset, informers, listers and zz_generated.deepcopy.go ->
```bash
execDir=~/go/src/k8s.io/code-generator
"${execDir}"/generate-groups.sh all github.com/raihankhan/httpApiServer-controller/pkg/client github.com/raihankhan/httpApiServer-controller/pkg/apis raihankhan.github.io:v1alpha1 --go-header-file "${execDir}"/hack/boilerplate.go.txt
```

`go get sigs.k8s.io/controller-tools/cmd/controller-gen@v0.7.0`

```bash
controller-gen rbac:roleName=controller-perms crd paths=./... output:crd:dir=/home/raihan/go/src/github.com/raihankhan/httpApiServer-controller/manifest output:stdout
```


