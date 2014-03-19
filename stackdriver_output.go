// Copyright 2014, Belly, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Code based on Stackdriver custom metrics API v1.

package stackdriver

import (
	"fmt"
	sd "github.com/bellycard/stackdriver"
	"github.com/mozilla-services/heka/message"
	. "github.com/mozilla-services/heka/pipeline"
	"strconv"
	"strings"
)

type StackdriverCustomMetric struct {
	Name       string
	Value      string
	InstanceId string
}

type StackdriverCustomMetricOutput struct {
	conf    *StackdriverCustomMetricOutputConfig
	gwm     sd.GatewayMessage
	metrics map[string]StackdriverCustomMetric
	or      OutputRunner
}

// Stackdriver Output config struct
type StackdriverCustomMetricOutputConfig struct {
	// Stackdriver API key.
	ApiKey string `toml:"api_key"`
	// Interval to send metrics to Stackdriver customer metrics API.
	// Defaults to one minute intervals to send messages as that is the current limit from Stackdriver.
	TickerInterval uint `toml:"ticker_interval"`
	// Set of metric templates this output should use, keyed by field name.
	Metric map[string]StackdriverCustomMetric
}

func (so *StackdriverCustomMetricOutput) ConfigStruct() interface{} {
	return &StackdriverCustomMetricOutputConfig{
		TickerInterval: uint(60),
	}
}

func (so *StackdriverCustomMetricOutput) Init(config interface{}) (err error) {
	so.conf = config.(*StackdriverCustomMetricOutputConfig)

	// Ensure Stackdriver API key value is set from TOML configuration.
	if so.conf.ApiKey == "" {
		return fmt.Errorf("api_key must contain a Stackdriver API key.")
	}

	// Populate StackdriverCustomMetricOutput config with values from TOML configuration.
	so.metrics = so.conf.Metric

	so.gwm = sd.NewGatewayMessage()

	return
}

// FomatUnixNano converts a Unix nanosecond-precision timestamp to a Unix seconds-precision timestamp.
func FormatUnixNano(t int64) int64 {
	return int64(t / 1000000000)
}

func (so *StackdriverCustomMetricOutput) Run(or OutputRunner, h PluginHelper) (err error) {
	inChan := or.InChan()
	ticker := or.Ticker()

	var (
		pack      *PipelinePack
		values    = make(map[string]string)
		tmpMetric sd.Metric
		ok        = true
	)

	so.or = or

	for ok {
		select {
		case pack, ok = <-inChan:
			if !ok {
				break
			}

			for _, field := range pack.Message.Fields {
				if field.GetValueType() == message.Field_STRING && len(field.ValueString) > 0 {
					values[field.GetName()] = field.ValueString[0]
				}
				if field.GetValueType() == message.Field_INTEGER {
					values[field.GetName()] = strconv.FormatInt(field.ValueInteger[0], 10)
				}
				if field.GetValueType() == message.Field_DOUBLE {
					values[field.GetName()] = strconv.FormatFloat(field.ValueDouble[0], 'f', -1, 64)
				}
			}

			for _, met := range so.metrics {
				tmpMetric.Name = InterpolateString(met.Name, values)
				if met.InstanceId != "" {
					tmpMetric.InstanceId = InterpolateString(met.InstanceId, values)
				}
				tmpMetric.Value = InterpolateString(met.Value, values)
				if strings.Contains(tmpMetric.Value.(string), ".") {
					val, _ := strconv.ParseFloat(tmpMetric.Value.(string), 64)
					so.gwm.CustomMetric(tmpMetric.Name, tmpMetric.InstanceId, FormatUnixNano(pack.Message.GetTimestamp()), val)
				} else {
					val, _ := strconv.ParseInt(tmpMetric.Value.(string), 10, 64)
					so.gwm.CustomMetric(tmpMetric.Name, tmpMetric.InstanceId, FormatUnixNano(pack.Message.GetTimestamp()), val)
				}
			}
			pack.Recycle()
		case <-ticker:
			client := sd.NewStackdriverClient(so.conf.ApiKey)
			err = client.Send(so.gwm)
			if err != nil {
				so.or.LogMessage(fmt.Sprintf("[StackdriverCustomMetricOutput] API submission fail: %s\n", err))
			}
			so.gwm = sd.NewGatewayMessage()
		}
	}

	return
}

func init() {
	RegisterPlugin("StackdriverCustomMetricOutput", func() interface{} {
		return new(StackdriverCustomMetricOutput)
	})
}
