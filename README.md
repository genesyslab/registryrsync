Simple utility to aid in having promotion jobs between docker registries


```
Usage:
  registryrsync [flags]

Flags:
      --config string            config file (default is $HOME/.registryrsync.yaml) (default "registryrsync.yml")
  -d, --debug                    turn on debug
      --poll duration            How frequently should we check the registries
      --port int                 Port to  listen to notifications on (default 8787)
      --source-password string   password for registry to read images from
      --source-url string        registry url to read images from
      --source-user string       username for registry to read images from
      --tag-regex string         regular expression of tags to match (default ".*")
      --target-password string   password for registry to send images to
      --target-url string        registry url to send images to
      --target-user string       username for registry to send images to
registryrsync(cleanup) $
```

You can also run this with docker, but as it uses the cli undeyr the covers you'll need to expose the docker socket


`docker run -v /var/run/docker.sock:/var/run/docker.sock registry.lab.arch.genesys.com/infra/registryrsync `