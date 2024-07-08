package main

import (
	"math"
)

// node creatation
func (chord *Chord) create() {
	println("Chord create")
	println("ts:", ts, "tff:", tff, "tcp:", tcp)
	mLog.Println("Chord create")
	// set all nodes to self
	chord.predecessor = chord.localNode
	for i := 1; i <= m; i++ {
		chord.fingerTable[i] = &Entry{}
		chord.fingerTable[i].start = (chord.localNode.NodeID + int(math.Pow(2, float64(i-1)))) % int(math.Pow(2, float64(m)))
		chord.fingerTable[i].node = chord.localNode
	}
	chord.successor = chord.fingerTable[1].node
	chord.maintainSuccessorList()
	chord.printState()
}

// node join
func (chord *Chord) join(node Node) {
	var err error
	println("Chord join")
	mLog.Println("Chord join")

	// Init finger table
	err = chord.initFingerTable(node)
	if err != nil {
		mLog.Println("Fail to init finger table:", err)
		cLog.Println("Fail to init finger table:", err)
		return
	}

	// Update others' fingers
	err = chord.updateOthers()
	if err != nil {
		mLog.Println("Fail to update others:", err)
		cLog.Println("Fail to update others:", err)
		return
	}

	// Maintain successor list and let successor update keys
	chord.maintainSuccessorList()
	chord.successor.updateKeys(chord.predecessor.NodeID, chord.localNode)
}

// Init finger table
func (chord *Chord) initFingerTable(node Node) error {
	mLog.Println("Init finger table")
	for i := 1; i <= m; i++ {
		chord.fingerTable[i] = &Entry{}
		chord.fingerTable[i].start = (chord.localNode.NodeID + int(math.Pow(2, float64(i-1)))) % int(math.Pow(2, float64(m)))
	}

	// Set successor and predecessor and update sucessor
	var err error
	chord.fingerTable[1].node, err = node.findSuccessor(chord.fingerTable[1].start)
	if err != nil {
		mLog.Println("Error in findSuccessor:", err)
		cLog.Println("Error in findSuccessor:", err)
		return err
	}
	chord.successor = chord.fingerTable[1].node
	mLog.Println("Set sucessor:", chord.successor.NodeID)
	chord.predecessor, err = chord.successor.getPredecessor()
	if err != nil {
		mLog.Println("Error in getPredecessor:", err)
		cLog.Println("Error in getPredecessor:", err)
		return err
	}
	if chord.predecessor != nil {
		mLog.Println("Set predecessor:", chord.predecessor.NodeID)
	} else {
		mLog.Println("Set predecessor: nil")
	}
	chord.successor.setPredecessor(chord.localNode)
	mLog.Println("Let successor set predecessor as:", chord.localNode.NodeID)

	// Set fingers' nodes
	mLog.Println("Update finger table nodes")
	for i := 1; i <= m-1; i++ {
		// not beyond last finger
		if inRange(chord.fingerTable[i+1].start, chord.localNode.NodeID, chord.fingerTable[i].node.NodeID, true, true) && chord.fingerTable[i+1].start != chord.fingerTable[i].node.NodeID {
			chord.fingerTable[i+1].node = chord.fingerTable[i].node
		} else { // beyond last finger
			chord.fingerTable[i+1].node, err = node.findSuccessor(chord.fingerTable[i+1].start)
			if err != nil {
				mLog.Println("Error in findSuccessor:", err)
				cLog.Println("Error in findSuccessor:", err)
				return err
			}
		}
		mLog.Println("start:", chord.fingerTable[i+1].start, "node id:", chord.fingerTable[i+1].node.NodeID)
	}
	return nil
}

// Update others' fingers
func (chord *Chord) updateOthers() error {
	mLog.Println("Update others")
	// find possible node which may need to update finger i
	for i := 1; i <= m; i++ {
		mLog.Println(i)
		p, err := chord.findPredecessor(minusInRing(chord.localNode.NodeID, int(math.Pow(2, float64(i-1)))))
		if err != nil {
			mLog.Println("Error in findPredecessor:", err)
			cLog.Println("Error in findPredecessor:", err)
			return err
		}
		mLog.Println("Posible predecessor:", p.NodeID)
		resNode, err := p.updateFingerTable(chord.localNode, i)
		if err != nil {
			mLog.Println("Error in updateFingerTable:", err)
			cLog.Println("Error in updateFingerTable:", err)
			return err
		}
		for resNode != nil && len(resNode.Ip) > 5 { // if resnode update, continue to update its predecessor
			mLog.Println(resNode.NodeID, resNode.Ip, resNode.Port)
			mLog.Println("Continue to update its predecessor:", resNode.NodeID)
			resNode, err = resNode.updateFingerTable(chord.localNode, i)
			if err != nil {
				mLog.Println("Error in updateFingerTable:", err)
				cLog.Println("Error in updateFingerTable:", err)
				return err
			}
		}
	}
	return nil
}

// Handle update finger table request
func (chord *Chord) updateFingerTable(sNode *Node, i int) *Node {
	var resNode *Node = nil
	mLog.Println("Update finger table")
	// check if the new joined node sNode is before finger i, if so, update
	if chord.fingerTable[i].node != nil && inRange(sNode.NodeID, chord.localNode.NodeID, chord.fingerTable[i].node.NodeID, false, false) && sNode.NodeID != chord.fingerTable[i].node.NodeID {
		mLog.Println("In range, need to be modified:", sNode.NodeID, chord.localNode.NodeID, chord.fingerTable[i].node.NodeID)
		chord.fingerTable[i].node = sNode
		if i == 1 {
			chord.successor = chord.fingerTable[i].node
		}
		mLog.Println("Set finger table", i, "node id to", sNode.NodeID)
		resNode = chord.predecessor
		if resNode != nil {
			mLog.Println("Predecessor may need to be updated:", resNode.NodeID)
		} else {
			mLog.Println("Predecessor is nil, no continue update")
		}
	}
	if chord.fingerTable[i].node != nil {
		mLog.Println("Not in range, don't need to be modified:", sNode.NodeID, chord.localNode.NodeID, chord.fingerTable[i].node.NodeID)
	} else {
		mLog.Println("finger", i, "node is nil, need to finish join first")
	}
	// if update return predecessor, otherwise return nil
	return resNode
}

// ### Romote fucntions ###
// To get info from a remote node or let the remote update using rpc

func (node *Node) updateFingerTable(sNode *Node, i int) (*Node, error) {
	if node.NodeID == chord.localNode.NodeID {
		return chord.updateFingerTable(sNode, i), nil
	}
	mLog.Println("Remote updateFingerTable")
	args := Args{i, 0, sNode, false, ""}
	return node.rpcCall("Chord.UpdateFingerTable", args)
}

func (node *Node) getPredecessor() (*Node, error) {
	if node.NodeID == chord.localNode.NodeID {
		return chord.predecessor, nil
	}
	mLog.Println("Remote get predecessor from", node.NodeID)
	args := Args{0, 0, nil, false, ""}
	return node.rpcCall("Chord.GetPredecessor", args)
}

func (node *Node) setPredecessor(predecessor *Node) {
	if node.NodeID == chord.localNode.NodeID {
		chord.predecessor = predecessor
		return
	}
	mLog.Println("Remote set predecessor")
	args := Args{0, 0, predecessor, false, ""}
	node.rpcCall("Chord.SetPredecessor", args)
}
