/*
Copyright 2022 Gravitational, Inc.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package types

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestMaintenceWindowAgentUpgrade(t *testing.T) {
	newTime := func(day int, hour int) time.Time {
		return time.Date(
			2000,
			time.January,
			day,
			hour,
			0, // min
			0, // sec
			0, // nsec
			time.UTC,
		)
	}

	from := newTime(1, 12)

	require.Equal(t, time.Sunday, from.Weekday()) // verify that newTime starts from expected pos

	conf := AgentUpgradeWindow{
		UTCStartHour: 2,
	}

	tts := []struct{ start, stop time.Time }{
		{newTime(1, 2), newTime(1, 3)},
		{newTime(2, 2), newTime(2, 3)},
		{newTime(3, 2), newTime(3, 3)},
		{newTime(4, 2), newTime(4, 3)},
		{newTime(5, 2), newTime(5, 3)},
		{newTime(6, 2), newTime(6, 3)},
		{newTime(7, 2), newTime(7, 3)},
		{newTime(8, 2), newTime(8, 3)},
		{newTime(9, 2), newTime(9, 3)},
	}

	gen := conf.Generator(from)

	for _, tt := range tts {
		start, stop := gen()
		require.Equal(t, tt.start, start)
		require.Equal(t, tt.stop, stop)
	}

	// set weekdays fileter s.t. windows limited to m-f.
	conf.Weekdays = []string{
		"Monday",
		"tue",
		"Wed",
		"thursday",
		"Friday",
	}

	tts = []struct{ start, stop time.Time }{
		// sun {newTime(1, 2), newTime(1, 3)},
		{newTime(2, 2), newTime(2, 3)},
		{newTime(3, 2), newTime(3, 3)},
		{newTime(4, 2), newTime(4, 3)},
		{newTime(5, 2), newTime(5, 3)},
		{newTime(6, 2), newTime(6, 3)},
		// sat {newTime(7, 2), newTime(7, 3)},
		// sun {newTime(8, 2), newTime(8, 3)},
		{newTime(9, 2), newTime(9, 3)},
	}

	gen = conf.Generator(from)

	for _, tt := range tts {
		start, stop := gen()
		require.Equal(t, tt.start, start)
		require.Equal(t, tt.stop, stop)
	}

	// verify that invalid weekdays are omitted from filter.
	conf.Weekdays = []string{
		"Monday",
		"tues", // invalid
		"Wed",
		"Th", // invalid
		"Friday",
	}

	tts = []struct{ start, stop time.Time }{
		// sun {newTime(1, 2), newTime(1, 3)},
		{newTime(2, 2), newTime(2, 3)},
		// tue {newTime(3, 2), newTime(3, 3)},
		{newTime(4, 2), newTime(4, 3)},
		// thu {newTime(5, 2), newTime(5, 3)},
		{newTime(6, 2), newTime(6, 3)},
		// sat {newTime(7, 2), newTime(7, 3)},
		// sun {newTime(8, 2), newTime(8, 3)},
		{newTime(9, 2), newTime(9, 3)},
	}

	gen = conf.Generator(from)

	for _, tt := range tts {
		start, stop := gen()
		require.Equal(t, tt.start, start)
		require.Equal(t, tt.stop, stop)
	}

	// if all weekdays are invalid, revert to firing every day
	conf.Weekdays = []string{
		"Mo",
		"Tu",
		"We",
		"Th",
		"Fr",
	}

	tts = []struct{ start, stop time.Time }{
		{newTime(1, 2), newTime(1, 3)},
		{newTime(2, 2), newTime(2, 3)},
		{newTime(3, 2), newTime(3, 3)},
		{newTime(4, 2), newTime(4, 3)},
		{newTime(5, 2), newTime(5, 3)},
		{newTime(6, 2), newTime(6, 3)},
		{newTime(7, 2), newTime(7, 3)},
		{newTime(8, 2), newTime(8, 3)},
		{newTime(9, 2), newTime(9, 3)},
	}

	gen = conf.Generator(from)

	for _, tt := range tts {
		start, stop := gen()
		require.Equal(t, tt.start, start)
		require.Equal(t, tt.stop, stop)
	}
}

// verify that the default (empty) maintenance window value is valid.
func TestMaintenanceWindowDefault(t *testing.T) {
	t.Parallel()

	mw := NewMaintenanceWindow()

	require.NoError(t, mw.CheckAndSetDefaults())
}

func TestMaintenanceWindowWeekdayParser(t *testing.T) {
	t.Parallel()

	tts := []struct {
		input  string
		expect time.Weekday
		fail   bool
	}{
		{
			input:  "Tue",
			expect: time.Tuesday,
		},
		{
			input:  "tue",
			expect: time.Tuesday,
		},
		{
			input: "tues",
			fail:  true, // only 3-letter shorthand is accepted
		},
		{
			input:  "Saturday",
			expect: time.Saturday,
		},
		{
			input:  "saturday",
			expect: time.Saturday,
		},
		{
			input:  "sun",
			expect: time.Sunday,
		},
		{
			input: "sundae", // containing a valid prefix is insufficient
			fail:  true,
		},
		{
			input: "",
			fail:  true,
		},
	}

	for _, tt := range tts {
		day, ok := parseWeekday(tt.input)
		if tt.fail {
			require.False(t, ok)
			continue
		}

		require.Equal(t, tt.expect, day)
	}
}
