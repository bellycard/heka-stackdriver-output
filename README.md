# [Stackdriver](http://www.stackdriver.com/) Output for [Mozilla Heka](http://hekad.readthedocs.org/en/latest/)

A Mozilla Heka Output for Stackdriver.com's custom metric API.

The Stackdriver custom metrics API throttles requests at one per minute. This output will collect metrics matched from the message_matcher and emit them to the Stackdriver custom metrics API as a collection of metrics every minute.


# Installation

* You will need to aquire a Stackdriver API key. The key must have an Access Role of "Agent and Custom Metric Data Key"
 
* Create or add to the file {heka_root}/cmake/plugin_loader.cmake

```
git_clone(https://github.com/bellycard/stackdriver master)
add_external_plugin(git https://github.com/bellycard/heka-stackdriver-output master)
```

* Build Heka per normal instructions for your platform.

Additional instructions can be found in the [Heka documentation for external plugins](http://hekad.readthedocs.org/en/latest/installing.html#build-include-externals).

# Parameters

- api_key (string, required)
    Stackdriver API key.
    (default: none)
- ticker_interval (uint, optional)
    Interval to send custom metrics to Stackdriver API. The Stackdriver custom metrics API currently enforces
    a one minute throttling for sending metrics.
    (default: 60 seconds)


## Example Stackdriver Output Configuration File

```
[Stackdriver]
type = "StackdriverCustomMetricsOutput"
api_key = "EXAMPLEAPIKEY00000000000000000000"

[Stackdriver.Metrics.cpu_steal]
name = "cpu-steal"
value = "%CpuSteal%"
instance_id = ""
```
