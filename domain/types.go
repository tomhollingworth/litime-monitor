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
package domain

import (
	"fmt"

	"github.com/spf13/viper"
)

type SolarChargeControllerSample struct {
	BatteryVoltage float32 // Volts
	BatteryCurrent float32 // Amps
	BatteryPower   int32   // Watts

	ControllerTemperature float32 // Celsius

	LoadVoltage float32 // Volts
	LoadCurrent float32 // Amps
	LoadPower   float32 // Watts

	PanelVoltage   float32 // Volts
	MaxChargePower int32   // kW
	EnergyToday    int32   // kW
	RunningDays    int32   // Days
	TotalEnergy    int32   // kW
}

func (s SolarChargeControllerSample) String() string {
	return "SolarChargeControllerSample{" +
		"BatteryVoltage: " + fmt.Sprintf("%.1f", s.BatteryVoltage) + "V, " +
		"BatteryCurrent: " + fmt.Sprintf("%.2f", s.BatteryCurrent) + "A, " +
		"BatteryPower: " + fmt.Sprintf("%d", s.BatteryPower) + "W, " +
		"ControllerTemperature: " + fmt.Sprintf("%.1f", s.ControllerTemperature) + "ºC, " +
		"LoadVoltage: " + fmt.Sprintf("%.1f", s.LoadVoltage) + "V, " +
		"LoadCurrent: " + fmt.Sprintf("%.2f", s.LoadCurrent) + "A, " +
		"LoadPower: " + fmt.Sprintf("%.1f", s.LoadPower) + "W, " +
		"PanelVoltage: " + fmt.Sprintf("%.1f", s.PanelVoltage) + "V, " +
		"MaxChargePower: " + fmt.Sprintf("%d", s.MaxChargePower) + "W, " +
		"EnergyToday: " + fmt.Sprintf("%d", s.EnergyToday) + "Wh, " +
		"RunningDays: " + fmt.Sprintf("%d", s.RunningDays) + "d, " +
		"TotalEnergy: " + fmt.Sprintf("%d", s.TotalEnergy) + "Wh"
}

func ConnectAddress() string {
	return viper.GetString("device.address")
}

func ServiceUUID() string {
	return viper.GetString("device.service")
}

func CharacteristicUUID() string {
	return viper.GetString("device.characteristic")
}

func HandleResponseData(data []byte) (sample *SolarChargeControllerSample, err error) {
	if len(data) < 42 {
		if len(data) >= 2 {
			if data[0] == 0x01 && data[1] == 0x06 {
				// Write operations (0106...) have a shorter return value
				return nil, nil
			} else {
				return nil, InvalidResponseError{Data: data}
			}
		}
		return nil, fmt.Errorf("invalid response data: too short, expected at least 42 bytes, got %d", len(data))
	}

	if data[0] != 0x01 || data[1] != 0x03 {
		return nil, InvalidHeaderError{Data: data}
	}

	sample = &SolarChargeControllerSample{}

	// Battery data
	rawBattVoltage := uint16(data[5])<<8 | uint16(data[6])
	sample.BatteryVoltage = float32(rawBattVoltage) * 0.1 // Convert to volts

	rawBattCurrent := uint16(data[7])<<8 | uint16(data[8])
	sample.BatteryCurrent = float32(rawBattCurrent) * 0.01 // Convert to amps

	rawBattPower := uint16(data[9])<<8 | uint16(data[10])
	sample.BatteryPower = int32(rawBattPower) // Convert to watts

	// Controller temperature
	sample.ControllerTemperature = float32(data[11])

	// Load data
	rawLoadVoltage := uint16(data[13])<<8 | uint16(data[14])
	sample.LoadVoltage = float32(rawLoadVoltage) * 0.1

	rawLoadCurrent := uint16(data[15])<<8 | uint16(data[16])
	sample.LoadCurrent = float32(rawLoadCurrent) * 0.01

	rawLoadPower := uint16(data[17])<<8 | uint16(data[18])
	sample.LoadPower = float32(rawLoadPower) * 0.1

	// PV input
	rawPVVoltage := uint16(data[19])<<8 | uint16(data[20])
	sample.PanelVoltage = float32(rawPVVoltage) * 0.1

	// Daily max power
	rawMaxPower := uint16(data[21])<<8 | uint16(data[22])
	sample.MaxChargePower = int32(rawMaxPower)

	// Energy charged today
	rawEnergyToday := uint16(data[23])<<8 | uint16(data[24])
	sample.EnergyToday = int32(rawEnergyToday)

	// Running days
	rawDays := uint16(data[31])<<8 | uint16(data[32])
	sample.RunningDays = int32(rawDays)

	// Total energy charged
	rawTotalEnergy := uint16(data[35])<<8 | uint16(data[36])
	sample.TotalEnergy = int32(rawTotalEnergy)

	return sample, nil
}
