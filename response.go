package main

// ### Response for rpc calls ###

// Response for find
func (chord *Chord) FindSuccessor(args Args, reply *Node) error { // Args need to be exported
	mLog.Println()
	mLog.Println("Get call to find sucessor for:", args.Arg1)
	node, err := chord.findSuccessor(args.Arg1)
	if err != nil {
		mLog.Println("Error in findSuccessor:", err)
		cLog.Println("Error in findSuccessor:", err)
		return err
	}
	reply.NodeID = node.NodeID
	reply.Ip = node.Ip
	reply.Port = node.Port
	// chord.printState()
	return nil
}

func (chord *Chord) FindPredecessor(args Args, reply *Node) error {
	mLog.Println()
	mLog.Println("Get call to find predecessor for:", args.Arg1)
	node, err := chord.findPredecessor(args.Arg1)
	if err != nil {
		mLog.Println("Error in findPredecessor:", err)
		cLog.Println("Error in findPredecessor:", err)
		return err
	}
	reply.NodeID = node.NodeID
	reply.Ip = node.Ip
	reply.Port = node.Port
	return nil
}

func (chord *Chord) ClosestPrecedingFinger(args Args, reply *Node) error {
	mLog.Println()
	mLog.Println("Get call to find closest preceding finger for:", args.Arg1)
	node := chord.closestPrecedingFinger(args.Arg1, args.Arg4)
	reply.NodeID = node.NodeID
	reply.Ip = node.Ip
	reply.Port = node.Port
	return nil
}

func (chord *Chord) GetSuccessor(args Args, reply *Node) error {
	mLog.Println()
	mLog.Println("Get call to get successor")
	mLog.Println("Get sucessor:", chord.successor.NodeID)
	node := chord.successor
	reply.NodeID = node.NodeID
	reply.Ip = node.Ip
	reply.Port = node.Port
	return nil
}

// Response for join
func (chord *Chord) UpdateFingerTable(args Args, reply *Node) error {
	mLog.Println()
	mLog.Println("Get call to update finger table", args.Arg1)
	resNode := chord.updateFingerTable(args.Arg3, args.Arg1)
	if resNode != nil {
		reply.NodeID = resNode.NodeID
		reply.Ip = resNode.Ip
		reply.Port = resNode.Port
	} else {
		mLog.Println("No continue update")
		reply = nil
	}
	return nil
}

func (chord *Chord) GetPredecessor(args Args, reply *Node) error {
	mLog.Println()
	mLog.Println("Get call to get predecessor")
	if chord.predecessor != nil {
		mLog.Println("Get predecessor:", chord.predecessor.NodeID)
		node := chord.predecessor
		reply.NodeID = node.NodeID
		reply.Ip = node.Ip
		reply.Port = node.Port
	} else {
		mLog.Println("Get predecessor: nil")
		reply = nil
	}
	return nil
}

func (chord *Chord) SetPredecessor(args Args, reply *Node) error {
	mLog.Println()
	mLog.Println("Get call to set predecessor")
	chord.predecessor = args.Arg3
	if chord.predecessor != nil {
		mLog.Println("Set predecessor to:", chord.predecessor.NodeID)
	} else {
		mLog.Println("Set predecessor to: nil")
	}

	err := chord.allKeysBackup()
	if err != nil {
		sLog.Print("Fail to backup all keys:", err)
	}
	return nil
}

// Response for stabilize
func (chord *Chord) Notify(args Args, reply *Node) error {
	sLog.Println()
	sLog.Println("Get call to handle notify")
	chord.notify(args.Arg3)
	return nil
}

func (chord *Chord) UpdateKeys(args Args, reply *bool) error {

	mLog.Println()
	mLog.Println("Get call to update the keys")
	chord.updateKeys(args.Arg1, args.Arg3)
	return nil
}

// Response for command functions
func (chord *Chord) IsLocalFileExist(args Args, reply *bool) error {
	mLog.Println()
	mLog.Println("Get call to tell if file:", args.Arg5, "exist")
	*reply = chord.isLocalFileExist(args.Arg5)
	return nil
}
