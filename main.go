package main

import (
	"TP2_Minimum_Spanning_Tree/environment"
	"TP2_Minimum_Spanning_Tree/logger"
	"TP2_Minimum_Spanning_Tree/message"
	"TP2_Minimum_Spanning_Tree/neighbour"
	"TP2_Minimum_Spanning_Tree/nodeState"
	"bufio"
	"fmt"
	"math"
	"net"
	"os"
	"strconv"
	"strings"
	"sync"
)

var (
	network          = "tcp"
	address          string
	port             string
	state            nodeState.NodeState
	fragmentId       string
	level                                 = 0
	neighbours                            = make([]*neighbour.Neighbour, 0)
	neighboursMap                         = make(map[string]*neighbour.Neighbour, 0)
	findCount                             = 0
	queuedMessages                        = make([]message.Message, 0)
	queuedNeighbours                      = make([]*neighbour.Neighbour, 0)
	bestEdge         *neighbour.Neighbour = nil
	bestWeight                            = math.MaxInt32
	inBranch         *neighbour.Neighbour = nil
	testEdge         *neighbour.Neighbour = nil
	finished                              = false
	env                                   = environment.Cloud
	mutexQueue       sync.Mutex
	mutexFinished    sync.Mutex
)

func getAddressPort() string {
	return fmt.Sprintf("%s:%s", address, port)
}

func main() {
	if len(os.Args) < 5 ||
		(environment.Environment(os.Args[4]) != environment.Cloud &&
			environment.Environment(os.Args[4]) != environment.Local) {
		fmt.Printf("Usage: %s address port nodeId env", os.Args[0])
		os.Exit(1)
	}
	address = os.Args[1]
	port = os.Args[2]
	state = nodeState.Sleeping
	fragmentId = os.Args[3]
	env = environment.Environment(os.Args[4])

	neighbours = neighbour.LoadFromFile("./neighbours-" + fragmentId + ".txt")
	if len(neighbours) == 0 {
		logger.Log(main, "Aucun voisin.")
		os.Exit(1)
	}
	neighbour.SortNeighbours(neighbours)
	logger.Log(main, "Affichage des voisins lus dans le fichier...")
	neighboursMap = neighbour.NeighboursToMap(neighbours, env)
	for _, n := range neighbours {
		fmt.Println(n.ToString())
	}
	go launchServer()
	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Println("=== Algorithme GHS ===")
		fmt.Println("1. Réveiller le nœud")
		fmt.Println("2. Quitter et attendre la fin de l'algorithme")
		fmt.Print("> ")
		choice, _ := reader.ReadString('\n')
		choice = strings.Trim(choice, "\n")
		if choice == "1" {
			if state == nodeState.Sleeping {
				wakeUp()
			}
		} else if choice == "2" {
			break
		}
	}
	for {
		mutexFinished.Lock()
		if finished {
			mutexFinished.Unlock()
			break
		}
		mutexFinished.Unlock()
	}
	logger.Log(main, "L'algorithme est fini, affichage de l'état des voisins...")
	for _, n := range neighbours {
		fmt.Println(n.ToString())
	}
}

func wakeUp() {
	logger.Log(wakeUp, "Je me réveille...")
	neighbours[0].Type = neighbour.Branch
	level = 0
	state = nodeState.Found
	findCount = 0
	sendConnectMessage(neighbours[0])
}

func sendMessage(neigh *neighbour.Neighbour, msg message.Message) {
	conn, err := net.Dial(network, neigh.ToAddressPort())
	if err != nil {
		logger.Log(sendConnectMessage, fmt.Sprintf("Impossible de se connecter au voisin : '%s'.", err.Error()))
		return
	}
	jsonMessage, err := msg.ToJSON()
	if err != nil {
		logger.Log(sendMessage, fmt.Sprintf("Impossible de convertir %s en JSON : '%s'", msg.ToString(), err.Error()))
		return
	}
	_, err = conn.Write(jsonMessage)
	if err != nil {
		logger.Log(sendMessage, fmt.Sprintf("Impossible d'envoyer %s : '%s'", msg.ToString(), err.Error()))
		return
	}
	logger.Log(sendMessage, fmt.Sprintf("J'ai envoyé ce %s à '%s'.", msg.ToString(), conn.RemoteAddr()))
}

func sendConnectMessage(neigh *neighbour.Neighbour) {
	connectMessage := message.New(message.Connect, "", level, 0, "", port)
	sendMessage(neigh, connectMessage)
}

// https://dev.to/alicewilliamstech/getting-started-with-sockets-in-golang-2j66
func launchServer() {
	listener, err := net.Listen(network, getAddressPort())
	if err != nil {
		logger.Log(launchServer, fmt.Sprintf("Erreur lors de l'écoute : '%s'.", err.Error()))
		_ = listener.Close()
		os.Exit(1)
	}
	defer listener.Close()
	logger.Log(launchServer, fmt.Sprintf("Ecoute sur %s://%s...", network, getAddressPort()))
	go handleQueue()
	for {
		client, err := listener.Accept()
		if err == nil {
			logger.Log(launchServer, fmt.Sprintf("Client %s connecté.", client.RemoteAddr()))
			handleClient(client)
		} else {
			logger.Log(launchServer, fmt.Sprintf("Erreur de connexion : %s.", err.Error()))
		}
	}
}

func splitAddr(address net.Addr) []string {
	ipPort := make([]string, 2)
	switch addr := address.(type) {
	case *net.UDPAddr:
		ipPort[0] = addr.IP.String()
		ipPort[1] = strconv.Itoa(addr.Port)
	case *net.TCPAddr:
		ipPort[0] = addr.IP.String()
		ipPort[1] = strconv.Itoa(addr.Port)
	}
	return ipPort
}

func handleClient(conn net.Conn) {
	msg, err := message.FromJSON(conn)
	if err != nil {
		logger.Log(handleClient, fmt.Sprintf("Impossible de convertir le JSON en Message (%s).", err.Error()))
		_ = conn.Close()
		return
	}
	_ = conn.Close()
	logger.Log(handleClient, fmt.Sprintf("J'ai décodé ce %s de %s.", msg.ToString(), conn.RemoteAddr()))
	mutexQueue.Lock()
	neigh := getNeighbour(conn.RemoteAddr(), msg.Port)
	queuedMessages = append(queuedMessages, msg)
	queuedNeighbours = append(queuedNeighbours, neigh)
	mutexQueue.Unlock()
}

func handleQueue() {
	for {
		mutexFinished.Lock()
		if finished {
			mutexFinished.Unlock()
			break
		}
		mutexFinished.Unlock()
		mutexQueue.Lock()
		if len(queuedMessages) > 0 {
			msg := queuedMessages[0]
			neigh := queuedNeighbours[0]

			logger.Log(handleQueue, fmt.Sprintf("Je vais gérer %s de %s.", msg.ToString(), neigh.ToAddressPort()))

			queuedMessages = queuedMessages[1:]
			queuedNeighbours = queuedNeighbours[1:]
			mutexQueue.Unlock()

			switch msg.Type {
			case message.Connect:
				respondToConnectMessage(neigh, msg)
			case message.Initiate:
				respondToInitiateMessage(neigh, msg)
			case message.Test:
				respondToTestMessage(neigh, msg)
			case message.Accept:
				respondToAcceptMessage(neigh)
			case message.Reject:
				respondToRejectMessage(neigh)
			case message.Report:
				respondToReportMessage(neigh, msg)
			case message.ChangeRoot:
				respondToChangeRootMessage()
			}
		} else {
			mutexQueue.Unlock()
		}
	}
}

func findMinimumWeightNeighbour() *neighbour.Neighbour {
	for _, neigh := range neighbours {
		/**
		* Remember, neighbours are sort by weight (asc), so the first basic neighbour is the one
		**/
		if neigh.Type == neighbour.Basic {
			return neigh
		}
	}
	return nil
}

func getNeighbour(remoteAddr net.Addr, port string) *neighbour.Neighbour {
	ipPort := splitAddr(remoteAddr)
	if env == environment.Cloud {
		return neighboursMap[ipPort[0]]
	} else {
		return neighboursMap[port]
	}
}

func respondToConnectMessage(neigh *neighbour.Neighbour, msg message.Message) {
	fmt.Println(neigh.ToString())
	if state == nodeState.Sleeping {
		wakeUp()
	}
	if msg.Level < level {
		neigh.Type = neighbour.Branch
		initiateMessage := message.New(message.Initiate, fragmentId, level, 0, state, port)
		sendMessage(neigh, initiateMessage)
		if state == nodeState.Find {
			findCount += 1
		}
	} else if neigh.Type == neighbour.Basic {
		mutexQueue.Lock()
		queuedMessages = append(queuedMessages, msg)
		queuedNeighbours = append(queuedNeighbours, neigh)
		mutexQueue.Unlock()
		logger.Log(respondToConnectMessage, fmt.Sprintf("%s ajouté à la file d'attente.", msg.ToString()))
	} else {
		initiateMessage := message.New(message.Initiate, strconv.Itoa(neigh.Weight), level+1, 0, nodeState.Find, port)
		sendMessage(neigh, initiateMessage)
	}
}

func respondToInitiateMessage(neigh *neighbour.Neighbour, msg message.Message) {
	level = msg.Level
	fragmentId = msg.FragmentId
	state = msg.State
	bestEdge = nil
	bestWeight = math.MaxInt32
	inBranch = neigh
	for _, n := range neighbours {
		if neigh != n && n.Type == neighbour.Branch {
			initiateMessage := message.New(message.Initiate, msg.FragmentId, msg.Level, 0, msg.State, port)
			sendMessage(n, initiateMessage)
			if msg.State == nodeState.Find {
				findCount += 1
			}
		}
	}
	if msg.State == nodeState.Find {
		test()
	}
}

func test() {
	neigh := findMinimumWeightNeighbour()
	if neigh != nil {
		testEdge = neigh
		testMessage := message.New(message.Test, fragmentId, level, 0, "", port)
		sendMessage(neigh, testMessage)
	} else {
		testEdge = nil
		report()
	}
}

func respondToTestMessage(neigh *neighbour.Neighbour, msg message.Message) {
	if state == nodeState.Sleeping {
		wakeUp()
	}
	if msg.Level > level {
		mutexQueue.Lock()
		queuedMessages = append(queuedMessages, msg)
		queuedNeighbours = append(queuedNeighbours, neigh)
		mutexQueue.Unlock()
		logger.Log(respondToTestMessage, fmt.Sprintf("%s ajouté à la file d'attente car %d > %d.", msg.ToString(), msg.Level, level))
	} else if msg.FragmentId != fragmentId {
		acceptMessage := message.New(message.Accept, "", 0, 0, "", port)
		sendMessage(neigh, acceptMessage)
	} else {
		if neigh.Type == neighbour.Basic {
			neigh.Type = neighbour.Rejected
		}
		if testEdge != neigh {
			rejectMessage := message.New(message.Reject, "", 0, 0, "", port)
			sendMessage(neigh, rejectMessage)
		} else {
			test()
		}
	}
}

func respondToAcceptMessage(neigh *neighbour.Neighbour) {
	testEdge = nil
	if neigh.Weight < bestWeight {
		bestEdge = neigh
		bestWeight = neigh.Weight
	}
	report()
}

func respondToRejectMessage(neigh *neighbour.Neighbour) {
	if neigh.Type == neighbour.Basic {
		neigh.Type = neighbour.Rejected
	}
	test()
}

func report() {
	if findCount == 0 && testEdge == nil {
		state = nodeState.Found
		reportMessage := message.New(message.Report, "", 0, bestWeight, "", port)
		sendMessage(inBranch, reportMessage)
	}
}

func respondToReportMessage(neigh *neighbour.Neighbour, msg message.Message) {
	if neigh != inBranch {
		findCount -= 1
		if msg.Weight < bestWeight {
			bestWeight = msg.Weight
			bestEdge = neigh
		}
		report()
	} else if state == nodeState.Find {
		mutexQueue.Lock()
		queuedMessages = append(queuedMessages, msg)
		queuedNeighbours = append(queuedNeighbours, neigh)
		mutexQueue.Unlock()
	} else if msg.Weight > bestWeight {
		changeRoot()
	} else if msg.Weight == bestWeight && bestWeight == math.MaxInt32 {
		mutexFinished.Lock()
		finished = true
		mutexFinished.Unlock()
		logger.Log(respondToReportMessage, fmt.Sprintf("Le noeud root est : %s.", fragmentId))
		for _, n := range neighbours {
			reportMessage := message.New(message.Report, "", 0, bestWeight, "", port)
			sendMessage(n, reportMessage)
		}
	}
}

func changeRoot() {
	if bestEdge.Type == neighbour.Branch {
		changeRootMessage := message.New(message.ChangeRoot, "", 0, 0, "", port)
		sendMessage(bestEdge, changeRootMessage)
	} else {
		bestEdge.Type = neighbour.Branch
		sendConnectMessage(bestEdge)
	}
}

func respondToChangeRootMessage() {
	changeRoot()
}
