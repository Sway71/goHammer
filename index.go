package main

import (
	_ "github.com/lib/pq"
	"goHammer/types"

	"fmt"
	"log"
	"net/http"
	"sync"

	"goHammer/controllers"

	"github.com/gorilla/websocket"
	"github.com/husobee/vestigo"
	"github.com/jmoiron/sqlx"
	"github.com/mediocregopher/radix.v2/pool"
)

func main() {

	dbInfo := fmt.Sprintf(
		"user='vick michael' password='' dbname='postgres' sslmode=disable",
	)
	db, err := sqlx.Connect("postgres", dbInfo)
	if err != nil {
		fmt.Println("check your Postgres connection")
		panic(err)
	}

	RedisPool, redisErr := pool.New("tcp", "localhost:6379", 10)
	if redisErr != nil {
		fmt.Println("check your Redis connection")
		panic(err)
	}

	var upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin: func(r *http.Request) bool {
			// TODO: actually check the origin
			return true
		},
	}
	var Clients = sync.Map{}
	var Broadcast = make(chan types.Message)
	var DirectMessage = make(chan types.Message)
	var RoomMessage = make(chan types.Message)
	var GameMove = make(chan types.Message)

	// Controller declarations (Redis, Postgres)
	characterController := controllers.CharacterController{RedisPool, db}
	battleManager := controllers.BattleManagementController{
		RedisPool,
		db,
		GameMove,
	}
	WebsocketController := controllers.WebsocketController{
		RedisPool,
		db,
		upgrader,
		Clients,
		Broadcast,
		DirectMessage,
		RoomMessage,
		GameMove,
	}
	// battleManager := BattleManagementController{redisPool, db}

	// Router declaration
	router := vestigo.NewRouter()

	// Setting up router global  CORS policy
	// These policy guidelines are overriddable at a per resource level shown below
	router.SetGlobalCors(&vestigo.CorsAccessControl{
		AllowOrigin:      []string{"*"},
		AllowCredentials: true,
		// ExposeHeaders:    []string{"X-Header", "X-Y-Header"},
		AllowHeaders:     []string{"Access-Control-Allow-Headers", "Content-Type"},
	})

	// Websocket routes
	router.Get("/init-websocket", WebsocketController.HandleConnections)

	// Characters routes
	router.Get("/characters", characterController.GetCharacters)
	router.Get("/characters/:id", characterController.GetCharacter)
	router.Post("/characters/create", characterController.CreateCharacter)

	// Characters' movement routes
	// TODO: transfer these routes and the columns in postgres to battle management routes and Redis data
	//router.Get("/battle/:battleId/movableLocations/:id", characterController.getMovableSpaces)
	//router.Post("/characters/:battleId/move/:id", characterController.move)
	//router.Post("/characters/:battleId/attack/:id", characterController.attack)

	// Battle managing routes
	router.Post("/battle/initialize", battleManager.InitializeBattle)
	router.Post("/battle/join", battleManager.JoinBattle)

	go WebsocketController.HandleBroadcasts(Broadcast)
	go WebsocketController.HandleDirectMessages(DirectMessage)
	go WebsocketController.HandleRoomMessages(RoomMessage)
	go WebsocketController.HandleGameMoves(GameMove)
	fmt.Println("Listening on localhost:1337")
	log.Fatal(http.ListenAndServe(":1337", router))
}