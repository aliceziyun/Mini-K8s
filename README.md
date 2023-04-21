# Mini-K8s
上海交通大学云操作系统大作业

## Pod Structure

位置：`./pkg/object/object.go`

- Pod
  - apiVersion
  - kind
  - metadata
    - name
    - labels
    - uid
    - namespace
  - spec
    - containers
    - volumes
    - nodeSelector
  - status
    - phase
    - conditions

目前，`Pod` 类型可以导出为 `json` 或者 `yaml`。
