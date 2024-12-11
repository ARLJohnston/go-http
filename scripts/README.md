# Scripts
## K6
Requires [xk6](https://github.com/grafana/xk6) and [xk6 faker](https://github.com/grafana/xk6-faker)

```console
go install go.k6.io/xk6/cmd/xk6@latest # Install xk6 into GOPATH

$GOPATH/bin/xk6 build --with github.com/grafana/xk6-faker@latest # Patch xk6 to have faker, creates k6 binary in current directory

./k6 run --no-usage-report script.js
```

To configure number of virtual users and run duration:
```console
./k6 run --no-usage-report --duration <Duration> --vus <Num VUs> script.js
```
