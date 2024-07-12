package main

// ### Local functions ###

// Find successor for id
func (chord *Chord) findSuccessor(id int) (*Node, error) {
	mLog.Println("Find successor:", id)
	node := chord.localNode
	nodeSuccessor := chord.successor
	var err error
	// get closest node util id is in (node, node.successor]
	for !inRange(id, node.NodeID, nodeSuccessor.NodeID, false, true) {
		mLog.Println("Not inRange of findSuccessor:", id, node.NodeID, nodeSuccessor.NodeID)
		node, err = node.closestPrecedingFinger(id, false)
		if err != nil {
			mLog.Println("Error in closestPrecedingFinger:", err)
			cLog.Println("Error in closestPrecedingFinger:", err)
			return nil, err
		}
		nodeSuccessor, err = node.getSuccessor()
		if err != nil {
			mLog.Println("Error in getSuccessor:", err)
			cLog.Println("Error in getSuccessor:", err)
			return nil, err
		}
	}
	// Return the node's successor
	nodeSuccessor, err = node.getSuccessor()
	if err != nil {
		mLog.Println("Error in getSuccessor:", err)
		cLog.Println("Error in getSuccessor:", err)
		return nil, err
	}
	mLog.Println("The successor is:", nodeSuccessor.NodeID)
	return nodeSuccessor, nil
}

// Find predecessor for id
func (chord *Chord) findPredecessor(id int) (*Node, error) {
	mLog.Println("Find predecessor:", id)
	node := chord.localNode
	nodeSuccessor := chord.successor
	var err error
	// get closest node util id is in [node, node.successor)
	for !inRange(id, node.NodeID, nodeSuccessor.NodeID, true, false) {
		node, err = node.closestPrecedingFinger(id, true)
		if err != nil {
			mLog.Println("Error in closestPrecedingFinger:", err)
			cLog.Println("Error in closestPrecedingFinger:", err)
			return nil, err
		}
		nodeSuccessor, err = node.getSuccessor()
		if err != nil {
			mLog.Println("Error in getSuccessor:", err)
			cLog.Println("Error in getSuccessor:", err)
			return nil, err
		}
	}
	// return node
	mLog.Println("The predecessor is:", node.NodeID)
	return node, nil
}

// Get closest node before id
func (chord *Chord) closestPrecedingFinger(id int, findPredecessor bool) *Node {
	mLog.Println("Find closest preceding finger")
	// Tranverse finger table and get the first finger that is before id
	for i := m; i >= 1; i-- {
		// For find predecessor, change right to true
		if chord.fingerTable[i].node != nil && inRange(chord.fingerTable[i].node.NodeID, chord.localNode.NodeID, id, false, findPredecessor) {
			mLog.Println("Found:", chord.fingerTable[i].node.NodeID)
			return chord.fingerTable[i].node
		}
	}
	mLog.Println("Found local node:", chord.localNode.NodeID)
	return chord.localNode
}

// ### Romote fucntions ###
// To get info from a remote node or let the remote update using rpc

func (node *Node) findSuccessor(id int) (*Node, error) {
	if node.NodeID == chord.localNode.NodeID {
		return chord.findSuccessor(id)
	}
	mLog.Println("Remote find successor")
	args := Args{id, 0, nil, false, ""}
	return node.rpcCall("Chord.FindSuccessor", args)
}

// func (node *Node) findPredecessor(id int) (*Node, error) {
// 	if node.NodeID == chord.localNode.NodeID {
// 		return chord.findPredecessor(id)
// 	}
// 	mLog.Println("Remote find predecessor")
// 	args := Args{id, 0, nil, false, ""}
// 	return node.rpcCall("Chord.FindPredecessor", args)
// }

func (node *Node) closestPrecedingFinger(id int, findPredecessor bool) (*Node, error) {
	if node.NodeID == chord.localNode.NodeID {
		return chord.closestPrecedingFinger(id, findPredecessor), nil
	}
	mLog.Println("Remote closestPrecedingFinger")
	args := Args{id, 0, nil, findPredecessor, ""}
	return node.rpcCall("Chord.ClosestPrecedingFinger", args)
}

func (node *Node) getSuccessor() (*Node, error) {
	if node.NodeID == chord.localNode.NodeID {
		return chord.successor, nil
	}
	mLog.Println("Remote getSuccessor")
	args := Args{0, 0, nil, false, ""}
	return node.rpcCall("Chord.GetSuccessor", args)
}
