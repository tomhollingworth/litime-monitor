// Copyright 2025 Tom Hollingworth
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at

//     http://www.apache.org/licenses/LICENSE-2.0

// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
package influxdb

import (
	"context"

	influxdb2 "github.com/influxdata/influxdb-client-go/v2"

	"git.tomhollingworth.io/tomhollingworth/litime-monitor/domain"
)

type InfluxDBSink struct {
	client influxdb2.Client
	org    string
	bucket string
}

func NewInfluxDBSink(url, token, org, bucket string) *InfluxDBSink {
	sink := &InfluxDBSink{client: influxdb2.NewClient(url, token), org: org, bucket: bucket}

	return sink
}

func (s *InfluxDBSink) Send(sample domain.SolarChargeControllerSample) error {
	// Convert the sample to InfluxDB line protocol format.

	writeApi := s.client.WriteAPIBlocking(s.org, s.bucket)

	point := influxdb2.NewPointWithMeasurement("solar_charge_controller").
		AddField("battery_voltage", sample.BatteryVoltage).
		AddField("battery_current", sample.BatteryCurrent).
		AddField("battery_power", sample.BatteryPower).
		AddField("controller_temperature", sample.ControllerTemperature).
		AddField("load_voltage", sample.LoadVoltage).
		AddField("load_current", sample.LoadCurrent).
		AddField("load_power", sample.LoadPower).
		AddField("load_status", sample.LoadStatus).
		AddField("panel_voltage", sample.PanelVoltage).
		AddField("max_charge_power", sample.MaxChargePower).
		AddField("energy_today", sample.EnergyToday).
		AddField("running_days", sample.RunningDays).
		AddField("total_energy", sample.TotalEnergy)

	return writeApi.WritePoint(context.Background(), point)
}
