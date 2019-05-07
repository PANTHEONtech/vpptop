# vpptop

Introduction
============

This is `vpptop`, a terminal only application implemented in Go.
The goal of vpptop is to provide VPP statistics which are updated
in real time.

The current version 1.11 supports statistics for:
 1. interfaces
 2. nodes
 3. errors
 4. memory usage per thread
 5. thread data
 
 [![asciicast](https://asciinema.org/a/NHODZM2ebcwWFPEEPcja8X19R.svg)](https://asciinema.org/a/NHODZM2ebcwWFPEEPcja8X19R)

Install
====== 
In order to run vpptop you will need to install the dependencies
first.
1. [termui-go](https://github.com/gizak/termui). ``go get github.com/gizak/termui``<br>
3. [GoVPP](https://github.com/FDio/govpp). ``go get git.fd.io/govpp.git``<br>

If you run into some problems with installing a dependency we suggest you read<br>
their ``README.md`` file which you can find in the links provided above.


You will also need [VPP](https://github.com/FDio/vpp) with version at least 19.04-6. You can read about the <br>
installation process here <https://wiki.fd.io/view/VPP/Installing_VPP_binaries_from_packages>.

After installing the dependencies you can run the application from the parent directory. <br>
You may need root access.<br>
``sudo go run ./govpp-statsviewer/``

NOTE:// Interfaces and Node names will not be displayed with VPP version below 19.04-6,
which is the result of adding [version-ing](https://github.com/FDio/govpp/blob/master/adapter/vppapiclient/stat_client_wrapper.h#L29) for the GoVPP stats-api <br>

Binary API
=============
<b>This step is not needed to run ``GoVPP stats-viewer``<br>Read this section only if you need to rebuild the binary APIs. <br>:For example if you
want to run stats-viewer on a different VPP version.</b>

You will need to install the ``binapi-generator`` in your $GOPATH. You can find the process of <br>
the installation here <https://github.com/FDio/govpp>.

In go, [go generate](https://blog.golang.org/generate) tool can be used to generate Go bindings from VPP APis in JSON format. <br>
You can generate the bindigs from the root directory of the project using the command ``go generate``.

By default it will use the JSON files stored in ``/usr/share/vpp/api/core/`` and will generate the output <br>
in the root directory of the project inside the ``bin_api`` directory.

In case you need to use different JSON files you can change it in the ``gen.go`` file.
Where you can specify <br> the input directory of the JSON files and where the bindings should be generated.

``//go:generate binapi-generator --input-dir=/usr/share/vpp/api/core/ --output-dir=bin_api``

Usage
====

Keybindings:
1. Keyboard arrows ``Up, Down, Left, Right`` to switch tabs, scroll.
2. ``Crtl-Space`` open/close menu for sort by column for the active table.
3. ``/`` to filter the active table.
4. ``Esc`` to cancel the previous operation.
5. ``PgDn PgUp`` to skip pages in active table.
6. ``Ctrl-C`` to clear counters for the active table.
7. ``q`` to quit from the application