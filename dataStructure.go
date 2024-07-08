package main

import "log"

type Node struct {
	NodeID int
	Ip     string
	Port   int
}

type Entry struct {
	start int
	node  *Node
}

type Chord struct {
	localNode   *Node
	successor   *Node
	predecessor *Node
	fingerTable [m + 1]*Entry
}

// Args for rpc
type Args struct {
	Arg1 int
	Arg2 int
	Arg3 *Node
	Arg4 bool
	Arg5 string
}

// m for Chord ring
const m int = 6

// Used for fixFingers
var next int = 0

// Stabilization time period
var ts int = 500
var tff int = 500
var tcp int = 500

var cLog log.Logger // command log
var mLog log.Logger // main thread log
var sLog log.Logger // stabilizing thread log

var chord = Chord{}

// Path to store files
var chordResourcePath string
var chordBackupPath string

var successorList [m]*Node

// key and nonce to encrypt files
var key []byte
var nonce []byte
