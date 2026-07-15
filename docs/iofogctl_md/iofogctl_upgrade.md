## iofogctl upgrade

Upgrade ioFog resources

### Synopsis

Upgrade ioFog resources to latest versions available.

```
iofogctl upgrade RESOURCE NAME [flags]
```

### Examples

```
iofogctl upgrade agent NAME
iofogctl upgrade agent NAME --semver v1.0.0
```

### Options

```
  -h, --help            help for upgrade
      --semver string   Target fog node version (semver.org; optional leading v)
```

### Options inherited from parent commands

```
      --debug              Toggle for displaying verbose output of API clients (HTTP and SSH)
  -n, --namespace string   Namespace to execute respective command within (default "default")
  -v, --verbose            Toggle for displaying verbose output of iofogctl
```

### SEE ALSO

* [iofogctl](iofogctl.md)	 - 


