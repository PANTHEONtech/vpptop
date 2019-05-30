vpptop
======

The **vpptop** is a Go implementation of real-time viewer for VPP metrics provided by dynamic terminal user interface.

## Preview

Below is short demo preview of **vpptop** in action.

[![preview](https://asciinema.org/a/NHODZM2ebcwWFPEEPcja8X19R.svg)](https://asciinema.org/a/NHODZM2ebcwWFPEEPcja8X19R)

## Features

Following VPP stats are currently supported:

 - **Interface stats** - RX/TX packets/bytes, packet errors/drops/punts/IPv4..
 - **Node stats** - clocks, vectors, calls, suspends..
 - **Error counters** - node, reason
 - **Memory usage** - free, used..
 - **Thread info** - name, type, PID..

## Requirements

In order to install and run vpptop you need to install following requirements:

 - [Go](https://golang.org/dl/) 1.11+
 - [VPP](https://wiki.fd.io/view/VPP) (`19.04-3~g1cb333cdf~b41` is recommended, more info below)

### Install VPP

To install VPP from packagecloud on Ubuntu 18.04, run following commands:

```sh
curl -s https://packagecloud.io/install/repositories/fdio/1904/script.deb.sh | sudo bash
sudo apt-get install -y vpp vpp-dev vpp-plugin-core
```

For more info about how to install VPP from packages, see: <https://wiki.fd.io/view/VPP/Installing_VPP_binaries_from_packages>

:warning: For full support of interface/node names in vpptop, the VPP version has to be `19.04-3~g1cb333cdf~b41` or newer. The release version of VPP 19.04 will not work, because [stats API versioning][stats-version-commit] was added after the release of VPP 19.04 (it was backported to _stable/1904_ branch).

### Configure VPP

The vpptop uses VPP stats API for retrieving statistics. The VPP stats API is disabled by default and to enable it, add [`statseg` section](https://wiki.fd.io/view/VPP/Command-line_Arguments#statseg_.7B_..._.7D) to your VPP config, like this:

```sh
# this will use /run/vpp/stats.sock for stats socket
statseg {
	default
}
```

## Install & Run vpptop

To install vpptop run the following command:

```sh
# this will install vpptop to $GOPATH/bin
$ go get -u github.com/PantheonTechnologies/vpptop
```

To start vpptop run following command:

```sh
$ sudo -E vpptop
# sudo might be required, because of permissions to stats socket file
```

NOTE: The VPP should be running before starting vpptop!

### Keybindings

1. Keyboard arrows ``Up, Down, Left, Right`` to switch tabs, scroll.
2. ``Crtl-Space`` open/close menu for sort by column for the active table.
3. ``/`` to filter the active table.
4. ``Esc`` to cancel the previous operation.
5. ``PgDn PgUp`` to skip pages in active table.
6. ``Ctrl-C`` to clear counters for the active table.
7. ``q`` to quit from the application

## Developing vpptop

This section is **not required** for running vpptop and provides information for development of vpptop.

### VPP binary API

The vpptop uses GoVPP's binapi-generator to generate Go bindings for VPP binary API from JSON format. This should not normally be needed unless you want to use vpptop with different VPP version that has changed it's API.

For installation instructions for GoVPP's binapi-generator, see: <https://github.com/FDio/govpp/blob/master/README.md>

The vpptop uses [go generate](https://blog.golang.org/generate) tool to actually run the binapi-generator. To run the generator simply go to the vpptop directory and run: `go generate`.

```go
//go:generate binapi-generator --input-dir=/usr/share/vpp/api/core/ --output-dir=bin_api
```

By default it will use the JSON files stored in `/usr/share/vpp/api/core/` that are installed together with VPP and outputs the generated files to `bin_api` directory.

In case you need to use different JSON files you can change the arguments in the `gen.go` file, where you can specify the input directory of the JSON files and where the bindings should be generated.


[wiki-tui]: https://en.wikipedia.org/wiki/Text-based_user_interface
[stats-version-commit]: https://github.com/FDio/vpp/commit/1cb333cdf5ce26557233c5bdb5a18738cb6e1e2c
