package utils_test

import (
	"testing"
	"time"

	"github.com/named-data/ndnd/std/utils"
	tu "github.com/named-data/ndnd/std/utils/testutils"
	"github.com/stretchr/testify/require"
)

func TestIdPtr(t *testing.T) {
	tu.SetT(t)

	p := utils.IdPtr(uint64(42))
	require.Equal(t, uint64(42), *p)
}

func TestMakeTimestamp(t *testing.T) {
	tu.SetT(t)

	date := time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC)
	require.Equal(t, uint64(1609459200000), utils.MakeTimestamp(date))
}

func TestConvertNonce(t *testing.T) {
	tu.SetT(t)

	nonce := []byte{0x01, 0x02, 0x03, 0x04}
	val := utils.ConvertNonce(nonce)
	require.Equal(t, uint32(0x01020304), val.Unwrap())

	nonce = []byte{0x42, 0x1C, 0xE1, 0x4B}
	val = utils.ConvertNonce(nonce)
	require.Equal(t, uint32(0x421ce14b), val.Unwrap())
}

func TestHeaderEqual(t *testing.T) {
	tu.SetT(t)

	a := []int{1, 2, 3, 4, 5, 6}
	b := []int{1, 2, 3, 4, 5, 6}
	c := []int{1, 2, 3, 4, 5, 6, 7}
	require.True(t, utils.HeaderEqual(a, a))
	require.False(t, utils.HeaderEqual(a, b))
	require.False(t, utils.HeaderEqual(a, c))

	d := a[1:4]
	e := a[1:4]
	f := a[1:3]
	require.True(t, utils.HeaderEqual(d, d))
	require.True(t, utils.HeaderEqual(d, e))
	require.False(t, utils.HeaderEqual(d, f))
	require.False(t, utils.HeaderEqual(e, f))
	require.False(t, utils.HeaderEqual(a, f))

	g := a[2:5]
	h := a[1:4]
	require.False(t, utils.HeaderEqual(g, h))
}
