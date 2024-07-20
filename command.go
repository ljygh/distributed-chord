package main

import (
	"bytes"
	"crypto/tls"
	"log"
	"net/http"
	"net/rpc"
	"os"
	"strconv"
)

// ### Command functions ###

// Find a file using filename
func (chord *Chord) lookup(filename string) error {
	// find the node which should store this file
	fileID, err := strconv.Atoi(filename)
	if err != nil {
		mLog.Println("Filename is not an integer:", err)
		println("Filename is not an integer:", err)
		return err
	}
	node, err := chord.findSuccessor(fileID)
	if err != nil {
		mLog.Println("Error in findSuccessor:", err)
		println("Error in findSuccessor:", err)
		return err
	}
	println("The possible node which stores this file is: ", node.NodeID, node.Ip, node.Port)

	// Ask the node if the file exists
	res := node.isLocalFileExist(filename)
	if res {
		println("File exists in the node:", node.NodeID, node.Ip, node.Port)
	} else {
		println("The file doesn't exist")
	}
	return nil
}

// Store a file
func (chord *Chord) storeFile(filename string) error {
	// find the node which should store this file
	println("Storing file:", filename)
	fileID, err := strconv.Atoi(filename)
	if err != nil {
		println("Filename is not an integer:", err)
		mLog.Println("Filename is not an integer:", err)
		return err
	}
	node, err := chord.findSuccessor(fileID)
	if err != nil {
		mLog.Println("Error in findSuccessor:", err)
		println("Error in findSuccessor:", err)
		return err
	}
	println("The file should be uploaded to node:", node.NodeID, node.Ip, node.Port)

	// create the file and send it to the node
	fileBytes := []byte("This is file " + filename)
	node.remoteStoreFile(filename, fileBytes, false)
	nodePredecessor, err := node.getPredecessor()
	if err == nil && nodePredecessor != nil && node.NodeID != nodePredecessor.NodeID {
		nodePredecessor.remoteStoreFile(filename, fileBytes, true)
	}
	return nil
}

// Print all info of this chord node
func (chord *Chord) printState() {
	println()
	println("Print Chord")
	println("Local node id:", chord.localNode.NodeID, "IP:", chord.localNode.Ip, "port:", chord.localNode.Port)
	if chord.predecessor != nil {
		println("Predecessor node id:", chord.predecessor.NodeID, "IP:", chord.predecessor.Ip, "port:", chord.predecessor.Port)
	} else {
		println("Predecessor: nil")
	}
	println("Sucessor node id:", chord.successor.NodeID, "IP:", chord.successor.Ip, "port:", chord.successor.Port)
	println("Print finger table")
	for i := 1; i <= m; i++ {
		entry := chord.fingerTable[i]
		println("start:", entry.start, "node id:", entry.node.NodeID, "IP:", entry.node.Ip, "port:", entry.node.Port)
	}
	println("Successor list:")
	for i := 0; i < m; i++ {
		print(successorList[i].NodeID, " ")
	}
	println()
	println()
}

// ### Local helper functions ###

// Tell if a file exists locally
func (chord *Chord) isLocalFileExist(filename string) bool {
	if _, err := os.Stat(chordResourcePath + filename); err == nil {
		mLog.Println("File", filename, "exists")
		return true
	} else {
		mLog.Println("File", filename, "does not exist")
		return false

	}

}

// Save a file to node resource folder
func (chord *Chord) localStoreFile(filename string, fileBytes []byte, isBackup bool) {
	var filepath string
	if isBackup {
		mLog.Println("Local backup file:", filename)
		filepath = chordBackupPath + filename
	} else {
		mLog.Println("Local store file:", filename)
		filepath = chordResourcePath + filename
	}

	file, err := os.Create(filepath)
	if err != nil {
		mLog.Println(err)
		log.Fatalln(err)
	}

	_, err = file.Write(fileBytes)
	if err != nil {
		mLog.Println(err)
		log.Fatalln(err)
	}

	err = file.Close()
	if err != nil {
		mLog.Println(err)
		log.Fatalln(err)
	}
}

// ### Remote helper functions ###

// Ask node if it stores the file
func (node *Node) isLocalFileExist(filename string) bool {
	if node.NodeID == chord.localNode.NodeID {
		return chord.isLocalFileExist(filename)
	}
	// dial rpc
	client, err := rpc.DialHTTP("tcp", node.Ip+":"+strconv.Itoa(node.Port))
	if err != nil {
		println("dialing:", err)
		mLog.Fatalln("dialing:", err)
	}
	defer client.Close()

	// call remote function and get reply
	args := Args{0, 0, nil, false, filename}
	var reply *bool
	err = client.Call("Chord.IsLocalFileExist", args, &reply)
	if err != nil {
		println("chord error:", err)
		mLog.Fatalln("chord error:", err)
	}
	return *reply
}

// Let node store the file remotely
func (node *Node) remoteStoreFile(filename string, fileBytes []byte, isBackup bool) {
	if node.NodeID == chord.localNode.NodeID && isBackup {
		chord.localStoreFile(filename, fileBytes, true)
		return
	}
	if node.NodeID == chord.localNode.NodeID && !isBackup {
		chord.localStoreFile(filename, fileBytes, false)
		return
	}

	// set http connection
	if isBackup {
		mLog.Println("Remote backup to", node.NodeID, node.Ip, ":", node.Port+1)
	} else {
		mLog.Println("Remote upload to", node.NodeID, node.Ip, ":", node.Port+1)
	}

	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
	}
	defer client.CloseIdleConnections()

	// send the file
	var url string
	if isBackup {
		url = "http://" + node.Ip + ":" + strconv.Itoa(node.Port+1) + "/resource" + "/chord" + strconv.Itoa(node.NodeID) + "_" + "backup/" + filename
	} else {
		url = "http://" + node.Ip + ":" + strconv.Itoa(node.Port+1) + "/resource" + "/chord" + strconv.Itoa(node.NodeID) + "/" + filename
	}
	mLog.Println("Post to:", url)
	resp, err := client.Post(url, "text/plain", bytes.NewReader(fileBytes))
	if err != nil {
		println(err)
		mLog.Fatalln(err)
	}

	resp.Body.Close()
}
