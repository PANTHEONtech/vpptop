In order for vpptop to work correctly, collecting vpp statistics from different k8s nodes, each node should have the following defined in its
`/etc/vpp/contiv-vswitch.conf` file.

```
    statseg {
        default
        per-node-counters on
    }
```
If the startup configuration file doesn't exists, follow the steps described [here](https://github.com/contiv/vpp/blob/master/docs/setup/VPP_CONFIG.md).