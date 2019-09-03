package search

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestCheckFileName(t *testing.T) {
	a := assert.New(t)

	from, err := time.Parse("2006-01-02T15:04:05.999-07:00", "2019-06-10T10:10:51.964-00:00")
	a.NoError(err)
	to, err := time.Parse("2006-01-02T15:04:05.999-07:00", "2019-06-10T13:13:51.964-00:00")
	a.NoError(err)
	filter := Filter{
		from: from,
		to:   to,
	}
	filesArray := []string{
		"127.0.0.1__2019-06-10T08-10-51.964.log",
		"127.0.0.1__2019-06-10T13-10-51.964.log",
		"127.0.0.1__2019-06-10T13-13-51.964.log",
	}

	ok, stop, err := checkFileName(filesArray[0], filter)
	a.NoError(err)
	a.False(ok)
	a.False(stop)

	ok, stop, err = checkFileName(filesArray[1], filter)
	a.NoError(err)
	a.True(ok)
	a.False(stop)

	ok, stop, err = checkFileName(filesArray[2], filter)
	a.NoError(err)
	a.True(ok)
	a.True(stop)
}
