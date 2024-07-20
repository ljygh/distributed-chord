package main

import (
	"bufio"
	"io"
	"log"
	"net"
	"net/http"
	"net/rpc"
	"os"
	"strconv"
	"strings"
	"time"
)

func main() {
	var setting Setting

	// Read settings
	if len(os.Args) < 2 {
		println("Please input a path of setting file")
		return
	} else {
		settingPath := os.Args[1]
		getSetting(settingPath, &setting)
	}

	// Init chord local node
	nodeID := setting.ID
	IP := setting.IP
	port := setting.Port
	chord.localNode = &Node{nodeID, IP, port}

	// Init ts tff tcp and m
	ts = setting.Ts
	tff = setting.Tff
	tcp = setting.Tcp

	// Init log path and resource path
	logPath := setting.LogPath
	resourcePath := setting.ResourcePath

	// While creating first node, remove log and resource folders.
	if setting.IfCreate {
		remove_dirs(logPath)
		remove_dirs(resourcePath)
	}

	// Create resource and log folders if they don't exist.
	if _, err := os.Stat(logPath); os.IsNotExist(err) {
		err := os.Mkdir(logPath, 0777)
		if err != nil {
			mLog.Println("Error while creating 'log' folder:", err)
			log.Fatalln("Error while creating 'log' folder:", err)
		}
	}

	if _, err := os.Stat(resourcePath); os.IsNotExist(err) {
		err := os.Mkdir(resourcePath, 0777)
		if err != nil {
			mLog.Println("Error while creating 'resource' folder:", err)
			log.Fatalln("Error while creating 'resource' folder:", err)
		}
	}

	// Set log
	if setting.IfMlog {
		mainLogFile, err := os.Create(logPath + "/chord_" + strconv.Itoa(chord.localNode.NodeID) + "_main.log")
		if err != nil {
			mLog.Println(err)
			log.Fatalln(err)
		}
		mLog = *log.New(mainLogFile, "", log.Lshortfile)
		println("Set mLog file:", logPath+"/chord_"+strconv.Itoa(chord.localNode.NodeID)+"_main.log")
		defer mainLogFile.Close()
	} else {
		mLog = *log.New(io.Discard, "", log.Lshortfile)
	}

	if setting.IfSlog {
		stableLogFile, err := os.Create(logPath + "/chord_" + strconv.Itoa(chord.localNode.NodeID) + "_stabilize.log")
		if err != nil {
			mLog.Fatalln(err)
			log.Fatalln(err)
		}
		sLog = *log.New(stableLogFile, "", log.Lshortfile)
		println("Set sLog file:", logPath+"/chord_"+strconv.Itoa(chord.localNode.NodeID)+"_stabilize.log")
		defer stableLogFile.Close()
	} else {
		sLog = *log.New(io.Discard, "", log.Lshortfile)
	}

	// Create chord resource folder and backup folder
	chordResourcePath = resourcePath + "/chord" + strconv.Itoa(chord.localNode.NodeID) + "/"
	if _, err := os.Stat(chordResourcePath); err != nil {
		err = os.Mkdir(chordResourcePath, 0777)
		if err != nil {
			println(err)
			mLog.Fatalln(err)
		}
	} else {
		err = os.RemoveAll(chordResourcePath)
		if err != nil {
			println(err)
			mLog.Fatalln(err)
		}
		err = os.Mkdir(chordResourcePath, 0777)
		if err != nil {
			println(err)
			mLog.Fatalln(err)
		}
	}
	println("Create resource folder:", chordResourcePath)

	chordBackupPath = resourcePath + "/chord" + strconv.Itoa(chord.localNode.NodeID) + "_backup/"
	if _, err := os.Stat(chordBackupPath); err != nil {
		err = os.Mkdir(chordBackupPath, 0777)
		if err != nil {
			println(err)
			mLog.Fatalln(err)
		}
	} else {
		err = os.RemoveAll(chordBackupPath)
		if err != nil {
			println(err)
			mLog.Fatalln(err)
		}
		err = os.Mkdir(chordBackupPath, 0777)
		if err != nil {
			println(err)
			mLog.Fatalln(err)
		}
	}
	println("Create backup folder:", chordBackupPath)

	// Register rpc and serve
	rpc.Register(&chord)
	rpc.HandleHTTP()
	l, e := net.Listen("tcp", ":"+strconv.Itoa(chord.localNode.Port))
	if e != nil {
		mLog.Println("listen error:", e)
		log.Fatal("listen error:", e)
	}
	go http.Serve(l, nil)

	// Set http server
	http.HandleFunc("/", httpHandler)
	go http.ListenAndServe(":"+strconv.Itoa(chord.localNode.Port+1), nil)

	// Create or join according number of arguments
	if setting.IfCreate {
		chord.create()
	} else {
		joinNodeIP := setting.RemoteIP
		joinNodePort := setting.RemotePort
		joinNode := Node{0, joinNodeIP, joinNodePort}
		chord.join(joinNode)
	}

	// Run stabilizing functions periodically
	go stabilize_periodic()
	go fixFingers_periodic()
	go checkPredecessor_periodic()

	// handle inputed commands
	in := bufio.NewReader(os.Stdin)
	for {
		var input string
		print("Command >>")
		input, _ = in.ReadString('\n')
		if len(input) > 0 {
			input = input[:len(input)-1]
		}

		if strings.HasPrefix(input, "print") {
			chord.printState()
		} else if strings.HasPrefix(input, "store") {
			filename := strings.SplitAfter(input, " ")[1]
			err := chord.storeFile(filename)
			if err != nil {
				println("Fail to store file:", err)
				mLog.Fatalln("Fail to store file:", err)
			}
		} else if strings.HasPrefix(input, "find") {
			filename := strings.SplitAfter(input, " ")[1]
			err := chord.lookup(filename)
			if err != nil {
				println("Fail to find file:", err)
				mLog.Fatalln("Fail to find file:", err)
			}
		} else {
			println("Please write the correct syntax")
		}
	}
}

func stabilize_periodic() {
	for {
		chord.stabilize()
		time.Sleep(time.Duration(ts) * time.Millisecond)
	}
}

func fixFingers_periodic() {
	for {
		chord.fixFingers()
		time.Sleep(time.Duration(tff) * time.Millisecond)
	}
}

func checkPredecessor_periodic() {
	for {
		chord.checkPredecessor()
		time.Sleep(time.Duration(tcp) * time.Millisecond)
	}
}

// Http handler
func httpHandler(w http.ResponseWriter, req *http.Request) {
	if req.Method == "POST" { // store file after getting post
		mLog.Println("Get post")

		filePath := "." + req.RequestURI
		mLog.Println("Post file to:", filePath)

		file, err := os.Create(filePath)
		if err != nil {
			println(err)
			mLog.Fatalln(err)
		}

		body, err := io.ReadAll(req.Body)
		if err != nil {
			println(err)
			mLog.Fatalln(err)
			return
		}
		_, err = file.Write(body)
		if err != nil {
			println(err)
			mLog.Fatalln(err)
			return
		}
	}
}
