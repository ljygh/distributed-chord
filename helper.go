package main

import (
	"encoding/json"
	"io"
	"log"
	"math"
	"net/rpc"
	"os"
	"strconv"
)

// Tell if a num in a range in the ring
func inRange(num int, rangeL int, rangeR int, includeL bool, includeR bool) bool {
	if rangeR > rangeL {
		return (num > rangeL && num < rangeR) || (includeL && num == rangeL) || (includeR && num == rangeR)
	} else if rangeL > rangeR {
		return (num > rangeL || num < rangeR) || (includeL && num == rangeL) || (includeR && num == rangeR)
	} else {
		return true
	}
}

// Minus in the ring
func minusInRing(x int, y int) int {
	if x >= y {
		return x - y
	} else {
		return x + int(math.Pow(2, float64(m))) - y
	}
}

// Call to node to execute a rpc function
func (node *Node) rpcCall(function string, args Args) (*Node, error) {
	// dial
	client, err := rpc.DialHTTP("tcp", node.Ip+":"+strconv.Itoa(node.Port))
	if err != nil {
		mLog.Println("Chord", chord.localNode.NodeID, "call to", node.NodeID, node.Ip+":"+strconv.Itoa(node.Port), "for", function)
		mLog.Println("dialing:", err)
		return nil, err
	}
	defer client.Close()

	// call and get reply
	var reply *Node
	err = client.Call(function, args, &reply)
	if err != nil {
		mLog.Println("chord error:", err)
		return nil, err
	}
	return reply, nil
}

// Get setting from json file
func getSetting(filePath string, setting *Setting) {
	// Open the JSON file
	jsonFile, err := os.Open(filePath)
	println("File path:", filePath)
	if err != nil {
		mLog.Println(err)
		log.Fatalln(err)
	}
	defer jsonFile.Close()

	// Read the file content
	byteValue, err := io.ReadAll(jsonFile)
	if err != nil {
		mLog.Println(err)
		log.Fatalln(err)
	}

	// Unmarshal the JSON data into the struct
	err = json.Unmarshal(byteValue, setting)
	if err != nil {
		mLog.Println(err)
		log.Fatalln(err)
	}
}

// remove all dirs in a directory
func remove_dirs(path string) {
	err := os.RemoveAll(path)
	if err != nil {
		mLog.Printf("Failed to remove directory: %v", err)
		log.Fatalf("Failed to remove directory: %v", err)
	}
}
