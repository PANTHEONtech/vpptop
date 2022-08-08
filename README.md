# VPPTop

**VPPTop** is a Go implementation of real-time data viewer for VPP interfaces and metrics displayed in dynamic terminal user interface.

## Preview

Below is a short demo preview of **VPPTop** in action:

[![preview][preview-svg]][preview]

## Branches

|Branch|Supported VPP versions|
|---|---|
|[![master][badge-master]][branch-master]| [21.01][vpp-21.01], [21.06][vpp-21.06], [22.02][vpp-22.02]|
|[![vpp1904][badge-1904]][branch-1904]|[19.04][vpp-19.04]|

## Features

VPPTop currently supports following metrics:

* **Interfaces** - shows full list of interfaces with associated data like VPP interface index, MTU, real-time Rx/Tx counters, dropped packets and so on. 
* **Node stats** - information about VPP runtime including node name, state, clocks, vectors, calls, suspends...      
* **Error counters** - number of errors with associated node and reason.
* **Memory usage** - data about free and used memory per thread.
* **Thread info** - displays data about thread ID and name, PID, number of cores, etc.

## VPP Requirements

VPPTop requires [Go][go-download] **1.11** (or later) to install and run. [VPP][wiki-vpp] versions supported through the VPP-Agent are as follows:
- **21.01**
- **21.06**
- **22.02**

VPP version supported by the local implementation:
- **v21.01-rc0~282**

All versions except the local one are enabled via Ligato [vpp-agent][vpp-agent] (current version is v3.1.0). The local version implementation resides directly in the VPPTop. The meaning of the local VPP support is to easily allow additional VPP versions without manipulating or updating external dependencies. The guide about how to change local in order to support custom version can be found later in the document.

**Note:** for full support of an interface/node names in VPPTop, the VPP version has to be **19.04.4** or newer. The release version of VPP 19.04 does not work properly, because of the [stats API versioning][stats-version-commit], which was added later after the release of VPP 19.04. If you want to use this version of the VPP, use the _stable/1904_ VPPTop branch where it was backported.  

### Install VPP

To install VPP from Packagecloud on Ubuntu 18.04, run following commands where you replace `<VERSION>` with either `2101`, `2106`, `2202` or `master` for the latest version:

```
curl -s https://packagecloud.io/install/repositories/fdio/<VERSION>/script.deb.sh | sudo bash
sudo apt-get install -y vpp vpp-dev vpp-plugin-core
```

For more information about how to install VPP from packages, see the [FD.io wiki page][vpp-install]. 

### Configure VPP

The VPPTop uses VPP stats API for retrieving statistics. Older VPP versions may have the stats API disabled by default. If this is the case, use following guide  [`statseg` section][stats-guide] to learn how to modify VPP config and enable it, or use the following in the VPP config:

```
# enable stats socket (default at /run/vpp/stats.sock)
statseg {
    default
}
```

## Install & Run VPPTop

### Prerequisites

- Go 1.17+

### Install

To install VPPTop run:

```shell
# install latest release version of vpptop
go install go.pantheon.tech/vpptop@latest
# install master branch version of vpptop
go install go.pantheon.tech/vpptop@master
```

### Run

To start VPPTop run:

```shell
# sudo might be required here, because of the permissions to stats socket file
sudo -E vpptop
```

In case you have cloned the repository, use can use `make` to build or install binaries:
```shell
make build
# or
make install
```

The command builds a single VPPTop binary supporting both, VPP-Agent-based VPP versions mentioned above, and the local VPP version:

VPPTop also supports a light terminal theme. To use darker colors which have better visibility on light background set `VPPTOP_THEME_LIGHT` environment variable.

**Note:** VPPTop expects VPP be running during the startup. Delayed start is currently not available.

### Keybindings

1. Keyboard arrows ``Up, Down, Left, Right`` to switch tabs, scroll.
2. ``Crtl-Space`` open/close menu for sort by a column for the active table.
3. ``/`` to filter the active table, `Enter` to keep the filter.
4. ``Esc`` to cancel the previous operation.
5. ``PgDn PgUp`` to skip pages in the active table.
6. ``Ctrl-C`` to clear counters for the active table.
7. ``q`` to quit from the application

## Custom VPP guide

As it was mentioned, VPPTop is tightly bound with the VPP version it tries to connect to. Supported versions are provided from two sources, the Ligato VPP-Agent and from the local implementation. 

The VPP-Agent usually supports multiple VPP versions. However, it may happen that the latest (or some specific) version is not compatible. In this case, VPPTop local may become helpful.

**1. Install desired VPP version.** Related binary API files must be available in the filesystem.
**2. Generate binary API.** The API is located in `/stats/local/binapi` directory. For VPPTop purposes, following are needed: 
    * dhcp
    * interfaces
    * ip
    * vpe
  Files can be easily generated using Makefile's `make generate` target.
**3. Update VPPTop vppcalls.** They are located in `stats/local/binapi/vppcalls`. Resolve all conflicts and rebuild the VPPTop binary.

Note that the agent and local implementations are independent, and may be used separately. The `vpptop-local` is and example - only locally supported VPP version can be used. 
Entity managing given implementation is **handler**, meaning there are two handlers available - the VPP handler (agent) and the local handler. The handler communicates with the VPP (reads data shown). Every handler has it own definition **HandlerDef** which validates whether associated handler is compatible with the connected VPP. Handler definitions are passed to the VPPTop client as follows:

```
client.Defs = append(client.Defs, &local.HandlerDef{}, &vpp.HandlerDef{})
```

In the code above, both handlers are provided which means VPPTop iterates over them until it founds the one suitable for the given VPP. Removing a definition, the handler is excluded.    

[badge-1904]: https://img.shields.io/badge/branch-vpp1904-orange.svg?logo=git&logoColor=white
[badge-master]: https://img.shields.io/badge/branch-master-blue.svg?logo=git&logoColor=white
[branch-master]: https://github.com/PANTHEONtech/vpptop/tree/master
[branch-1904]: https://github.com/PANTHEONtech/vpptop/tree/vpp1904
[go-download]: https://golang.org/dl/
[preview]: https://asciinema.org/a/NHODZM2ebcwWFPEEPcja8X19R
[preview-svg]: https://asciinema.org/a/NHODZM2ebcwWFPEEPcja8X19R.svg
[stats-guide]: https://wiki.fd.io/view/VPP/Command-line_Arguments#statseg_.7B_..._.7D
[stats-version-commit]: https://github.com/FDio/vpp/commit/1cb333cdf5ce26557233c5bdb5a18738cb6e1e2c
[vpp-19.04]: https://packagecloud.io/fdio/1904
[vpp-21.01]: https://packagecloud.io/fdio/2101
[vpp-21.06]: https://packagecloud.io/fdio/2106
[vpp-22.02]: https://packagecloud.io/fdio/2202
[vpp-agent]: https://github.com/ligato/vpp-agent
[vpp-install]: https://wiki.fd.io/view/VPP/Installing_VPP_binaries_from_packages
[wiki-tui]: https://en.wikipedia.org/wiki/Text-based_user_interface
[wiki-vpp]: https://wiki.fd.io/view/VPP
