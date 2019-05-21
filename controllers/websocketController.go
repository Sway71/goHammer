package controllers

import (
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/jmoiron/sqlx"
	"github.com/mediocregopher/radix.v2/pool"
	"goHammer/types"
	"reflect"
	"strings"

	"goHammer/utils"

	"log"
	"net/http"
	"sync"
)

// TODO: remove broadcast channel
type WebsocketController struct {
	RedisPool			*pool.Pool
	DB					*sqlx.DB
	Upgrader			websocket.Upgrader
	Clients				sync.Map
	Broadcast			chan types.Message
	DirectMessage		chan types.Message
	RoomMessage			chan types.Message
	GameMove			chan types.Message
}

// TODO: look into authentication with JWT through websocket. JWT might make some of the fields below obsolete.
//type Message struct {
//	UserId				string				`json:"userId"`
//	RoomId				string				`json:"roomId"`
//	Type				string				`json:"type"`
//	Email    			string 				`json:"email"`
//	Username 			string 				`json:"username"`
//	Message  			string 				`json:"message"`
//	Recipient			string				`json:"recipient"'`
//	Data				[]byte				`json:"data"`
//}

func (wsController *WebsocketController) HandleBroadcasts(broadcast chan types.Message) {
	for {
		// Grab the next message from the broadcast channel
		msg := <-broadcast

		// Send it out to every client that is currently connected
		wsController.Clients.Range(func(key, value interface{}) bool {

			// get client ID and their connection with reflect
			clientId := reflect.ValueOf(key).Interface().(string)
			clientSocketInterface, _ := wsController.Clients.Load(clientId)
			clientSocket := reflect.ValueOf(clientSocketInterface).Interface().(*websocket.Conn)

			// send the message
			clientSocket.WriteJSON(msg)

			return true
		})
	}
}

func (wsController *WebsocketController) HandleDirectMessages(directMessage chan types.Message) {
	for {
		// Grab the next message from the broadcast channel
		msg := <-directMessage
		fmt.Println("Direct message: ", msg.Message)


		clientSocketInterface, _ := wsController.Clients.Load(msg.Recipient)
		clientSocket := reflect.ValueOf(clientSocketInterface).Interface().(*websocket.Conn)

		// send the message
		clientSocket.WriteJSON(msg)

		// Send it out to every client that is currently connected
		//wsController.Clients.Range(func(key, value interface{}) bool {
		//	fmt.Println(msg.Message)
		//
		//	// get client ID and their connection with reflect
		//	clientId := reflect.ValueOf(key).Interface().(string)
		//	clientSocketInterface, _ := wsController.Clients.Load(clientId)
		//	clientSocket := reflect.ValueOf(clientSocketInterface).Interface().(*websocket.Conn)
		//
		//	// send the message
		//	clientSocket.WriteJSON(msg)
		//
		//	return true
		//})
	}
}

func (wsController *WebsocketController) HandleRoomMessages(roomMessage chan types.Message) {
	for {
		// Grab the next message from the broadcast channel
		msg := <-roomMessage
		fmt.Println("Room message: ", msg.Message)

		// Send it out to every client that has the same room ID
		wsController.Clients.Range(func(key, value interface{}) bool {
			fmt.Println(msg.Message)

			// get client ID and their connection with reflect
			clientId := reflect.ValueOf(key).Interface().(string)
			clientSocketInterface, _ := wsController.Clients.Load(clientId)
			clientSocket := reflect.ValueOf(clientSocketInterface).Interface().(*websocket.Conn)

			// TODO: format the message based on its purpose

			// BOARD_UPDATE is a type can will not be accepted externally for security reasons
			if msg.Type == "BOARD_UPDATE" {

			}
			// send the message
			clientSocket.WriteJSON(msg)

			return true
		})
	}
}

func (wsController *WebsocketController) HandleGameMoves(gameMove chan types.Message) {
	for {
		// Grab the next message from the broadcast channel
		msg := <-gameMove
		fmt.Println("Game move: ", msg.Message)

		// Send it out to every client that is currently connected
		//wsController.Clients.Range(func(key, value interface{}) bool {
		//	fmt.Println(msg.Message)
		//
		//	// get client ID and their connection with reflect
		//	clientId := reflect.ValueOf(key).Interface().(string)
		//	clientSocketInterface, _ := wsController.Clients.Load(clientId)
		//	clientSocket := reflect.ValueOf(clientSocketInterface).Interface().(*websocket.Conn)
		//
		//	// send the message
		//	clientSocket.WriteJSON(msg)
		//
		//	return true
		//})
	}
}

func (wsController *WebsocketController) HandleConnections(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Websocket endpoint hit")
	// Upgrade initial GET request to a websocket

	newWebsocket, err := wsController.Upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Fatal(err)
	}
	// Make sure we close the connection when the function returns
	// TODO: check to make sure the connection reference in the map is removed as well
	defer newWebsocket.Close()

	var clientId string
	for clientId == "" {
		clientId = utils.RandomString(16)
		if _, ok := wsController.Clients.Load(clientId); ok {
			clientId = ""
		}
	}

	wsController.Clients.Store(clientId, newWebsocket)
	newWebsocket.WriteJSON(types.Message{
		clientId,
		"none",
		"HANDSHAKE",
		"email@email.com",
		"myNameIsWhat?",
		"Connection established User #" + clientId,
		clientId,
		nil,
	})

	for {
		var msg types.Message
		// Read in a new message as JSON and map it to a Message object
		err := newWebsocket.ReadJSON(&msg)
		if err != nil {
			log.Printf("error: %v", err)
			wsController.Clients.Delete(newWebsocket)
			break
		}
		// Send the newly received message to the correct channel
		if strings.ToUpper(msg.Type) == "BROADCAST" {
			wsController.Broadcast <- msg
		} else if strings.ToUpper(msg.Type) == "MESSAGE" {
			wsController.DirectMessage <- msg
		} else if strings.ToUpper(msg.Type) == "ROOM" {
			wsController.RoomMessage <- msg
		} else if strings.ToUpper(msg.Type) == "GAME_MOVE" {
			wsController.GameMove <- msg
		}

	}
}