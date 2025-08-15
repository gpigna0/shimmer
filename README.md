# shimmer

A utility to automatically control your screen's brightness on Linux.
_*shimmer*_ can also be used to control the brightness similarly to
[brightnessctl](https://github.com/Hummer12007/brightnessctl) from which this program
takes inspiration. _*shimmer*_ is also highly customizable, giving you the possibility to adjust
different aspects of how the brightness is updated

## Configuration

> [!CAUTION]
> Right now support for multiple screens is lacking as there are no options to
> target specific screens. For set and auto the same brightness value will be applied
> to every screen. This could lead to problems when using absolute values and auto brightness

> [!IMPORTANT]
> Config options are subject to breaking changes as shimmer is in its early stage

This is an example configuration, more details can be found in `config.toml`.
Only one sensor can be used at a time, while support for multiple screens control is still a WIP

```toml
[sensor]
path = "/sys/bus/iio/devices/iio:device0"
[sensor.bounds]
min = 0
max = 500
[sensor.params]
evolution = 0.5
smoothness = 10
convexity = 125

[[screen]]
name = "my-screen"
path = "/sys/class/backlight/intel_backlight"


[[screen]]
name = "my-other-screen"
path = "/sys/class/backlight/acpi_video0"
```

## Installation

### From source

If you have `$GOPATH/bin` in your PATH simply clone the repo and run

```sh
cd shimmer
go install
```

Alternatively download the pre-built binary and use that directly

## Permissions

It is not recommended to run _*shimmer*_ as sudo. Instead, install the provided udev rules and add your user to the video group

## Usage

_*shimmer*_ provides basic functionality through `shimmer get` to print info about the current status of managed screens
and `shimmer set -- <value>` to control the brightness where `value` can be:

- `N` as an integer absolute value
- `N%` as a percentage of the maximum brightness
- `+-N%` as a delta from the current brightness

### Auto

The main feature of this utility is the ability to control the screen using an ambient light sensor.
To access this functionality _*shimmer*_'s daemon must be running.
Start the daemon with `shimmer daemon`. After this you can use `shimmer auto` to control
the state of auto brightness

For further help on commands and options use `shimmer [command] --help`

## IPC

The daemon can accept connections to `$XDG_RUNTIME_DIR/shimmer.sock`.
While it is technically possible to send compatible commands to the daemon,
the only one intended for external use is `listen`.
Other commands sent will result in undefined behaviour.  
After sending `listen`, the daemon will send two message types when state changes:

- `BRIGHTNESS::/path/to/device::raw_brightness::percent_brightness`
- `AUTO::active` where `active` is bool
