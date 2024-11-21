Requires [xk6](https://github.com/grafana/xk6) and [xk6 faker](https://github.com/grafana/xk6-faker)

```shell
# Install xk6 into GOPATH
go install go.k6.io/xk6/cmd/xk6@latest

# Patch xk6 to have faker, creates k6 binary in current directory
$GOPATH/bin/xk6 build --with github.com/grafana/xk6-faker@latest

./k6 run --no-usage-report script.js
```
