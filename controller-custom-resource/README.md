# SampleCRD

Download source code.

```shell
go mod vendor
```

Generate code by k8s.io/code-generator.

```shell
chmod +x ./vendor/k8s.io/code-generator/generate-groups.sh \
  && ./hack/update-codegen.sh
```

Copy zz_generated.deepcopy.go to pkg.

```shell
cp ./github.com/justxuewei/k8s_starter/samplecrd/pkg/apis/samplecrd/v1/zz_generated.deepcopy.go \
  ./pkg/apis/samplecrd/v1/
```

Copy client to pkg.

```shell
cp -p ./github.com/justxuewei/k8s_starter/samplecrd/pkg/client \
  ./pkg/
```
