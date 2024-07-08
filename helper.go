package main

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"io"
	"math"
	"net/rpc"
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
		cLog.Println("dialing:", err)
		return nil, err
	}
	defer client.Close()

	// call and get reply
	var reply *Node
	err = client.Call(function, args, &reply)
	if err != nil {
		mLog.Println("chord error:", err)
		cLog.Println("chord error:", err)
		return nil, err
	}
	return reply, nil
}

// Encrypt a file, return encrypted content
func encrypt(plainTxt []byte) []byte {
	block, err := aes.NewCipher(key)
	if err != nil {
		panic(err.Error())
	}

	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		panic(err.Error())
	}

	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		panic(err.Error())
	}

	ciphertext := aesgcm.Seal(nil, nonce, plainTxt, nil)
	return ciphertext
}
