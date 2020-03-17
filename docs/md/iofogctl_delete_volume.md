## iofogctl delete volume

Delete an Volume

### Synopsis

Delete an Volume.

The Volume will be deleted from the Agents that it is stored on.

```
iofogctl delete volume NAME [flags]
```

### Examples

```
iofogctl delete volume NAME
```

### Options

```
  -h, --help   help for volume
      --soft   Don't delete iofog-volume from remote host
```

### Options inherited from parent commands

```
      --detached           Use/Show detached resources
      --http-verbose       Toggle for displaying verbose output of API client
  -n, --namespace string   Namespace to execute respective command within (default "default")
  -v, --verbose            Toggle for displaying verbose output of iofogctl
```

### SEE ALSO

* [iofogctl delete](iofogctl_delete.md)	 - Delete an existing ioFog resource


