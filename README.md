# tele
A simple Telepresence wrapper tool for microservice development.

```
Usage:
  tele --run "<shell command>" [flags]

Flags:
      --port string        expose http server port
      --user string        user name for prefix of deployment name. (default "home directory name")
      --run string         shell command (default "echo hello world")
      --namespace string   name space of kubernetes (default "default")
  -h, --help               help for tele
```

This command uses the --new-deployment option of telepresense, as shown below example.

## tele uses the --new-deployment

```
 telepresence --namespace {--namespace} --method inject-tcp --new-deployment {--user}-{repository}-{branch} --expose {--port} --run bash -c "{--run}"
```



## Requirement

* Telepresence
* kubectl
* git
