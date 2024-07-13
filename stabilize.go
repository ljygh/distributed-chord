package main

import (
	"net/rpc"
	"os"
	"strconv"
)

// ### Local functions ###

// Stabilize
func (chord *Chord) stabilize() error {
	sLog.Println()
	sLog.Println("Stabilize")

	// check if successor is alive
	err := chord.checkSuccessor()
	if err != nil {
		sLog.Println("Fail to check successor:", err)
		mLog.Println("Fail to check successor:", err)
		println("Fail to check successor:", err)
		return err
	}

	// maintain successor list
	err = chord.maintainSuccessorList()
	if err != nil {
		sLog.Println("Fail to maintain successor list:", err)
		mLog.Println("Fail to maintain successor list:", err)
		println("Fail to maintain successor list:", err)
		return err
	}

	// check if there is a new joined node between local node and successor, if so, set it as successor
	x, err := chord.successor.getPredecessor()
	if err != nil {
		sLog.Println("Error in getSuccessor:", err)
		mLog.Println("Error in getSuccessor:", err)
		return err
	}
	if x != nil {
		sLog.Println("Tell if insert successor", x.NodeID, chord.localNode.NodeID, chord.successor.NodeID)
		sLog.Printf("In range or not:%v", inRange(x.NodeID, chord.localNode.NodeID, chord.successor.NodeID, false, false))
	}
	if x != nil && inRange(x.NodeID, chord.localNode.NodeID, chord.successor.NodeID, false, false) && x.NodeID != chord.localNode.NodeID && x.NodeID != chord.successor.NodeID {
		sLog.Println("Stabilize inRange:", x.NodeID, chord.localNode.NodeID, chord.successor.NodeID)
		if chord.checkNode(x) {
			chord.successor = x
			chord.fingerTable[1].node = x
			sLog.Println("Update successor to:", chord.successor.NodeID)
		} else {
			sLog.Println("Node", x.NodeID, "has failed, no need to update successor to it")
		}
	}

	// notify successor to set self as predecessor
	chord.successor.notify(chord.localNode)

	// check if all keys should stay in the local node
	err = chord.checkKeys()
	if err != nil {
		sLog.Println("Fail to check keys:", err)
		mLog.Println("Fail to check keys:", err)
		println("Fail to check keys:", err)
		return err
	}

	err = chord.checkBackup()
	if err != nil {
		sLog.Println("Fail to check backup:", err)
		mLog.Println("Fail to check backup:", err)
		println("Fail to check backup:", err)
		return err
	}
	return nil
}

// Handle notify from predecessor
func (chord *Chord) notify(node *Node) {
	sLog.Println("Get notify")
	if chord.predecessor == nil || inRange(node.NodeID, chord.predecessor.NodeID, chord.localNode.NodeID, false, false) && node.NodeID != chord.predecessor.NodeID && node.NodeID != chord.localNode.NodeID {
		chord.predecessor = node
		sLog.Println("Update predecessor to:", chord.predecessor.NodeID)

		err := chord.allKeysBackup()
		if err != nil {
			println("Fail to backup all keys:", err)
			sLog.Println("Fail to backup all keys:", err)
			mLog.Fatalln("Fail to backup all keys:", err)
		}
	}
}

// Fix fingers, fix one finger each time, use findSuccessor to check if the finger is correct
func (chord *Chord) fixFingers() error {
	next++
	if next > m {
		next = 1
	}
	sLog.Println("Fix finger table", next)
	var err error
	chord.fingerTable[next].node, err = chord.findSuccessor(chord.fingerTable[next].start)
	if err != nil {
		sLog.Println("Error in findSuccessor:", err)
		mLog.Println("Error in findSuccessor:", err)
		return err
	}
	if next == 1 {
		chord.successor = chord.fingerTable[1].node
	}
	sLog.Println("Update finger table", next, "to", chord.fingerTable[next].node.NodeID)
	return nil
}

// Update keys while a node join before the local node
func (chord *Chord) updateKeys(pre int, node *Node) error {
	sLog.Println("Update keys")
	// Open the resource directory
	dir, err := os.Open(chordResourcePath)
	if err != nil {
		sLog.Println("Error in Open", err)
		println("Error in Open", err)
		mLog.Println("Error in Open", err)
		return err
	}

	defer dir.Close()

	// Get all files of the directory
	files, err := dir.Readdir(0)
	if err != nil {
		sLog.Println("Error in Readdir", err)
		println("Error in Readdir", err)
		mLog.Println("Error in Readdir", err)
		return err
	}

	// Tranverse all files and see if it should be transfered to the new joined predecessor
	for _, file := range files {
		fileID, err := strconv.Atoi(file.Name())
		if err != nil {
			sLog.Println("Error in Atoi", err)
			println("Error in Atoi", err)
			mLog.Println("Error in Atoi", err)
			return err
		}
		if inRange(fileID, pre, node.NodeID, false, true) {
			sLog.Println("File", file.Name(), "should be sent to node", node.NodeID)
			fileBytes, err := os.ReadFile(chordResourcePath + file.Name())
			if err != nil {
				sLog.Println("Error in ReadFile:", err)
				println("Error in ReadFile:", err)
				mLog.Println("Error in ReadFile:", err)
				return err
			}
			node.remoteStoreFile(file.Name(), fileBytes, false)
			os.Remove(chordResourcePath + file.Name())
		}
	}
	return nil
}

// Check all keys while stabilization
func (chord *Chord) checkKeys() error {
	sLog.Println("Check keys")
	if chord.predecessor == nil {
		sLog.Print("Predecessor is nil, can't check keys")
		return nil
	}

	// Open the resource directory
	dir, err := os.Open(chordResourcePath)
	if err != nil {
		sLog.Println("Error in Open", err)
		println("Error in Open", err)
		mLog.Println("Error in Open", err)
		return err
	}
	defer dir.Close()

	// Get all files of the directory
	files, err := dir.Readdir(0)
	if err != nil {
		sLog.Println("Error in Readdir", err)
		println("Error in Readdir", err)
		mLog.Println("Error in Readdir", err)
		return err
	}

	// Tranverse all files and see if it should be transfered to other nodes
	for _, file := range files {
		fileID, err := strconv.Atoi(file.Name())
		if err != nil {
			sLog.Println("Error in Atoi", err)
			println("Error in Atoi", err)
			mLog.Println("Error in Atoi", err)
			return err
		}
		if !inRange(fileID, chord.predecessor.NodeID, chord.localNode.NodeID, false, true) {
			node, err := chord.findSuccessor(fileID)
			if err != nil {
				sLog.Println("Error in findSuccessor:", err)
				mLog.Println("Error in findSuccessor:", err)
				return err
			}
			sLog.Println("File", file.Name(), "should be sent to node", node.NodeID)
			if node.NodeID != chord.localNode.NodeID {
				fileBytes, err := os.ReadFile(chordResourcePath + file.Name())
				if err != nil {
					sLog.Println("Error in ReadFile:", err)
					println("Error in ReadFile:", err)
					mLog.Println("Error in ReadFile:", err)
					return err
				}
				node.remoteStoreFile(file.Name(), fileBytes, false)
				os.Remove(chordResourcePath + file.Name())
			}
		}
	}
	return nil
}

// ### Romote functions ###

// Notify successor to set self as predecessor
func (node *Node) notify(anotherNode *Node) {
	if node.NodeID == chord.localNode.NodeID {
		chord.notify(anotherNode)
		return
	}
	sLog.Println("Remote notify")
	args := Args{0, 0, anotherNode, false, ""}
	node.rpcCall("Chord.Notify", args)
}

// Let successor to update keys and send corresponding keys to self
func (node *Node) updateKeys(pre int, localNode *Node) {
	sLog.Println("Remote update keys:", node.NodeID)
	client, err := rpc.DialHTTP("tcp", node.Ip+":"+strconv.Itoa(node.Port))
	if err != nil {
		sLog.Println("dialing:", err)
		println("dialing:", err)
		mLog.Println("dialing:", err)
		return
	}
	defer client.Close()

	args := Args{pre, 0, localNode, false, ""}
	var reply *bool
	err = client.Call("Chord.UpdateKeys", args, &reply)
	if err != nil {
		println("chord error:", err)
		mLog.Println("chord error:", err)
		sLog.Fatalln("chord error:", err)
	}
}
