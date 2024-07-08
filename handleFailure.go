package main

import (
	"io/ioutil"
	"net/rpc"
	"os"
	"strconv"
)

// Check if predecessor is alive
func (chord *Chord) checkPredecessor() {
	sLog.Println()
	sLog.Println("Check predecessor")
	if chord.predecessor == nil {
		return
	}
	client, err := rpc.DialHTTP("tcp", chord.predecessor.Ip+":"+strconv.Itoa(chord.predecessor.Port))
	// if it is not alive, set predecessor as nil
	if err != nil {
		sLog.Println("Predecessor", chord.predecessor.NodeID, "failed")
		chord.predecessor = nil
	} else {
		client.Close()
	}
}

// Check if successor is alive
func (chord *Chord) checkSuccessor() error {
	sLog.Println()
	sLog.Println("Check successor")
	client, err := rpc.DialHTTP("tcp", chord.successor.Ip+":"+strconv.Itoa(chord.successor.Port))
	// if it is not alive, get the first alive node in successor list and set it as successor
	if err != nil {
		sLog.Println("Successor", chord.successor.NodeID, "failed")
		successor, err := chord.findFirstAliveSuccessor()
		if err != nil {
			cLog.Println("Error in findFirstAliveSuccessor", err)
			mLog.Println("Error in findFirstAliveSuccessor", err)
			return err
		}
		chord.successor = successor
		chord.fingerTable[1].node = successor
		sLog.Println("Set successor to", chord.successor.NodeID)

		err = chord.recoverBackupKeys()
		if err != nil {
			sLog.Print("Fail to recover all backup keys:", err)
		}
	} else {
		client.Close()
	}
	return nil
}

// Get the first alive node in successor list
func (chord *Chord) findFirstAliveSuccessor() (*Node, error) {
	sLog.Println()
	sLog.Println("Find first alive successor")
	var err error
	for i := 0; i < m; i++ {
		node := successorList[i]
		var client *rpc.Client
		client, err = rpc.DialHTTP("tcp", node.Ip+":"+strconv.Itoa(node.Port))
		if err == nil {
			client.Close()
			return node, nil
		}
	}
	return nil, err
}

// Maintain successor list
func (chord *Chord) maintainSuccessorList() error {
	sLog.Println()
	sLog.Println("Maintain successor list")
	node := chord.successor
	var err error
	// From successor, getSuccessor iteratively and set it to list
	for i := 0; i < m; i++ {
		node, err = node.getSuccessor()
		if err != nil {
			cLog.Println("Error in getSuccessor", err)
			mLog.Println("Error in getSuccessor", err)
			return err
		}
		successorList[i] = node
	}
	return nil
}

// Check if a node is alive
func (chord *Chord) checkNode(node *Node) bool {
	client, err := rpc.DialHTTP("tcp", node.Ip+":"+strconv.Itoa(node.Port))
	if err != nil {
		return false
	} else {
		client.Close()
		return true
	}
}

// Check all backup keys while stabilization
func (chord *Chord) checkBackup() error {
	mLog.Println("Check backup")
	// Open the resource directory
	dir, err := os.Open(chordBackupPath)
	if err != nil {
		cLog.Println("Error in Open", err)
		mLog.Println("Error in Open", err)
		return err
	}

	defer dir.Close()

	// Get all files of the directory
	files, err := dir.Readdir(0)
	if err != nil {
		cLog.Println("Error in Readdir", err)
		mLog.Println("Error in Readdir", err)
		return err
	}

	// Tranverse all files and see if it should be transfered to other nodes
	for _, file := range files {
		fileID, err := strconv.Atoi(file.Name())
		if err != nil {
			cLog.Println("Error in Atoi", err)
			mLog.Println("Error in Atoi", err)
			return err
		}
		if !inRange(fileID, chord.localNode.NodeID, chord.successor.NodeID, false, true) {
			os.Remove(chordBackupPath + file.Name())
		}
	}
	return nil
}

func (chord *Chord) allKeysBackup() error {
	mLog.Println("Backup all keys")
	// Open the resource directory
	dir, err := os.Open(chordResourcePath)
	if err != nil {
		cLog.Println("Error in Open", err)
		mLog.Println("Error in Open", err)
		return err
	}

	defer dir.Close()

	// Get all files of the directory
	files, err := dir.Readdir(0)
	if err != nil {
		cLog.Println("Error in Readdir", err)
		mLog.Println("Error in Readdir", err)
		return err
	}

	// Tranverse all files and see if it should be transfered to other nodes
	for _, file := range files {
		fileBytes, err := ioutil.ReadFile(chordResourcePath + file.Name())
		if err != nil {
			cLog.Println("Error in ReadFile:", err)
			mLog.Println("Error in ReadFile:", err)
			return err
		}
		chord.predecessor.remoteStoreFile(file.Name(), fileBytes, true)
	}
	return nil
}

func (chord *Chord) recoverBackupKeys() error {
	mLog.Println("Recover backup keys")
	// Open the resource directory
	dir, err := os.Open(chordBackupPath)
	if err != nil {
		cLog.Println("Error in Open", err)
		mLog.Println("Error in Open", err)
		return err
	}

	defer dir.Close()

	// Get all files of the directory
	files, err := dir.Readdir(0)
	if err != nil {
		cLog.Println("Error in Readdir", err)
		mLog.Println("Error in Readdir", err)
		return err
	}

	// Tranverse all files and see if it should be transfered to other nodes
	for _, file := range files {
		fileBytes, err := ioutil.ReadFile(chordBackupPath + file.Name())
		if err != nil {
			cLog.Println("Error in ReadFile:", err)
			mLog.Println("Error in ReadFile:", err)
			return err
		}
		chord.successor.remoteStoreFile(file.Name(), fileBytes, false)
	}
	return nil
}
