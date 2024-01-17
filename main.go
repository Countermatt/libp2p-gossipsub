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
	ma "github.com/multiformats/go-multiaddr"
)

// DiscoveryInterval is how often we re-publish our mDNS records.
const DiscoveryInterval = time.Hour

// DiscoveryServiceTag is used in our mDNS advertisements to discover other chat peers.
const DiscoveryServiceTag = "PANDAS-gossipsub-mDNS"
const sizeBlock = 512

const colRow = 0 // 0 for column and 1 for Row parcels

type Config struct {
	Size          int
	Debug         bool
	Duration      int
	BootstrapPeer string
}

func main() {

	//========== Experiment arguments ==========
	config := Config{}
	nickFlag := flag.String("nick", "", "nickname for node")
	nodeType := flag.String("nodeType", "builder", "type of node: builder, nonvalidator, builder, validator")
	flag.IntVar(&config.Size, "size", 512, "parcel size")
	flag.BoolVar(&config.Debug, "debug", false, "debug mode")
	flag.IntVar(&config.Duration, "duration", 10, "Experiment duration (in seconds).")
	flag.StringVar(&config.BootstrapPeer, "bootstrap", "", "multiaddress in string form /ip4/0.0.0.0/tcp/port")

	flag.Parse()
	ctx := context.Background()
	nodeRole := *nodeType
	log.Printf("Size:", config.Size)
	if config.Debug {
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
	/*
		logfile := nodeRole + "-" + defaultNick(h.ID()) + "-Log.pb"

		if err := touchFile(logfile); err != nil {
			fmt.Printf("Error: %v\n", err)
		}

		tracer, err := pubsub.NewPBTracer(logfile)
		if err != nil {
			panic(err)
		}
	*/
	pi, err := peer.AddrInfoFromP2pAddr(ma.StringCast("/ip4/127.0.0.1/tcp/4001/p2p/QmaWz1FJ8VQapx6Q8CtjDw2GGzzKE1nFbL3FZSQrRVBTCA"))
	if err != nil {
		panic(err)
	}
	var pi2 peer.AddrInfo
	pi2 = *pi
	tracer, err := pubsub.NewRemoteTracer(ctx, h, pi2)
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

	//========== Initialise Logger ==========
	//Create Log file
	file, err := os.OpenFile("./"+nick+".log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatal("Error opening log file:", err)
	}
	defer file.Close()
	logger := log.New(file, "", 0)
	logger.SetFlags(0)
	logger.SetOutput(file)

	//Start time for load metrics
	if nodeRole == "builder" {
		handleEventsBuilder(cr, file, config.Debug, config.Size, sizeBlock, config.Duration, logger)
	} else {
		handleEventsValidator(cr, file, config.Debug, nodeRole, config.Size, sizeBlock, colRow, logger)
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
	pretty := p.ShortString()
	return pretty[len(pretty)-8:]
}

// discoveryNotifee gets notified when we find a new peer via mDNS discovery
type discoveryNotifee struct {
	h host.Host
}

func (n *discoveryNotifee) HandlePeerFound(pi peer.AddrInfo) {
	fmt.Printf("discovered new peer %s\n", pi.ID.ShortString())
	err := n.h.Connect(context.Background(), pi)
	if err != nil {
		fmt.Printf("error connecting to peer %s: %s\n", pi.ID.ShortString(), err)
	}
}

func touchFile(filename string) error {
	file, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	// Update the modification time of the file to the current time
	currentTime := time.Now()
	if err := os.Chtimes(filename, currentTime, currentTime); err != nil {
		return err
	}

	fmt.Printf("Touched file: %s\n", filename)
	return nil
}
