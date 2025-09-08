## omlox

The Omlox Hub CLI tool

### Synopsis

The Omlox Hub CLI tool

Common actions for omlox client:

- omlox get trackables
- omlox sub location_updates
- omlox get trackables -o json > backup.trackables.json
- omlox create trackables < backup.trackables.json
- omlox update trackables < backup.trackables.json

Environment variables:

| Name                 | Description                                                         |
|----------------------|---------------------------------------------------------------------|
| OMLOX_HUB_API        | Omlox hub API endpoint.                                             |


### Options

```
      --addr string   omlox hub API endpoint (default "localhost:8081")
      --debug         enable debug logging
  -h, --help          help for omlox
```

### SEE ALSO

* [omlox create](omlox_create.md)	 - Create hub resources
* [omlox delete](omlox_delete.md)	 - Delete hub resources
* [omlox gen](omlox_gen.md)	 - Generate commands
* [omlox get](omlox_get.md)	 - Get hub resources
* [omlox subscribe](omlox_subscribe.md)	 - Subscribes to real-time events
* [omlox update](omlox_update.md)	 - Update hub resources
* [omlox version](omlox_version.md)	 - Show version information

