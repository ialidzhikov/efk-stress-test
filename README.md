# efk-stress-test

### Examples

```bash
go run cmd/main.go \
  --seed-kubeconfig /path/to/kubeconfig \
  --shoot-namespace shoot-namespace \
  --clients 2 \
  --elastic-host http://localhost:9200
```
