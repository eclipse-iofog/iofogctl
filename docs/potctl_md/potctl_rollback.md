## potctl rollback

Rollback ioFog resources

### Synopsis

Rollback ioFog resources to latest versions available.

```
potctl rollback RESOURCE NAME [flags]
```

### Examples

```
potctl rollback agent NAME
potctl rollback agent NAME --semver v1.0.0
```

### Options

```
  -h, --help            help for rollback
      --semver string   Target fog node version (semver.org; optional leading v)
```

### Options inherited from parent commands

```
      --debug              Toggle for displaying verbose output of API clients (HTTP and SSH)
  -n, --namespace string   Namespace to execute respective command within (default "default")
  -v, --verbose            Toggle for displaying verbose output of potctl
```

### SEE ALSO

* [potctl](potctl.md)	 - 


