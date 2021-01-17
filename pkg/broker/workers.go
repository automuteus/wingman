package broker

import (
	"encoding/json"
	socketio "github.com/googollee/go-socket.io"
	"log"
)

// tasksListener is a worker that waits for incoming mute/deafen tasks for a particular game (via a Redis pub/sub).
// This worker broadcasts the task to the Socket.io room associated with the connectcode, to see if any Capture Bots can
// issue the task. This worker doesn't care about the response; the response is handled/passed on by the socket.io server
// itself in broker.go
func (broker *Broker) tasksListener(server *socketio.Server, connectCode string, killchan <-chan bool) {
	for {
		select {
		case <-killchan:
			// TODO should also short-circuit the long-poll request being executed currently
			return

		default:
			task, err := broker.client.GetCaptureTask(connectCode)
			if err != nil {
				log.Println(err)
			} else if task != nil {
				jBytes, err := json.Marshal(task)
				if err != nil {
					log.Println(err)
				} else {
					log.Println("Broadcasting " + string(jBytes) + " to room " + connectCode)
					server.BroadcastToRoom("/", connectCode, "modify", jBytes)
				}
			}
		}
	}
}

//
//// ackWorker functions as a healthcheck for the bot, if the bot resumes a game (by connectcode) when it starts up after
//// being down/offline. This worker receives the ack, and, if the connection is still active, responds with a
//// connection=true message. If the client has terminated the connection, then this worker is also terminated, and thus
//// the bot never receives a healthy response for the connectcode/game in question
//func (broker *Broker) ackWorker(ctx context.Context, connCode string, killChan <-chan bool) {
//	pubsub := task.AckSubscribe(ctx, broker.client, connCode)
//	channel := pubsub.Channel()
//	defer pubsub.Close()
//
//	for {
//		select {
//		case <-killChan:
//			return
//		case <-channel:
//			err := task.PushJob(ctx, broker.client, connCode, task.ConnectionJob, "true")
//			if err != nil {
//				log.Println(err)
//			}
//			break
//		}
//	}
//}
