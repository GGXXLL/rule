package dto

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestPayload_HoursAgo(t *testing.T) {
	p := &Payload{}
	assert.Equal(t, p.HoursAgo("2021-01-01 00:00:00"),
		int(time.Since(time.Date(
			2021,
			01,
			01,
			0,
			0,
			0,
			0,
			time.Local,
		)).Hours()))
}

func TestPayload_MinutesAgo(t *testing.T) {
	p := &Payload{}
	assert.Equal(t, p.MinutesAgo("2021-01-01 00:00:00"),
		int(time.Since(time.Date(
			2021,
			01,
			01,
			0,
			0,
			0,
			0,
			time.Local,
		)).Minutes()))
}
