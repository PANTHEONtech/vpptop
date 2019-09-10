<h1 align="center">vpptop</h1>

<p align="center"><b>vpptop</b> is a Go implementation of real-time viewer for VPP metrics shown in dynamic terminal user interface.</p>

---

## Preview

Below is short demo preview of **vpptop** in action.

[![preview](https://asciinema.org/a/NHODZM2ebcwWFPEEPcja8X19R.svg)](https://asciinema.org/a/NHODZM2ebcwWFPEEPcja8X19R)

## Branches

|Branch|Info|
|---|---|
|[![vpp1908](https://img.shields.io/badge/branch-master-blue.svg?logo=git&logoColor=white)](https://github.com/PantheonTechnologies/vpptop/tree/vpp1908)| support for [vpp 19.08](https://packagecloud.io/fdio/1908)|
|[![vpp1904](https://img.shields.io/badge/branch-vpp1904-orange.svg?logo=git&logoColor=white)](https://github.com/PantheonTechnologies/vpptop/tree/vpp1904)|support for [vpp 19.04](https://packagecloud.io/fdio/1904)|

## Features

Following VPP metrics are currently supported:

 - **Interface stats** - RX/TX packets/bytes, packet errors/drops/punts/IPv4..
 - **Node stats** - clocks, vectors, calls, suspends..
 - **Error counters** - node, reason
 - **Memory usage** - free, used..
 - **Thread info** - name, type, PID..

## Requirements

In order to install and run vpptop you need to install following requirements:
 - [Go](https://golang.org/dl/) **1.11**
 - [VPP](https://wiki.fd.io/view/VPP) (**19.04.1** recommended, more info below)

### Install VPP

To install VPP from packagecloud on Ubuntu 18.04, run following commands:

```sh
export VPP_VER=19.04.1-release
curl -s https://packagecloud.io/install/repositories/fdio/release/script.deb.sh | sudo bash
sudo apt-get install -y vpp=$VPP_VER vpp-dev=$VPP_VER vpp-plugin-core=$VPP_VER libvppinfra-dev=$VPP_VER libvppinfra=$VPP_VER
```

For more info about how to install VPP from packages, see: <https://wiki.fd.io/view/VPP/Installing_VPP_binaries_from_packages>

For full support of interface and node names in vpptop, the VPP version has to be **19.04.1** or newer. The release version of VPP 19.04 will not work, because [stats API versioning][stats-version-commit] was added after the release of VPP 19.04 (it was backported to _stable/1904_ branch).

### Configure VPP

The vpptop uses VPP stats API for retrieving statistics. The VPP stats API is disabled by default and to enable it, add [`statseg` section](https://wiki.fd.io/view/VPP/Command-line_Arguments#statseg_.7B_..._.7D) to your VPP config, like this:

```sh
# enable stats socket (default at /run/vpp/stats.sock)
statseg {
    default
}
```

## Install & Run vpptop

To install vpptop run:

```sh
# install vpptop to $GOPATH/bin
$ go get -u github.com/PantheonTechnologies/vpptop
```

To start vpptop run:

```sh
# sudo might be required here, because of the permissions to stats socket file
$ sudo -E vpptop
```

vpptop also supports light terminal theme. To use darker colors which are better visible on light background, <br>
you can set the `VPPTOP_THEME_LIGHT` environment variable.


**NOTE:** The VPP should be running before starting vpptop!

### Keybindings

1. Keyboard arrows ``Up, Down, Left, Right`` to switch tabs, scroll.
2. ``Crtl-Space`` open/close menu for sort by column for the active table.
3. ``/`` to filter the active table, `Enter` to keep the filter.
4. ``Esc`` to cancel the previous operation.
5. ``PgDn PgUp`` to skip pages in active table.
6. ``Ctrl-C`` to clear counters for the active table.
7. ``q`` to quit from the application

## Developing vpptop

This section is **not required** for running vpptop and provides info about vpptop development.

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
