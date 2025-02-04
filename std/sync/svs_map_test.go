package sync_test

import (
	"testing"

	ndn_sync "github.com/named-data/ndnd/std/sync"
	tu "github.com/named-data/ndnd/std/utils/testutils"
	"github.com/stretchr/testify/require"
)

func makeSvMap() ndn_sync.SvMap[uint64] {
	m := ndn_sync.NewSvMap[uint64](0)
	m.Set("/ndn/alice", 100, 1)
	m.Set("/ndn/alice", 200, 4)
	m.Set("/ndn/bob", 150, 3)
	return m
}

func TestSvMapBasic(t *testing.T) {
	tu.SetT(t)

	m := makeSvMap()

	// Basic entries
	require.Equal(t, uint64(1), m.Get("/ndn/alice", 100))
	require.Equal(t, uint64(4), m.Get("/ndn/alice", 200))
	require.Equal(t, uint64(3), m.Get("/ndn/bob", 150))

	// Empty entries
	require.Equal(t, uint64(0), m.Get("/ndn/bob", 100))
	require.Equal(t, uint64(0), m.Get("/ndn/cathy", 100))

	// Update entries
	m.Set("/ndn/bob", 150, 5)
	require.Equal(t, uint64(5), m.Get("/ndn/bob", 150))

	// Test editing the value directly
	m["/ndn/alice"][0].Value = 138
	require.Equal(t, uint64(138), m.Get("/ndn/alice", 100))
}

func TestSvMapSet(t *testing.T) {
	tu.SetT(t)

	m := makeSvMap()

	// Set new
	m.Set("/ndn/alice", 120, 138)
	require.Equal(t, uint64(138), m.Get("/ndn/alice", 120))

	// Set existing
	m.Set("/ndn/alice", 120, 190)
	require.Equal(t, uint64(190), m.Get("/ndn/alice", 120))

	// Set ordering
	m.Set("/ndn/alice", 110, 138)
	require.Equal(t, uint64(138), m.Get("/ndn/alice", 110))
	boots := []uint64{100, 110, 120, 200}
	for i, entry := range m["/ndn/alice"] {
		require.Equal(t, boots[i], entry.Boot)
	}
}

func TestSvMapNewer(t *testing.T) {
	tu.SetT(t)

	m1 := makeSvMap()
	m2 := makeSvMap()

	exist := func(_, _ uint64) bool { return false }
	order := func(a, b uint64) bool { return a > b }

	// Equal
	require.False(t, m1.IsNewerThan(m2, order))
	require.False(t, m1.IsNewerThan(m2, exist))

	// Different sequence number
	m2.Set("/ndn/alice", 200, 99)
	require.True(t, m2.IsNewerThan(m1, order))
	require.False(t, m2.IsNewerThan(m1, exist))
	require.False(t, m1.IsNewerThan(m2, order))
	require.False(t, m1.IsNewerThan(m2, exist))

	// Different entry exist
	m2.Set("/ndn/cathy", 100, 99)
	require.True(t, m2.IsNewerThan(m1, order))
	require.True(t, m2.IsNewerThan(m1, exist))
	require.False(t, m1.IsNewerThan(m2, order))
	require.False(t, m1.IsNewerThan(m2, exist))

	// Both are new (m1 seq only)
	m1.Set("/ndn/bob", 150, 99)
	require.True(t, m2.IsNewerThan(m1, order))
	require.True(t, m2.IsNewerThan(m1, exist))
	require.True(t, m1.IsNewerThan(m2, order))
	require.False(t, m1.IsNewerThan(m2, exist))
}

func TestSvMapTLV(t *testing.T) {
	tu.SetT(t)

	// Add entries to test ordering
	m := ndn_sync.NewSvMap[uint64](0)
	m.Set("/ndn/alice", 100, 1)
	m.Set("/ndn/alice", 200, 4)
	m.Set("/ndn/cathy", 150, 3)
	m.Set("/ndn/bob", 150, 3)
	m.Set("/ndn/bob", 50, 5)
	sv := m.Encode(func(s uint64) uint64 { return s })

	// Name Ordering should be in NDN canonical order.
	// Bootstrap time is in ascending order.
	// https://docs.named-data.net/NDN-packet-spec/current/name.html#canonical-order

	bob := sv.Entries[0]
	require.Equal(t, "/ndn/bob", bob.Name.String())
	require.Equal(t, uint64(50), bob.SeqNoEntries[0].BootstrapTime)
	require.Equal(t, uint64(5), bob.SeqNoEntries[0].SeqNo)
	require.Equal(t, uint64(150), bob.SeqNoEntries[1].BootstrapTime)
	require.Equal(t, uint64(3), bob.SeqNoEntries[1].SeqNo)

	alice := sv.Entries[1]
	require.Equal(t, "/ndn/alice", alice.Name.String())
	require.Equal(t, uint64(100), alice.SeqNoEntries[0].BootstrapTime)
	require.Equal(t, uint64(1), alice.SeqNoEntries[0].SeqNo)
	require.Equal(t, uint64(200), alice.SeqNoEntries[1].BootstrapTime)
	require.Equal(t, uint64(4), alice.SeqNoEntries[1].SeqNo)

	cathy := sv.Entries[2]
	require.Equal(t, "/ndn/cathy", cathy.Name.String())
	require.Equal(t, uint64(150), cathy.SeqNoEntries[0].BootstrapTime)
	require.Equal(t, uint64(3), cathy.SeqNoEntries[0].SeqNo)
}
