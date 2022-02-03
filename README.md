cpufrequi for raspberry pi
==========================

![cpufrequi sample](cpufrequi-sample.png)

cpufrequi is a tool to display cpu clock averages over a chosen time window on a 
Raspberry Pi 4 Model B with current kernels (developed on 5.10.x)
(https://www.raspberrypi.com/products/raspberry-pi-4-model-b/)

The data is obtained from

```
/sys/devices/system/cpu/cpu0/cpufreq/stats/time_in_state
```
see https://www.kernel.org/doc/Documentation/cpu-freq/cpufreq-stats.txt


It also displays the current clock according to

```
/sys/devices/system/cpu/cpufreq/policy0/scaling_cur_freq
```

Lastly, the current cpu temperature is read from

```
/sys/devices/platform/soc/soc:firmware/raspberrypi-hwmon/hwmon/hwmon1/device/hwmon/hwmon1/subsystem/hwmon0/temp1_input
```


MOTIVATION
==========

I love benchmarking and inspecting performance metrics of computers. Its no different for my raspberry,
but for various reasons i cant use the tools i usually do (see "other platforms").
I specifically developed cpufrequi to compare the behavior of cpufreq governors "schedutil" and "ondemand" with various settings.


USAGE
=====

```
# cpufrequi -h
Usage of cpufrequi:
  -i int
        interval in ms (default 1000)
  -s int
        size of history (default 1000)
  -w int
        size of avg window (default 5)
```
For now, the history switch isnt useful. The idea is to be able to change the interval and window size at runtime in the future.
Right now you can only set it via flags when starting the program, so you would never need more history than the window size.
Setting the history size lower than the window size will crash the program :]



OTHER PLATFORMS
===============

TL;DR
use https://linux.die.net/man/1/cpupower-monitor and https://github.com/lm-sensors/lm-sensors


In theory, cpufrequi could be used for any cpu platform supporting cpufreq-stats and hwmon.
In reality though, there are some things to consider:

* Intel:
Current Intel CPUs (2022) all use the intel p state cpufreq driver. Afaik they dont support cpufreq-stats at all.

* AMD:
Current AMD CPUs (2022) use acpi-cpufreq and expose cpufreq-stats. On at least Ryzen 2700X though, it contains only 3 clock states:
```
3700000
3200000
2200000
```
Actually the CPU clocks very dynamically from 1.8GHz to up to 4.3 GHz depending on BIOS settings regarding Boost / Precision Boost Overdrive.
So the tool just doesnt make very much sense in this context, as the clock stats dont really reflect whats going on in the system (e.g. no boost info at all)

With Ryzen 5000 Series, apparently AMD introduces their own cpufreq-driver "AMD-PSTATE"
see https://www.phoronix.com/scan.php?page=news_item&px=AMD-PSTATE-2021

Another thing to consider is: 
The AMD / Intel CPUs use seperate clocks for each logical core, and thus sys-fs exposes
cpufreq endpoints for each logical core:

```
# ls -la /sys/devices/system/cpu/cpufreq/
total 0
drwxr-xr-x 19 root root    0 Feb  3 10:08 .
drwxr-xr-x 25 root root    0 Feb  3 10:03 ..
-rw-r--r--  1 root root 4096 Feb  3 12:14 boost
drwxr-xr-x  3 root root    0 Feb  3 10:08 policy0
drwxr-xr-x  3 root root    0 Feb  3 10:08 policy1
drwxr-xr-x  3 root root    0 Feb  3 10:08 policy10
drwxr-xr-x  3 root root    0 Feb  3 10:08 policy11
drwxr-xr-x  3 root root    0 Feb  3 10:08 policy12
drwxr-xr-x  3 root root    0 Feb  3 10:08 policy13
drwxr-xr-x  3 root root    0 Feb  3 10:08 policy14
drwxr-xr-x  3 root root    0 Feb  3 10:08 policy15
drwxr-xr-x  3 root root    0 Feb  3 10:08 policy2
drwxr-xr-x  3 root root    0 Feb  3 10:08 policy3
drwxr-xr-x  3 root root    0 Feb  3 10:08 policy4
drwxr-xr-x  3 root root    0 Feb  3 10:08 policy5
drwxr-xr-x  3 root root    0 Feb  3 10:08 policy6
drwxr-xr-x  3 root root    0 Feb  3 10:08 policy7
drwxr-xr-x  3 root root    0 Feb  3 10:08 policy8
drwxr-xr-x  3 root root    0 Feb  3 10:08 policy9
drwxr-xr-x  2 root root    0 Feb  3 12:14 schedutil
```

On the raspberry, all cores clock the same, therefor cpufreq only
has one policy for the whole chip

```
# ls -la /sys/devices/system/cpu/cpufreq/
total 0
drwxr-xr-x  4 root root 0 Jan 27 03:00 .
drwxr-xr-x 10 root root 0 Jan  1  1970 ..
drwxr-xr-x  3 root root 0 Jan 27 03:00 policy0
drwxr-xr-x  2 root root 0 Feb  3 12:14 schedutil
```


So on systems with different clocks for each core, the tool just doesnt make sense.
You would really need an instance of the tool for each logical core, or have the tool
build some kind of average over all cores.

For these systems i suggest to use cpupower-monitor, which displays averages for cpu clocks per core
in the Mperf section

https://linux.die.net/man/1/cpupower-monitor
https://git.kernel.org/pub/scm/linux/kernel/git/torvalds/linux.git/tree/tools/power/cpupower/




