package awareness

import (
	"sync"
	"time"

	"github.com/named-data/ndnd/repo/tlv"
)

type RepoAwarenessStore struct {
	mutex sync.RWMutex

	nodeStates map[string]*RepoNodeAwareness

	upNodes      *LRU[string, *RepoNodeAwareness]
	pfailedNodes *LRU[string, *RepoNodeAwareness]
	failedNodes  *LRU[string, *RepoNodeAwareness]

	// Callbacks for node state changes
	onNodeUp        func(*RepoNodeAwareness)
	onNodePFailed   func(*RepoNodeAwareness)
	onNodeFailed    func(*RepoNodeAwareness)
	onNodeForgotten func(*RepoNodeAwareness)
}

func NewRepoAwarenessStore() *RepoAwarenessStore {
	return &RepoAwarenessStore{
		nodeStates:   make(map[string]*RepoNodeAwareness),
		upNodes:      NewLRU[string, *RepoNodeAwareness](),
		pfailedNodes: NewLRU[string, *RepoNodeAwareness](),
		failedNodes:  NewLRU[string, *RepoNodeAwareness](),
	}
}

func (s *RepoAwarenessStore) String() string {
	return "repo-awareness-store"
}

// GetNode retrieves a node's awareness by its name.
// Returns nil if the node does not exist.
func (s *RepoAwarenessStore) GetNode(name string) *RepoNodeAwareness {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	node := s.nodeStates[name]
	return node
}

func (s *RepoAwarenessStore) ProcessUpdate(update *tlv.AwarenessUpdate) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	name := update.NodeName
	node := s.nodeStates[name]
	if node == nil { // initialize the node state
		node = &RepoNodeAwareness{
			name:       name,
			lastKnown:  time.Now(),
			partitions: update.Partitions,
			state:      Up,
		}
		s.nodeStates[name] = node
		s.upNodes.Put(name, node, true)
		return
	}

	// Update the node's partitions and reset its state to Up
	node.Update(update.Partitions)
	if node.state != Up {
		// Reset state to Up when updating partitions
		node.state = Up
		s.upNodes.Put(name, node, true)

		// Remove from the previous state LRU
		s.Pos(node).Remove(name)

		// Call the onNodeUp callback if set
		s.onNodeUp(node)
	}
}

func (s *RepoAwarenessStore) Pos(node *RepoNodeAwareness) *LRU[string, *RepoNodeAwareness] {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	switch node.state {
	case Up:
		return s.upNodes
	case PFailed:
		return s.pfailedNodes
	case Failed:
		return s.failedNodes
	case Forgotten:
		return nil // Forgotten nodes are not tracked in any LRU
	default:
		return nil // Unknown state, return nil
	}
}

// Callbacks for node state changes
func (s *RepoAwarenessStore) SetOnNodeUp(callback func(*RepoNodeAwareness)) {
	s.onNodeUp = callback
}

func (s *RepoAwarenessStore) SetOnNodePFailed(callback func(*RepoNodeAwareness)) {
	s.onNodePFailed = callback
}

func (s *RepoAwarenessStore) SetOnNodeFailed(callback func(*RepoNodeAwareness)) {
	s.onNodeFailed = callback
}

func (s *RepoAwarenessStore) SetOnNodeForgotten(callback func(*RepoNodeAwareness)) {
	s.onNodeForgotten = callback
}

// TODO: handle expiration and state transitions of nodes

// TODO: this struct should also handle partition underreplication scenario, probably through handlers
