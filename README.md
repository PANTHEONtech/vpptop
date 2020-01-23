<h1 align="center">vpptop</h1>

<p align="center"><b>vpptop</b> is a Go implementation of real-time viewer for VPP metrics shown in dynamic terminal user interface.</p>

---

## Preview

Below is short demo preview of **vpptop** in action.

[![preview](https://asciinema.org/a/NHODZM2ebcwWFPEEPcja8X19R.svg)](https://asciinema.org/a/NHODZM2ebcwWFPEEPcja8X19R)

## Branches

|Branch|Info|
|---|---|
|[![master](https://img.shields.io/badge/branch-master-blue.svg?logo=git&logoColor=white)](https://github.com/PantheonTechnologies/vpptop/tree/master)| support for the latest vpp releases [vpp 19.08](https://packagecloud.io/fdio/1908), [vpp 19.04](https://packagecloud.io/fdio/1904), work in progress [vpp 2001](https://packagecloud.io/fdio/master) |
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
 - [Go](https://golang.org/dl/) **1.11** (or later versions)
 - [VPP](https://wiki.fd.io/view/VPP) with one of the following versions:
    - **19.04.2**
    - **19.08.1**
    - **20.01** (work in progress)

### Install VPP

To install VPP from packagecloud on Ubuntu 18.04, run following commands 
where you replace `<VERSION>` with either `1904`, `1908`, or `master` for the latest version:

```sh
curl -s https://packagecloud.io/install/repositories/fdio/<VERSION>/script.deb.sh | sudo bash
sudo apt-get install -y vpp vpp-dev vpp-plugin-core
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

[wiki-tui]: https://en.wikipedia.org/wiki/Text-based_user_interface
[stats-version-commit]: https://github.com/FDio/vpp/commit/1cb333cdf5ce26557233c5bdb5a18738cb6e1e2c
