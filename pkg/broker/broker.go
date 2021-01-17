package broker

import (
	"encoding/json"
	galactus_client "github.com/automuteus/galactus/pkg/client"
	"github.com/automuteus/utils/pkg/capture"
	"github.com/automuteus/utils/pkg/game"
	socketio "github.com/googollee/go-socket.io"
	"github.com/gorilla/mux"
	"go.uber.org/zap"
	"log"
	"net/http"
	"strconv"
	"sync"
	"time"
)

const ConnectCodeLength = 8

type Broker struct {
	client *galactus_client.GalactusClient

	// map of socket IDs to connection codes
	connections map[string]string

	ackKillChannels map[string]chan bool
	connectionsLock sync.RWMutex
}

func NewBroker(galactusAddr string, logger *zap.Logger) *Broker {
	client, err := galactus_client.NewGalactusClient(galactusAddr, logger)
	for err != nil {
		logger.Error("error connecting to galactus. Retrying every second until it is reachable",
			zap.Error(err))
		time.Sleep(time.Second)
		client, err = galactus_client.NewGalactusClient(galactusAddr, logger)
	}
	return &Broker{
		client:          client,
		connections:     map[string]string{},
		ackKillChannels: map[string]chan bool{},
		connectionsLock: sync.RWMutex{},
	}
}

func (broker *Broker) Start(port string) {
	server, err := socketio.NewServer(nil)
	if err != nil {
		log.Fatal(err)
	}

	server.OnConnect("/", func(s socketio.Conn) error {
		s.SetContext("")
		log.Println("connected:", s.ID())
		return nil
	})
	server.OnEvent("/", "connectCode", func(s socketio.Conn, msg string) {
		log.Printf("Received connection code: \"%s\"", msg)

		if len(msg) != ConnectCodeLength {
			s.Close()
		} else {
			killChannel := make(chan bool)

			broker.connectionsLock.Lock()
			broker.connections[s.ID()] = msg
			broker.ackKillChannels[s.ID()] = killChannel
			broker.connectionsLock.Unlock()

			event := capture.Event{
				EventType: capture.Connection,
				Payload:   []byte("true"),
			}
			err := broker.client.AddCaptureEvent(msg, event)
			if err != nil {
				log.Println(err)
			}
		}
	})

	// only join the room for the connect code once we ensure that the bot actually connects with a valid discord session
	server.OnEvent("/", "botID", func(s socketio.Conn, msg int64) {
		log.Printf("Received bot ID: \"%d\"", msg)

		broker.connectionsLock.RLock()
		if code, ok := broker.connections[s.ID()]; ok {
			// this socket is now listening for mutes that can be applied via that connect code
			s.Join(code)
			killChan := broker.ackKillChannels[s.ID()]
			if killChan != nil {
				go broker.tasksListener(server, code, killChan)
			} else {
				log.Println("Null killchannel for conncode: " + code + ". This means we got a Bot ID before a connect code!")
			}
		}
		broker.connectionsLock.RUnlock()
	})

	server.OnEvent("/", "taskFailed", func(s socketio.Conn, msg string) {
		err := broker.client.SetCaptureTaskStatus(msg, "false")
		if err != nil {
			log.Println("error marking task " + msg + " as unsuccessful")
			log.Println(err)
		}
	})

	server.OnEvent("/", "taskComplete", func(s socketio.Conn, msg string) {
		err := broker.client.SetCaptureTaskStatus(msg, "true")
		if err != nil {
			log.Println("error marking task " + msg + " as successful")
			log.Println(err)
		}
	})

	server.OnEvent("/", "lobby", func(s socketio.Conn, msg string) {
		log.Println("lobby:", msg)

		// validation
		var lobby game.Lobby
		err := json.Unmarshal([]byte(msg), &lobby)
		if err != nil {
			log.Println(err)
		} else {
			broker.connectionsLock.RLock()
			if cCode, ok := broker.connections[s.ID()]; ok {
				event := capture.Event{
					EventType: capture.Lobby,
					Payload:   []byte(msg),
				}
				err := broker.client.AddCaptureEvent(cCode, event)
				if err != nil {
					log.Println(err)
				}
			}
			broker.connectionsLock.RUnlock()
		}
	})
	server.OnEvent("/", "state", func(s socketio.Conn, msg string) {
		log.Println("phase received from capture: ", msg)
		_, err := strconv.Atoi(msg)
		if err != nil {
			log.Println(err)
		} else {
			broker.connectionsLock.RLock()
			if cCode, ok := broker.connections[s.ID()]; ok {
				event := capture.Event{
					EventType: capture.State,
					Payload:   []byte(msg),
				}
				err := broker.client.AddCaptureEvent(cCode, event)
				if err != nil {
					log.Println(err)
				}
			}
			broker.connectionsLock.RUnlock()
		}
	})
	server.OnEvent("/", "player", func(s socketio.Conn, msg string) {
		log.Println("player received from capture: ", msg)

		broker.connectionsLock.RLock()
		if cCode, ok := broker.connections[s.ID()]; ok {
			event := capture.Event{
				EventType: capture.Player,
				Payload:   []byte(msg),
			}
			err := broker.client.AddCaptureEvent(cCode, event)
			if err != nil {
				log.Println(err)
			}
		}
		broker.connectionsLock.RUnlock()
	})
	server.OnEvent("/", "gameover", func(s socketio.Conn, msg string) {
		broker.connectionsLock.RLock()
		if cCode, ok := broker.connections[s.ID()]; ok {
			event := capture.Event{
				EventType: capture.GameOver,
				Payload:   []byte(msg),
			}
			err := broker.client.AddCaptureEvent(cCode, event)
			if err != nil {
				log.Println(err)
			}
		}
		broker.connectionsLock.RUnlock()
	})
	server.OnError("/", func(s socketio.Conn, e error) {
		log.Println("meet error:", e)
	})
	server.OnDisconnect("/", func(s socketio.Conn, reason string) {
		log.Println("Client connection closed: ", reason)

		broker.connectionsLock.RLock()
		if cCode, ok := broker.connections[s.ID()]; ok {
			event := capture.Event{
				EventType: capture.Connection,
				Payload:   []byte("false"),
			}
			err := broker.client.AddCaptureEvent(cCode, event)
			if err != nil {
				log.Println(err)
			}
			server.ClearRoom("/", cCode)
		}
		broker.connectionsLock.RUnlock()

		broker.connectionsLock.Lock()
		if c, ok := broker.ackKillChannels[s.ID()]; ok {
			c <- true
		}
		delete(broker.ackKillChannels, s.ID())
		delete(broker.connections, s.ID())
		broker.connectionsLock.Unlock()
	})
	go server.Serve()
	defer server.Close()

	router := mux.NewRouter()
	router.Handle("/socket.io/", server)

	router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})
	log.Printf("Message broker is running on port %s...\n", port)
	log.Fatal(http.ListenAndServe(":"+port, router))
}
