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
package main

import (
	"log/slog"
	"time"

	"git.tomhollingworth.io/tomhollingworth/litime-monitor/adapters/influxdb"
	"git.tomhollingworth.io/tomhollingworth/litime-monitor/domain"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"tinygo.org/x/bluetooth"
)

var (
	sink    *influxdb.InfluxDBSink
	adapter = bluetooth.DefaultAdapter
)

func init() {
	pflag.String("influxdb.url", "http://127.0.0.1:8086", "InfluxDB URL")
	pflag.String("influxdb.token", "my-token", "InfluxDB token")
	pflag.String("influxdb.org", "my-org", "InfluxDB Organization")
	pflag.String("influxdb.bucket", "solar_charge_controller", "InfluxDB bucket name")
	pflag.String("device.address", "00:00:00:00:00:00", "Bluetooth device MAC address")
	pflag.String("device.service", "0000ffe0-0000-1000-8000-00805f9b34fb", "Bluetooth service UUID")
	pflag.String("device.characteristic", "0000ffe1-0000-1000-8000-00805f9b34fb", "Bluetooth characteristic UUID")
}

func main() {
	// Configure help flag
	pflag.BoolP("help", "h", false, "Show help information")

	// Load configuration
	viper.SetDefault("influxdb.url", "http://127.0.0.1:8086")
	viper.SetDefault("influxdb.token", "my-token")
	viper.SetDefault("influxdb.org", "my-org")
	viper.SetDefault("influxdb.bucket", "solar_charge_controller")

	viper.SetDefault("device.address", "00:00:00:00:00:00")
	viper.SetDefault("device.service", "0000ffe0-0000-1000-8000-00805f9b34fb")
	viper.SetDefault("device.characteristic", "0000ffe1-0000-1000-8000-00805f9b34fb")

	viper.SetConfigName("config")                // name of config file (without extension)
	viper.SetConfigType("yaml")                  // REQUIRED if the config file does not have the extension in the name
	viper.AddConfigPath("/etc/litime-monitor/")  // path to look for the config file in
	viper.AddConfigPath("$HOME/.litime-monitor") // call multiple times to add many search paths
	viper.AddConfigPath(".")                     // optionally look for config in the working directory
	viper.SetEnvPrefix("LITIME_MONITOR")

	pflag.Parse()

	// Check for help flag
	if help, _ := pflag.CommandLine.GetBool("help"); help {
		println("Litime Monitor - Bluetooth BLE device monitor")
		println("This application connects to a Bluetooth device and monitors solar charge controller data.")
		println("It sends the data to an InfluxDB instance for storage and analysis.")
		println()
		println("Usage:")
		println("  litime-monitor [flags]")
		println()
		println("Flags:")
		pflag.PrintDefaults()
		return
	}

	viper.BindPFlags(pflag.CommandLine)

	viper.AutomaticEnv() // read in environment variables that match
	slog.Info("reading config")
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			// Config file not found; ignore error if desired
		} else {
			// Config file was found but another error was produced
			slog.Error("failed to read config", "error", err.Error())
			return
		}
	}

	slog.Info("enabling")

	// Enable BLE interface.
	must("enable BLE stack", adapter.Enable())

	ch := make(chan bluetooth.ScanResult, 1)

	// Start scanning.
	slog.Info("scanning...")
	err := adapter.Scan(func(adapter *bluetooth.Adapter, result bluetooth.ScanResult) {
		slog.Info("found device", "device.address", result.Address.String(), "device.rssi", result.RSSI, "device.local_name", result.LocalName())
		if result.Address.String() == domain.ConnectAddress() {
			adapter.StopScan()
			ch <- result
		}
	})
	if err != nil {
		slog.Error("failed to start scanning", "error", err.Error())
		return
	}

	// Connect to the device.
	var device bluetooth.Device
	result := <-ch
	device, err = adapter.Connect(result.Address, bluetooth.ConnectionParams{})
	if err != nil {
		slog.Error(err.Error())
		return
	}
	slog.Info("connected", "device.address", result.Address.String())

	// Get services
	slog.Info("discovering services/characteristics")
	srvcs, err := device.DiscoverServices([]bluetooth.UUID{})
	must("discover services", err)
	if len(srvcs) == 0 {
		panic("could not find service")
	}
	var srvc bluetooth.DeviceService
	found := false
	for i, srv := range srvcs {
		slog.Info("found service", "service.uuid", srv.UUID().String())

		characteristics, err := srv.DiscoverCharacteristics([]bluetooth.UUID{})
		if err != nil {
			slog.Error("failed to discover characteristics", "error", err.Error())
			return
		}
		for _, char := range characteristics {
			slog.Info("characteristic found", "characteristic.uuid", char.UUID().String(), "characteristic.properties", char.Properties())
		}

		if srv.UUID().String() == domain.ServiceUUID() {
			srvc = srvcs[i]
			found = true
		}

	}
	if !found {
		slog.Error("service not found", "service", domain.ServiceUUID())
		return
	}
	slog.Info("found service", "service", srvc.UUID().String())

	// Discover characteristics
	chars, err := srvc.DiscoverCharacteristics([]bluetooth.UUID{})
	if err != nil {
		slog.Error(err.Error())
	}

	if len(chars) == 0 {
		panic("could not find characteristic")
	}

	var char *bluetooth.DeviceCharacteristic
	for _, c := range chars {
		if c.UUID().String() == domain.CharacteristicUUID() {
			char = &c
			break
		}
	}

	if char == nil {
		slog.Error("characteristic not found", "characteristic", domain.CharacteristicUUID())
		return
	}

	slog.Info("found characteristic", "characteristic", char.UUID().String())

	// Initialize InfluxDB sink
	slog.Info("establishing connection to influx")
	sink = influxdb.NewInfluxDBSink(viper.GetString("influxdb.url"), viper.GetString("influxdb.token"), viper.GetString("influxdb.org"), viper.GetString("influxdb.bucket"))

	// Enable notifications on the characteristic
	count := 0
	char.EnableNotifications(func(buf []byte) {
		sample, err := domain.HandleResponseData(buf)
		if err != nil {
			switch err.(type) {
			case domain.InvalidResponseError:
				slog.Error("failed to handle response data", "error", err.Error())
			case domain.InvalidHeaderError:
				// Supress invalid header errors, they are expected when the device is not ready
				return
			default:
				slog.Error("failed to handle response data", "error", err.Error())
			}
			return
		}

		// If sample is nil, it means the response was a write operation (e.g., 0106...)
		if sample == nil {
			return
		}

		// Send the sample to Sink
		if err := sink.Send(*sample); err != nil {
			slog.Error("failed to send sample to InfluxDB", "error", err.Error())
			return
		}

		// Log every 100 samples
		if count%100 == 0 {
			slog.Info("received sample",
				"BatteryVoltage", sample.BatteryVoltage,
				"BatteryCurrent", sample.BatteryCurrent,
				"BatteryPower", sample.BatteryPower,
				"ControllerTemperature", sample.ControllerTemperature,
				"LoadVoltage", sample.LoadVoltage,
				"LoadCurrent", sample.LoadCurrent,
				"LoadPower", sample.LoadPower,
				"PanelVoltage", sample.PanelVoltage,
				"MaxChargePower", sample.MaxChargePower,
				"EnergyToday", sample.EnergyToday,
				"RunningDays", sample.RunningDays,
				"TotalEnergy", sample.TotalEnergy,
			)
			count = 0
		}
		count++
	})

	// Periodically write to the characteristic to request data
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()
	for {
		_, err := char.Write([]byte{0x01, 0x03, 0x01, 0x01, 0x00, 0x13, 0x54, 0x3B})
		if err != nil {
			slog.Error("failed to write to characteristic", "error", err.Error())
			return
		}
	}
}

func must(action string, err error) {
	if err != nil {
		panic("failed to " + action + ": " + err.Error())
	}
}
