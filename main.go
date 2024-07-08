package main

import (
	"bufio"
	"crypto/rand"
	"io"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"net/rpc"
	"os"
	"path"
	"strconv"
	"strings"
	"time"
)

func main() {
	// Init chord local node
	nodeID, _ := strconv.Atoi(os.Args[1])
	IP := os.Args[2]
	port, _ := strconv.Atoi(os.Args[3])
	chord.localNode = &Node{nodeID, IP, port}

	// Init ts tff tcp and m
	ts, _ = strconv.Atoi(os.Args[4])
	tff, _ = strconv.Atoi(os.Args[5])
	tcp, _ = strconv.Atoi(os.Args[6])

	// Clear resource dirs and logs
	remove_dirs_logs()

	// Set log
	mainLogFile, err := os.Create("./log/chord_" + strconv.Itoa(chord.localNode.NodeID) + "_main.log")
	if err != nil {
		log.Fatalln(err)
	}
	defer mainLogFile.Close()

	// stableLogFile, err := os.Create("./log/chord_" + strconv.Itoa(chord.localNode.NodeID) + "_stabilize.log")
	// if err != nil {
	// 	log.Fatalln(err)
	// }
	// defer stableLogFile.Close()

	// cLog = *log.New(os.Stdout, "", log.Flags())
	cLog = *log.New(io.Discard, "", log.Flags())
	cLog.SetFlags(0)
	// mLog = *log.New(mainLogFile, "", log.Lshortfile)
	mLog = *log.New(io.Discard, "", log.Lshortfile)
	// sLog = *log.New(mainLogFile, "", log.Lshortfile)
	sLog = *log.New(io.Discard, "", log.Lshortfile)

	// Init key and nonce for encrypt
	key = make([]byte, 32)
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		println("Error while generating key:", err)
		mLog.Fatalln("Error while generating key:", err)
	}

	nonce = make([]byte, 12)
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		println("Error while generating nonce:", err)
		mLog.Fatalln("Error while generating nonce:", err)
	}

	// Create chord resource folder and backup folder
	chordResourcePath = "./resource/chord" + strconv.Itoa(chord.localNode.NodeID) + "/"
	if _, err = os.Stat(chordResourcePath); err != nil {
		err = os.Mkdir(chordResourcePath, 0777)
	} else {
		err = os.RemoveAll(chordResourcePath)
		err = os.Mkdir(chordResourcePath, 0777)
	}
	if err != nil {
		println(err)
		mLog.Fatal(err)
	}

	chordBackupPath = "./resource/chord" + strconv.Itoa(chord.localNode.NodeID) + "_backup/"
	if _, err = os.Stat(chordBackupPath); err != nil {
		err = os.Mkdir(chordBackupPath, 0777)
	} else {
		err = os.RemoveAll(chordBackupPath)
		err = os.Mkdir(chordBackupPath, 0777)
	}
	if err != nil {
		println(err)
		mLog.Fatal(err)
	}

	// Register rpc and serve
	rpc.Register(&chord)
	rpc.HandleHTTP()
	l, e := net.Listen("tcp", ":"+strconv.Itoa(chord.localNode.Port))
	if e != nil {
		mLog.Println("listen error:", e)
		log.Fatal("listen error:", e)
	}
	go http.Serve(l, nil)

	// Set https server
	httpsL, e := net.Listen("tcp", ":"+strconv.Itoa(chord.localNode.Port+1))
	if e != nil {
		mLog.Println("https listen error:", e)
		log.Fatal("https listen error:", e)
	}
	http.HandleFunc("/", httpsHandler)
	go http.ServeTLS(httpsL, nil, "example.crt", "example.key")

	// Create or join according number of arguments
	if len(os.Args) <= 7 {
		chord.create()
	} else {
		joinNodeIP := os.Args[7]
		joinNodePort, _ := strconv.Atoi(os.Args[8])
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
			}
		} else if strings.HasPrefix(input, "find") {
			filename := strings.SplitAfter(input, " ")[1]
			err := chord.lookup(filename)
			if err != nil {
				println("Fail to find file:", err)
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

// Https handler
func httpsHandler(w http.ResponseWriter, req *http.Request) {
	if req.Method == "POST" { // store file after getting post
		mLog.Println("Get post")
		urlPath := req.URL.Path
		dir := path.Base(path.Dir(urlPath))
		filename := path.Base(urlPath)
		mLog.Println("Receive file:", filename)
		mLog.Println("Save to:", dir)

		var file *os.File
		var err error
		if dir == "resource" {
			file, err = os.Create(chordResourcePath + filename)
		} else {
			file, err = os.Create(chordBackupPath + filename)
		}
		if err != nil {
			cLog.Println(err)
			mLog.Println(err)
			return
		}

		body, err := ioutil.ReadAll(req.Body)
		if err != nil {
			cLog.Println(err)
			mLog.Println(err)
			return
		}
		_, err = file.Write(body)
		if err != nil {
			cLog.Println(err)
			mLog.Println(err)
			return
		}
	}
}

// remove all dirs in resource
func remove_dirs_logs() {
	files, err := ioutil.ReadDir("./resource")
	if err != nil {
		log.Fatal("Error while reading dir: ", err)
	}

	for _, file := range files {
		filename := file.Name()
		if file.IsDir() {
			os.RemoveAll("./resource/" + filename)
		}
	}

	files, err = ioutil.ReadDir("./log")
	if err != nil {
		log.Fatal("Error while reading dir: ", err)
	}

	for _, file := range files {
		filename := file.Name()
		os.RemoveAll("./log/" + filename)
	}
}
