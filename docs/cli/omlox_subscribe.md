## omlox subscribe

Subscribes to real-time events

### Synopsis


This command subscribes to read-time events from the Omlox Hub.
Subscriptions are made using named topics.

The Omlox standard supports a few:
	- location_updates
	- collision_events
	- fence_events
	- trackable_motions
	- location_updates:geojson
	- fence_events:geojson

Extra topics can be supported by vendors.


```
omlox subscribe [flags]
```

### Options

```
  -h, --help   help for subscribe
```

### Options inherited from parent commands

```
      --addr string   omlox hub API endpoint (default "localhost:8081")
      --debug         enable debug logging
```

### SEE ALSO

* [omlox](omlox.md)	 - The Omlox Hub CLI tool

