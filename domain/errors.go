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

import "fmt"

type InvalidResponseError struct {
	Data []byte
}

func (e InvalidResponseError) Error() string {
	return "invalid response data: " + fmt.Sprintf("%x", e.Data)
}

type InvalidHeaderError struct {
	Data []byte
}

func (e InvalidHeaderError) Error() string {
	if len(e.Data) < 2 {
		return "invalid header in response data: too short"
	}
	return "invalid header in response data (want 0x01 0x03): " + fmt.Sprintf("0x%x 0x%x", e.Data[0], e.Data[1])
}
