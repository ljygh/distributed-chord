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

var mLog log.Logger // main thread log
var sLog log.Logger // stabilizing thread log

var chord = Chord{}

// Path to store files
var chordResourcePath string
var chordBackupPath string

var successorList [m]*Node

// Struct used for setting
type Setting struct {
	Ts           int    `json:"ts"`
	Tff          int    `json:"tff"`
	Tcp          int    `json:"tcp"`
	ID           int    `json:"id"`
	IP           string `json:"IP"`
	Port         int    `json:"port"`
	IfMlog       bool   `json:"ifMlog"`
	IfSlog       bool   `json:"ifSlog"`
	LogPath      string `json:"logPath"`
	ResourcePath string `json:"resourcePath"`
	IfCreate     bool   `json:"ifCreate"`
	RemoteIP     string `json:"remoteIP"`
	RemotePort   int    `json:"remotePort"`
}
