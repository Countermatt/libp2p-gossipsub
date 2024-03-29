package main

import (
	"context"
	"crypto/rand"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p-core/crypto"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/p2p/discovery/mdns"
)

// DiscoveryInterval is how often we re-publish our mDNS records.
const DiscoveryInterval = time.Hour

// DiscoveryServiceTag is used in our mDNS advertisements to discover other chat peers.
const DiscoveryServiceTag = "PANDAS-gossipsub-mDNS"
const sizeBlock = 512
const sizeHeader = 508
const Blocktime = 5

const colRow = 0 // 0 for column and 1 for Row parcels

type Config struct {
	Size          int
	Debug         bool
	Duration      int
	BootstrapPeer string
	NbNodes       int
	LogDir        string
}

func main() {

	//========== Experiment arguments ==========
	config := Config{}
	//nickFlag := flag.String("nick", "", "nickname for node")
	nodeType := flag.String("nodeType", "builder", "type of node: builder, regular, validator")
	flag.IntVar(&config.Size, "size", 512, "parcel size")
	flag.BoolVar(&config.Debug, "debug", false, "debug mode")
	flag.IntVar(&config.Duration, "duration", 30, "Experiment duration (in seconds).")
	//bootstrap := flag.String("bootstrap", "", "multiaddress in string form /ip4/0.0.0.0/tcp/port")
	flag.IntVar(&config.NbNodes, "nodes", 1, "temp for g5k")
	flag.StringVar(&config.LogDir, "logOutput", "/home/mapigaglio/log/", "Folder for log output")

	flag.Parse()
	ctx := context.Background()
	nodeRole := *nodeType
	//bootstrapIp := *bootstrap
	//bootstrapIp = "/ip4/0.0.0.0/tcp/0"
	size := config.Size
	log.Printf("Size:%d", size)

	// Check if the logOutput ends with a /
	if config.LogDir[len(config.LogDir)-1:] != "/" {
		config.LogDir = config.LogDir + "/"
	}

	//========== Initialise pubsub service ==========

	// create a new libp2p Host that listens on a random TCP port
	r := rand.Reader
	prvKey, _, err := crypto.GenerateKeyPairWithReader(crypto.RSA, 2048, r)
	if err != nil {
		panic(err)
	}
	h, err := libp2p.New(libp2p.ListenAddrStrings("/ip4/0.0.0.0/tcp/0"), libp2p.Identity(prvKey))
	if err != nil {
		panic(err)
	}

	// setup mDNS discovery
	if err := setupDiscovery(h); err != nil {
		panic(err)
	} // create a PubSub service using the GossipSub router
	ps, err := pubsub.NewGossipSub(ctx, h)
	if err != nil {
		panic(err)
	}
	// Generate a random nickname for node
	nick := h.ID().String()
	// join the room from the cli flag, or the flag default

	if config.Debug {
		log.Printf("Running libp2p-das-gossipsub with the following config:\n")
		log.Printf("\tNickName: %s\n", nick)
		log.Printf("\tNode Type: %s\n", nodeRole)
	}

	// join the chat room
	cr, err := CreateHost(ctx, ps, h.ID(), nick, nodeRole, sizeBlock)
	if err != nil {
		panic(err)
	}

	//========== Initialise Logger ==========
	//Create Log file
	file, err := os.OpenFile(config.LogDir+nodeRole+"-"+h.ID().String()+".log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatal("Error opening log file:", err)
	}
	defer file.Close()
	logger := log.New(file, "", 0)
	logger.SetFlags(0)
	logger.SetOutput(file)

	//Start time for load metrics
	if nodeRole == "builder" {
		handleEventsBuilder(cr, file, false, config.Size, sizeBlock, config.Duration, logger, config.NbNodes)
	} else if nodeRole == "validator" {
		handleEventsValidator(cr, file, false, nodeRole, config.Size, sizeBlock, colRow, logger, config.Duration, config.NbNodes)
	} else {
		handleEventsNonValidator(cr, file, false, nodeRole, config.Size, sizeBlock, colRow, logger, config.Duration, config.NbNodes)
	}
	//cr.messageMetrics.WriteMessageGlobalCSV()
	log.Printf("Timer expired, shutting down...\n")
}

func topicName(roomName string) string {
	return "chat-room:" + roomName
}

func setupDiscovery(h host.Host) error {
	// setup mDNS discovery to find local peers
	s := mdns.NewMdnsService(h, DiscoveryServiceTag, &discoveryNotifee{h: h})
	return s.Start()
}

// discoveryNotifee gets notified when we find a new peer via mDNS discovery
type discoveryNotifee struct {
	h host.Host
}

func (n *discoveryNotifee) HandlePeerFound(pi peer.AddrInfo) {
	fmt.Printf("discovered new peer %s\n", pi.ID.String())

	try := 0
	var err error
	for {
		err := n.h.Connect(context.Background(), pi)
		if err == nil {
			break
		} else {
			try += 1
		}
		if try == 5 {
			break
		}
	}
	if err != nil {
		fmt.Println("AAAAA")
		fmt.Printf("error connecting to peer %s: %s\n", pi.ID.String(), err)
	}
}
