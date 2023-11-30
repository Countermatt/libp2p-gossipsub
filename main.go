package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/libp2p/go-libp2p"
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

const colRow = 0 // 0 for column and 1 for Row parcels
func main() {

	//========== Experiment arguments ==========
	var duration int
	var size int
	var debug bool
	nickFlag := flag.String("nick", "", "nickname for node")
	nodeType := flag.String("nodeType", "builder", "type of node: builder, nonvalidator, builder, validator")
	flag.IntVar(&size, "size", 512, "parcel size")
	flag.BoolVar(&debug, "debug", false, "debug mode")
	flag.IntVar(&duration, "duration", 10, "Experiment duration (in seconds).")

	flag.Parse()
	ctx := context.Background()
	nodeRole := *nodeType
	log.Printf("Size:", size)
	if debug {
		log.Printf("Running libp2p-das-gossipsub with the following config:\n")
		log.Printf("\tNickName: %s\n", nickFlag)
		log.Printf("\tNode Type: %s\n", nodeRole)
	}

	//========== Initialise pubsub service ==========

	//create libp2p tracer

	// create a new libp2p Host that listens on a random TCP port
	h, err := libp2p.New(libp2p.ListenAddrStrings("/ip4/0.0.0.0/tcp/0"))
	if err != nil {
		panic(err)
	}

	//create libp2p tracer
	logfile := nodeRole + "-" + defaultNick(h.ID()) + "-Log.json"
	tracer, err := pubsub.NewJSONTracer(logfile)
	if err != nil {
		panic(err)
	}

	// create a PubSub service using the GossipSub router
	ps, err := pubsub.NewGossipSub(ctx, h, pubsub.WithEventTracer(tracer))
	if err != nil {
		panic(err)
	}

	// setup mDNS discovery
	if err := setupDiscovery(h); err != nil {
		panic(err)
	}

	// Generate a random nickname for node
	nick := defaultNick(h.ID())
	// join the room from the cli flag, or the flag default

	// join the chat room
	cr, err := CreateHost(ctx, ps, h.ID(), nick, nodeRole, sizeBlock)
	if err != nil {
		panic(err)
	}

	//Create CSV file for logging
	file, err := os.Create(cr.messageMetrics.messaeLogFile)
	if err != nil {
		log.Fatal("Error creating file:", err)
	}
	defer file.Close()

	//Start time for load metrics
	if nodeRole == "builder" {
		handleEventsBuilder(cr, file, debug, size, sizeBlock, duration)
	} else {
		handleEventsValidator(cr, file, debug, nodeRole, size, sizeBlock, colRow)
	}
	cr.messageMetrics.WriteMessageGlobalCSV()
	log.Printf("Timer expired, shutting down...\n")

}

func topicName(roomName string) string {
	return "chat-room:" + roomName
}

func defaultNick(p peer.ID) string {
	return fmt.Sprintf("%s-%s", os.Getenv("USER"), shortID(p))
}

func setupDiscovery(h host.Host) error {
	// setup mDNS discovery to find local peers
	s := mdns.NewMdnsService(h, DiscoveryServiceTag, &discoveryNotifee{h: h})
	return s.Start()
}

func shortID(p peer.ID) string {
	pretty := p.Pretty()
	return pretty[len(pretty)-8:]
}

// discoveryNotifee gets notified when we find a new peer via mDNS discovery
type discoveryNotifee struct {
	h host.Host
}

func (n *discoveryNotifee) HandlePeerFound(pi peer.AddrInfo) {
	fmt.Printf("discovered new peer %s\n", pi.ID.Pretty())
	err := n.h.Connect(context.Background(), pi)
	if err != nil {
		fmt.Printf("error connecting to peer %s: %s\n", pi.ID.Pretty(), err)
	}
}
