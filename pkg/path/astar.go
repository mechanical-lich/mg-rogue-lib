// Package path provides an allocation-efficient, graph-centric A* pathfinding
// implementation. The graph supplies neighbor IDs, costs, and heuristics using
// plain integer node IDs — so the internal node array contains no pointer fields
// and the GC never needs to scan it, regardless of how many nodes are explored.
package path

// Graph provides the pathfinding topology using integer node IDs.
// Level in rlworld implements this interface; each chunk can be its own Graph.
type Graph interface {
	// PathNeighborIDs appends the IDs of walkable neighbors of nodeID to buf
	// and returns the result. Uses the append pattern to avoid allocation.
	PathNeighborIDs(nodeID int, buf []int) []int

	// PathCost returns the movement cost between two adjacent node IDs.
	PathCost(fromID, toID int) float64

	// PathEstimate returns the heuristic estimate of remaining cost
	// from nodeID to goalID (e.g. squared Euclidean distance).
	PathEstimate(nodeID, goalID int) float64
}

// node holds A* bookkeeping for a single graph node.
// All fields are plain integers/floats — zero pointer fields — so the GC
// skips the entire nodes slice during collection.
//
// Memory layout (64-bit):
//
//	cost   float64  @ 0   (8 bytes)
//	rank   float64  @ 8   (8 bytes)
//	parent int32    @ 16  (4 bytes)  -1 = no parent
//	index  int32    @ 20  (4 bytes)  position in heap, -1 = not in heap
//	id     int32    @ 24  (4 bytes)  graph node ID (tile flat index)
//	open   bool     @ 28  (1 byte)
//	closed bool     @ 29  (1 byte)
//	_      [2]byte  @ 30  (padding)
//	Total: 32 bytes, GC-invisible
type node struct {
	cost   float64
	rank   float64
	parent int32
	index  int32
	id     int32
	open   bool
	closed bool
	_      [2]byte // explicit padding
}

// AStar is a reusable A* pathfinder that minimises allocations.
// For concurrent use, create separate instances via NewAStar.
// Each graph (level, chunk) can share one instance sequentially — it fully
// resets between calls.
type AStar struct {
	nodeIndex   map[int]int // nodeID → index into nodes
	nodes       []node      // GC-invisible: no pointer fields
	openSet     []int32     // heap of node indices
	neighborBuf []int       // reused neighbor ID buffer
	result      []int       // reused result (node IDs, start→goal)
}

// NewAStar creates a new AStar pathfinder with pre-allocated capacity.
// estimatedNodes is the expected number of nodes to explore per call.
func NewAStar(estimatedNodes int) *AStar {
	if estimatedNodes < 64 {
		estimatedNodes = 64
	}
	return &AStar{
		nodeIndex:   make(map[int]int, estimatedNodes),
		nodes:       make([]node, 0, estimatedNodes),
		openSet:     make([]int32, 0, estimatedNodes/4),
		neighborBuf: make([]int, 0, 8),
		result:      make([]int, 0, 64),
	}
}

func (a *AStar) reset() {
	if len(a.nodeIndex) > cap(a.nodes) {
		a.nodeIndex = make(map[int]int, cap(a.nodes))
	} else {
		for k := range a.nodeIndex {
			delete(a.nodeIndex, k)
		}
	}
	a.nodes = a.nodes[:0]
	a.openSet = a.openSet[:0]
	a.neighborBuf = a.neighborBuf[:0]
	a.result = a.result[:0]
}

func (a *AStar) getOrCreate(id int) int {
	if idx, ok := a.nodeIndex[id]; ok {
		return idx
	}
	idx := len(a.nodes)
	a.nodes = append(a.nodes, node{
		id:     int32(id),
		cost:   1e18,
		rank:   1e18,
		parent: -1,
		index:  -1,
	})
	a.nodeIndex[id] = idx
	return idx
}

func (a *AStar) heapPush(ni int) {
	n := &a.nodes[ni]
	n.index = int32(len(a.openSet))
	a.openSet = append(a.openSet, int32(ni))
	a.heapUp(int(n.index))
}

func (a *AStar) heapPop() int {
	if len(a.openSet) == 0 {
		return -1
	}
	top := int(a.openSet[0])
	last := len(a.openSet) - 1
	a.heapSwap(0, last)
	a.openSet = a.openSet[:last]
	if len(a.openSet) > 0 {
		a.heapDown(0)
	}
	a.nodes[top].index = -1
	return top
}

func (a *AStar) heapRemove(heapIdx int) {
	last := len(a.openSet) - 1
	if heapIdx != last {
		a.heapSwap(heapIdx, last)
		a.openSet = a.openSet[:last]
		if heapIdx < len(a.openSet) {
			a.heapDown(heapIdx)
			a.heapUp(heapIdx)
		}
	} else {
		a.openSet = a.openSet[:last]
	}
}

func (a *AStar) heapUp(idx int) {
	for idx > 0 {
		parent := (idx - 1) / 2
		if a.nodes[a.openSet[parent]].rank <= a.nodes[a.openSet[idx]].rank {
			break
		}
		a.heapSwap(parent, idx)
		idx = parent
	}
}

func (a *AStar) heapDown(idx int) {
	n := len(a.openSet)
	for {
		smallest := idx
		left := 2*idx + 1
		right := 2*idx + 2
		if left < n && a.nodes[a.openSet[left]].rank < a.nodes[a.openSet[smallest]].rank {
			smallest = left
		}
		if right < n && a.nodes[a.openSet[right]].rank < a.nodes[a.openSet[smallest]].rank {
			smallest = right
		}
		if smallest == idx {
			break
		}
		a.heapSwap(idx, smallest)
		idx = smallest
	}
}

func (a *AStar) heapSwap(i, j int) {
	a.openSet[i], a.openSet[j] = a.openSet[j], a.openSet[i]
	a.nodes[a.openSet[i]].index = int32(i)
	a.nodes[a.openSet[j]].index = int32(j)
}

// Path finds the shortest path from node 'from' to node 'to' through graph.
// Returns the path as a slice of node IDs (start→goal), total cost, and whether
// a path was found. The returned slice is reused between calls — copy it if you
// need to keep it past the next Path call.
//
// Node IDs are graph-defined integers (for tile levels: the flat tile index).
// Resolve IDs back to tiles with level.GetTilePtrIndex(id).
func (a *AStar) Path(graph Graph, from, to int) (path []int, distance float64, found bool) {
	a.reset()

	startIdx := a.getOrCreate(from)
	a.nodes[startIdx].cost = 0
	a.nodes[startIdx].rank = graph.PathEstimate(from, to)
	a.nodes[startIdx].open = true
	a.heapPush(startIdx)

	for len(a.openSet) > 0 {
		currentIdx := a.heapPop()
		current := &a.nodes[currentIdx]
		current.open = false
		current.closed = true

		if int(current.id) == to {
			// Reconstruct path by walking parent chain
			a.result = a.result[:0]
			idx := int32(currentIdx)
			for idx != -1 {
				a.result = append(a.result, int(a.nodes[idx].id))
				idx = a.nodes[idx].parent
			}
			for i, j := 0, len(a.result)-1; i < j; i, j = i+1, j-1 {
				a.result[i], a.result[j] = a.result[j], a.result[i]
			}
			return a.result, current.cost, true
		}

		a.neighborBuf = graph.PathNeighborIDs(int(current.id), a.neighborBuf[:0])
		for _, neighborID := range a.neighborBuf {
			cost := current.cost + graph.PathCost(int(current.id), neighborID)
			neighborIdx := a.getOrCreate(neighborID)
			neighborNode := &a.nodes[neighborIdx]

			if cost < neighborNode.cost {
				if neighborNode.open {
					a.heapRemove(int(neighborNode.index))
					neighborNode.open = false
				}
				neighborNode.closed = false
			}

			if !neighborNode.open && !neighborNode.closed {
				neighborNode.cost = cost
				neighborNode.rank = cost + graph.PathEstimate(neighborID, to)
				neighborNode.parent = int32(currentIdx)
				neighborNode.open = true
				a.heapPush(neighborIdx)
			}
		}
	}

	return nil, 0, false
}

// Default is a shared AStar instance for single-threaded convenience.
// For concurrent pathfinding, create separate instances with NewAStar.
var Default = NewAStar(1024)

// Path is a convenience function using the default AStar instance.
// Not safe for concurrent use.
func Path(graph Graph, from, to int) ([]int, float64, bool) {
	return Default.Path(graph, from, to)
}
