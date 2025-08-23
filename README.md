# shimmer

A utility to automatically control your devices' brightness on Linux.
_*shimmer*_ can also be used to control the brightness similarly to
[brightnessctl](https://github.com/Hummer12007/brightnessctl) from which this program
takes inspiration. _*shimmer*_ is also highly customizable, giving you the possibility to adjust
different aspects of how the brightness is updated

## Configuration

> [!IMPORTANT]
> Config options are subject to breaking changes as shimmer is in its early stage

This is an example configuration, more details can be found in `config.toml`.
Only one sensor can be used at a time and its path needs to be specified.
Devices are found automatically by searching `/sys/class/backlight` and `/sys/class/leds`

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
```

## Installation

### From source

Just download the binary with

```sh
wget https://www.github.com/gpigna0/shimmer/blob/main/shimmer
```

and put it in your `PATH`

## Permissions

It is not recommended to run _*shimmer*_ as sudo. Instead, install the provided udev rules and add your user to the `video` group

## Usage

The core features are retrieving info about controlled devices, setting their brightness or adjusting it automatically. Here are described the most relevant commands:
for the complete list of commands and more details about their usage
call `shimmer [command] --help`

### Get

`shimmer get` prints info about the current status of managed devices.

### Set

> [!WARNING]
> Some devices are multicolor but `set` (and `auto`) ignore that, resulting
> in unintended colors. Support for multicolor devices is planned in the future

`shimmer set` to controls a device's brightness which can be expressed in different formats:

- `N` as an integer absolute value
- `N%` as a percentage of the maximum brightness
- `+-N%` as a delta from the current brightness

To use this command you must specify a set of target devices with one or more
`--device <dev_name>` or `--all`.

### Auto

`shimmer auto` controls the state of automatic brightness. In order for auto to
work _*shimmer's*_ daemon, which can be started with `shimmer daemon`, must be running.
`auto` uses the sensor specified in the config. If you don't have a sensor simply
put `path = ""` to disable `auto`.  
Like in the case of `set` auto requires at least one device to be targeted
with `--device <dev_name>`.

## IPC

The daemon can accept connections to `$XDG_RUNTIME_DIR/shimmer.sock`.
When it is active it will broadcast changes in the state of
the devices through two types of messages:

- `BRIGHTNESS::dev_name::raw_brightness::percent_brightness`
- `AUTO::dev_name::active` where `active` is either `true` or `false`

To listen for this massages connect to `$XDG_RUNTIME_DIR/shimmer.sock` and send
`listen\n`. After doing so, the updates will be sent on the connection with
one message per line
