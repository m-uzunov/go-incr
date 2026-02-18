package incr

import (
	"fmt"
)

func newRecomputeHeap(maxHeight int) *recomputeHeap {
	return &recomputeHeap{
		heights: make([]queue[INode], maxHeight),
	}
}

type recomputeHeap struct {
	minHeight int
	maxHeight int
	heights   []queue[INode]
	numItems  int
}

func (rh *recomputeHeap) clear() (aborted []INode) {
	aborted = make([]INode, 0, rh.numItems)
	for x := range rh.heights {
		for rh.heights[x].len() > 0 {
			node, _ := rh.heights[x].pop()
			node.Node().heightInRecomputeHeap = HeightUnset
			aborted = append(aborted, node)
		}
	}

	rh.heights = make([]queue[INode], len(rh.heights))
	rh.minHeight = 0
	rh.maxHeight = 0
	rh.numItems = 0
	return
}

func (rh *recomputeHeap) len() int {
	return rh.numItems
}

func (rh *recomputeHeap) add(nodes ...INode) {
	for _, n := range nodes {
		rh.addNodeUnsafe(n)
	}
}

func (rh *recomputeHeap) addIfNotPresent(n INode) {
	if n.Node().heightInRecomputeHeap == HeightUnset {
		rh.addNodeUnsafe(n)
	}
}

func (rh *recomputeHeap) fix(n INode) {
	rh.removeNodeUnsafe(n)
	rh.addNodeUnsafe(n)
}

func (rh *recomputeHeap) has(s INode) bool {
	return s.Node().heightInRecomputeHeap != HeightUnset
}

type recomputeHeapListIter struct {
	nodes []INode
	index int
}

func (i *recomputeHeapListIter) Initialize(nodes []INode) {
	i.nodes = nodes
	i.index = 0
}

func (i *recomputeHeapListIter) Next() (INode, bool) {
	if i.index >= len(i.nodes) {
		return nil, false
	}
	node := i.nodes[i.index]
	i.index++
	node.Node().heightInRecomputeHeap = HeightUnset
	return node, true
}

type RecomputeHeapListIterator interface {
	Initialize([]INode)
	Next() (INode, bool)
}

func (rh *recomputeHeap) setIterToMinHeight(iter RecomputeHeapListIterator) {
	if iter == nil {
		return
	}

	for x := 0; x < len(rh.heights); x++ {
		if rh.heights[x].len() > 0 {
			items := rh.heights[x].values()
			rh.numItems -= rh.heights[x].size
			rh.heights[x].clear()
			iter.Initialize(items)
			rh.minHeight = rh.nextMinHeightUnsafe()
			return
		}
	}
}

func (rh *recomputeHeap) remove(node INode) {
	rh.removeNodeUnsafe(node)
}

//
// utils
//

func (rh *recomputeHeap) removeMinUnsafe() (node INode, ok bool) {
	for x := rh.minHeight; x <= rh.maxHeight; x++ {
		if rh.heights[x].len() > 0 {
			node, ok = rh.heights[x].pop()
			rh.numItems--
			node.Node().heightInRecomputeHeap = HeightUnset
			if rh.heights[x].len() > 0 {
				rh.minHeight = x
			} else {
				rh.minHeight = rh.nextMinHeightUnsafe()
			}
			return
		}
	}
	return
}

func (rh *recomputeHeap) addNodeUnsafe(s INode) {
	sn := s.Node()
	height := sn.height
	sn.heightInRecomputeHeap = height
	rh.maybeUpdateMinMaxHeightsUnsafe(height)
	rh.maybeAddNewHeightsUnsafe(height)
	rh.heights[height].push(s)
	rh.numItems++
}

func (rh *recomputeHeap) removeNodeUnsafe(item INode) {
	rh.numItems--
	id := item.Node().id
	height := item.Node().heightInRecomputeHeap
	q := &rh.heights[height]
	n := q.size
	for i := 0; i < n; i++ {
		node, _ := q.pop()
		if node.Node().id != id {
			q.push(node)
		}
	}
	isLastAtHeight := q.len() == 0
	if height == rh.minHeight && isLastAtHeight {
		rh.minHeight = rh.nextMinHeightUnsafe()
	}
	item.Node().heightInRecomputeHeap = HeightUnset
}

func (rh *recomputeHeap) maybeUpdateMinMaxHeightsUnsafe(newHeight int) {
	if rh.numItems == 0 {
		rh.minHeight = newHeight
		rh.maxHeight = newHeight
		return
	}
	if rh.minHeight > newHeight {
		rh.minHeight = newHeight
	}
	if rh.maxHeight < newHeight {
		rh.maxHeight = newHeight
	}
}

func (rh *recomputeHeap) maybeAddNewHeightsUnsafe(newHeight int) {
	if len(rh.heights) <= newHeight {
		required := (newHeight - len(rh.heights)) + 1
		for x := 0; x < required; x++ {
			rh.heights = append(rh.heights, queue[INode]{})
		}
	}
}

func (rh *recomputeHeap) nextMinHeightUnsafe() (next int) {
	if rh.numItems == 0 {
		return
	}
	for x := 0; x < len(rh.heights); x++ {
		if rh.heights[x].len() > 0 {
			next = x
			break
		}
	}
	return
}

// sanityCheck loops through each item in each height block
// and checks that all the height values match.
func (rh *recomputeHeap) sanityCheck() error {
	if rh.numItems > 0 && rh.heights[rh.minHeight].len() == 0 {
		return fmt.Errorf("recompute heap; sanity check; lookup has items but min height block is empty")
	}
	for heightIndex := range rh.heights {
		for _, node := range rh.heights[heightIndex].values() {
			if node.Node().heightInRecomputeHeap != heightIndex {
				return fmt.Errorf("recompute heap; sanity check; at height %d item has height %d", heightIndex, node.Node().heightInRecomputeHeap)
			}
			if node.Node().heightInRecomputeHeap != node.Node().height {
				return fmt.Errorf("recompute heap; sanity check; at height %d item has height %d and node has height %d", heightIndex, node.Node().heightInRecomputeHeap, node.Node().height)
			}
		}
	}
	return nil
}
