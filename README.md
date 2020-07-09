# tele
A simple Telepresence wrapper tool for microservice development.

```
Usage:
  tele [flags]

Flags:
      --port string        http server port (default "5004:5004")
      --user string        user name for prefix of deployment name. (default "home directory name")
      --run string         shell command (default "go run main.go")
      --namespace string   name space of kubernetes (default "default")
  -h, --help               help for tele
```

This command uses the --new-deployment option of telepresense.
[example]
```
 telepresence --namespace {--namespace} --method inject-tcp --new-deployment {--user}-{repository}-{branch} --expose {--port} --run bash -c "{--run}"
```



## Requirement

* Telepresence
* kubectl
* git
