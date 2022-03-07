//go:build !bench
// +build !bench

package hw10programoptimization

import (
	"bytes"
	"io"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGetDomainStat(t *testing.T) {
	data := `{"Id":1,"Name":"Howard Mendoza","Username":"0Oliver","Email":"aliquid_qui_ea@Browsedrive.gov","Phone":"6-866-899-36-79","Password":"InAQJvsq","Address":"Blackbird Place 25"}
{"Id":2,"Name":"Jesse Vasquez","Username":"qRichardson","Email":"mLynch@broWsecat.com","Phone":"9-373-949-64-00","Password":"SiZLeNSGn","Address":"Fulton Hill 80"}
{"Id":3,"Name":"Clarence Olson","Username":"RachelAdams","Email":"RoseSmith@Browsecat.com","Phone":"988-48-97","Password":"71kuz3gA5w","Address":"Monterey Park 39"}
{"Id":4,"Name":"Gregory Reid","Username":"tButler","Email":"5Moore@Teklist.net","Phone":"520-04-16","Password":"r639qLNu","Address":"Sunfield Park 20"}
{"Id":5,"Name":"Janice Rose","Username":"KeithHart","Email":"nulla@Linktype.com","Phone":"146-91-01","Password":"acSBF5","Address":"Russell Trail 61"}`

	t.Run("find 'com'", func(t *testing.T) {
		result, err := GetDomainStat(bytes.NewBufferString(data), "com")
		require.NoError(t, err)
		require.Equal(t, DomainStat{
			"browsecat.com": 2,
			"linktype.com":  1,
		}, result)
	})

	t.Run("find 'gov'", func(t *testing.T) {
		result, err := GetDomainStat(bytes.NewBufferString(data), "gov")
		require.NoError(t, err)
		require.Equal(t, DomainStat{"browsedrive.gov": 1}, result)
	})

	t.Run("find 'unknown'", func(t *testing.T) {
		result, err := GetDomainStat(bytes.NewBufferString(data), "unknown")
		require.NoError(t, err)
		require.Equal(t, DomainStat{}, result)
	})
}

func TestGetDomainStatErrors(t *testing.T) {
	testData := []struct {
		data          string
		expectedError error
	}{
		{
			data: `{"Id":1,"Name":"Howard Mendoza","Username":"0Oliver","Email":"aliquid_qui_ea@Browsedrive.gov","Phone":"6-866-899-36-79","Password":"InAQJvsq","Address":"Blackbird Place 25"}
{"Id":2,"Name":"Jesse Vasquez","Username":"qRichardson","Email":"mLynch_broWsecat.gov","Phone":"9-373-949-64-00","Password":"SiZLeNSGn","Address":"Fulton Hill 80"}`,
			expectedError: ErrInvalidEmail,
		},

		{
			data: `{{"Id":1,"Name":"Howard Mendoza","Username":"0Oliver","Email":"aliquid_qui_ea@Browsedrive.gov","Phone":"6-866-899-36-79","Password":"InAQJvsq","Address":"Blackbird Place 25"}
{"Id":2,"Name":"Jesse Vasquez","Username":"qRichardson","Email":"mLynch_broWsecat.gov","Phone":"9-373-949-64-00","Password":"SiZLeNSGn","Address":"Fulton Hill 80"}`,
			expectedError: ErrMalformedJSON,
		},
	}

	t.Run("expect errors", func(t *testing.T) {
		for _, tt := range testData {
			result, err := GetDomainStat(strings.NewReader(tt.data), "gov")
			require.Nil(t, result)
			require.ErrorIs(t, err, tt.expectedError)
		}
	})
}

// go test -benchtime=5s -benchmem -bench=BenchmarkGetDomainStat
func BenchmarkGetDomainStat(b *testing.B) {
	data := `{"Id":1,"Name":"Howard Mendoza","Username":"0Oliver","Email":"aliquid_qui_ea@Browsedrive.com","Phone":"6-866-899-36-79","Password":"InAQJvsq","Address":"Blackbird Place 25"}
{"Id":2,"Name":"Justin Oliver Jr. Sr. I II III IV V MD DDS PhD DVM","Username":"oPerez","Email":"MelissaGutierrez@Twinte.com","Phone":"106-05-18","Password":"f00GKr9i","Address":"Oak Valley Lane 19"}
{"Id":3,"Name":"Brian Olson","Username":"non_quia_id","Email":"FrancesEllis@Quinu.com","Phone":"237-75-34","Password":"cmEPhX8","Address":"Butterfield Junction 74"}
{"Id":4,"Name":"Jesse Vasquez Jr. Sr. I II III IV V MD DDS PhD DVM","Username":"qRichardson","Email":"mLynch@Dabtype.name","Phone":"9-373-949-64-00","Password":"SiZLeNSGn","Address":"Fulton Hill 80"}
{"Id":5,"Name":"Clarence Olson","Username":"RachelAdams","Email":"RoseSmith@Browsecat.com","Phone":"988-48-97","Password":"71kuz3gA5w","Address":"Monterey Park 39"}
{"Id":6,"Name":"Gregory Reid","Username":"tButler","Email":"5Moore@Teklist.net","Phone":"520-04-16","Password":"r639qLNu","Address":"Sunfield Park 20"}
{"Id":7,"Name":"Janice Rose","Username":"KeithHart","Email":"nulla@Linktype.com","Phone":"146-91-01","Password":"acSBF5","Address":"Russell Trail 61"}
{"Id":8,"Name":"Jacqueline Young","Username":"CraigKnight","Email":"kCunningham@Skiptube.gov","Phone":"6-954-746-32-77","Password":"rHBCvD5JpLGs","Address":"4th Pass 91"}
{"Id":9,"Name":"Steve Burns","Username":"bRoberts","Email":"perferendis@Skippad.biz","Phone":"246-85-85","Password":"68xyVtL1AaO6","Address":"Jenifer Circle 24"}
{"Id":10,"Name":"Paula Gonzales","Username":"4Ramirez","Email":"BrianBradley@Zoomcast.com","Phone":"363-62-16","Password":"atpnGIr","Address":"Barnett Park 43"}`

	reader := strings.NewReader(data)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		reader.Seek(0, io.SeekStart)
		_, err := GetDomainStat(reader, "com")
		if err != nil {
			b.Error(err)
		}
	}
}
