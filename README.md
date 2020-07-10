# tele
A simple Telepresence wrapper tool for microservice development.

```
Usage:
  tele [flags] <shell command>

Examples:
tele --port 8080:8080 go run main.go

Flags:
  -h, --help               help for tele
      --namespace string   name space of kubernetes (default "default")
      --port string        expose http server port
      --user string        user name for prefix of deployment name. (default "home directory name")
  -v, --version            version for tele
```

This command uses the --new-deployment option of telepresense, as shown below example.

## tele uses the --new-deployment

```
 telepresence --namespace {--namespace} --method inject-tcp --new-deployment {--user}-{repository}-{branch} --expose {--port} --run bash -c "<shell command>"
```



## Requirement

* Telepresence
* kubectl
* git
