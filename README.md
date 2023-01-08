# gocfgrender

render config with values in go template 😎

## install
```
go install github.com/phosae/gocfgrender@latest
```

## default file names in current directory
config file names: cfg, config, c
values file names: vars, vars.yml, vars.yaml, vars.json, values, values.yml, values.yaml, values.json

## examples
common usage
```shell
$ cat vars
name: myvm
namespace: test
disks:
- hostPath: /disk1
  capacity: 3000Gi
- hostPath: /disk2
  capacity: 3000Gi

$ cat c.yml
apiVersion: kubevirt.io/v1
kind: VirtualMachine
metadata:
  name: {{.name}}
  namespace: {{.namespace}}
spec:
  spec:
    volumes:
{{- range $i, $disk := .disks}}
      - name: datadisk{{$i}}
        hostDisk:
          capacity: {{$disk.capacity}}
          path: {{$disk.hostPath}}/{{$.namespace}}/vmdata.img
          type: DiskOrCreate
{{- end}}

$ gocfgrender -c c.yml -v vars
apiVersion: kubevirt.io/v1
kind: VirtualMachine
metadata:
  name: myvm
  namespace: test
spec:
  spec:
    volumes:
      - name: datadisk0
        hostDisk:
          capacity: 3000Gi
          path: /disk1/test/vmdata.img
          type: DiskOrCreate
      - name: datadisk1
        hostDisk:
          capacity: 3000Gi
          path: /disk2/test/vmdata.img
          type: DiskOrCreate
```

read from stdin
```
cat vars
name: myvm
namespace: test

cat << EOF | gocfgrender -c -
pipe heredoc> apiVersion: kubevirt.io/v1
kind: VirtualMachine
metadata:
  name: {{.name}}
  namespace: {{.namespace}}
pipe heredoc> EOF
apiVersion: kubevirt.io/v1
kind: VirtualMachine
metadata:
  name: myvm
  namespace: test
```