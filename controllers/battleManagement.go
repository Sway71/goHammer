package controllers

import (
	"encoding/json"
	"fmt"
	"github.com/jmoiron/sqlx"
	"github.com/mediocregopher/radix.v2/pool"
	"goHammer/types"
	"goHammer/utils"
	"io/ioutil"
	"log"
	"math"
	"net/http"
	"strconv"
	"strings"
)

type BattleManagementController struct {
	RedisPool			*pool.Pool
	DB					*sqlx.DB
	GameMove			chan types.Message
}

type BattleConfiguration struct {
	BattleId				string					`json:"battleId"`
	MapSize					int						`json:"mapSize"`
	Allies					[]int					`json:"allies"`
	// AllyLocations			[]types.Location		`json:"allyLocations"`
	//Enemies					[]int					`json:"enemies"`
	//EnemyLocations			[]types.Location		`json:"enemyLocations"`
}

func (bmController *BattleManagementController) InitializeBattle(w http.ResponseWriter, r *http.Request) {
	b, err := ioutil.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	var battleConfiguration BattleConfiguration
	err = json.Unmarshal(b, &battleConfiguration)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}


	// TODO: Check that this works...
	var allies []Character
	for _, ally := range battleConfiguration.Allies {
		var character Character
		err := bmController.DB.Get(&character, "SELECT id, name, wounds FROM units WHERE id=$1", ally)
		if err != nil {
			log.Fatalln(err)
		}
		allies = append(allies, character)
	}

	//var enemies []Character
	//for _, enemy := range battleConfiguration.Enemies {
	//	var character Character
	//	err := bmController.DB.Get(&character, "SELECT * FROM character WHERE id=$1", enemy)
	//	if err != nil {
	//		log.Fatalln(err)
	//	}
	//	allies = append(enemies, character)
	//}

	conn, err := bmController.RedisPool.Get()
	if err != nil {
		fmt.Println("couldn't get Redis pool connection")
		log.Fatalln(err)
		return
	}
	defer bmController.RedisPool.Put(conn)

	// create battle id to store all pertinent information
	// TODO: create better reference than the teams list
	var battleId string
	exists := 1
	for exists == 1 {
		battleId = "battle:" + utils.RandomString(32)[:30]
		exists, err = conn.Cmd("EXISTS", battleId + ":teams").Int()
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
	}

	// TODO: Determine whether set maps are a thing
	// err = conn.Cmd("SET", battleId + ":mapId", battleConfiguration.MapId).Err

	// Add allies to battle in Redis
	// TODO: add in direction character is facing.
	numAllies := len(battleConfiguration.Allies)
	var generatedAllyLocations []Character

	// add to list of teams
	err = conn.Cmd(
		"RPUSH",
		battleId + ":teams",
		"team_1",
	).Err

	xAlternator := 1
	lastX := 5
	for i := 0; i < numAllies; i++ {
		currAlly := allies[i]
		// TODO: have the reference be the user's unique username or email
		allyRef := battleId + ":team_1:" + strconv.Itoa(i)
		newX := lastX + (i * xAlternator)
		newY := int(9 - (math.Floor(float64(i / 4.0))))

		lastX = newX
		xAlternator = xAlternator * -1

		err = conn.Cmd(
			"HMSET",
			allyRef,
			"id",
			currAlly.Id,
			"index",
			i,
			"x",
			newX,
			"y",
			newY,
			"directionFacing",
			"N",
		).Err
		if err != nil {
			fmt.Println("adding allies error")
			log.Fatalln(err)
		}
		// Adds location to a list of occupied spaces as "x:y"
		err = conn.Cmd(
			"SADD",
			battleId + ":allySpaces",
			strconv.Itoa(newX) + ":" + strconv.Itoa(newY),
			// strconv.Itoa(battleConfiguration.AllyLocations[i].X) + ":" + strconv.Itoa(battleConfiguration.AllyLocations[i].Y),
		).Err
		if err != nil {
			log.Fatalln(err)
		}

		currAlly.X = newX
		currAlly.Y = newY
		generatedAllyLocations = append(generatedAllyLocations, currAlly)
	}

	// TODO: generate obstacles
	// TODO: add function that wipes everything if you encounter an error.

	// TODO: check next steps
	// 1. host gets redirected to game-room and gets a websocket connection + battle URL
	// 2. As other people join with the URL, connect them to room and add their party to the Redis data
	// 3. If at least one person has joined, host should have the opportunity to

	json.NewEncoder(w).Encode(struct {
		BattleId 			string	 				`json:"battleId"`
		Allies				[]Character				`json:"allies"`
	}{
		battleId[7:],
		generatedAllyLocations,
	})
}

func (bmController *BattleManagementController) JoinBattle(w http.ResponseWriter, r *http.Request) {
	b, err := ioutil.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	var battleConfiguration BattleConfiguration
	err = json.Unmarshal(b, &battleConfiguration)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	// TODO: Check that this works...
	var allies []Character
	for _, ally := range battleConfiguration.Allies {
		var character Character
		err := bmController.DB.Get(&character, "SELECT id, name, wounds FROM units WHERE id=$1", ally)
		if err != nil {
			log.Fatalln(err)
		}
		allies = append(allies, character)
	}

	conn, err := bmController.RedisPool.Get()
	if err != nil {
		fmt.Println("couldn't get Redis pool connection")
		log.Fatalln(err)
		return
	}
	defer bmController.RedisPool.Put(conn)

	redisBattleId := "battle:" + battleConfiguration.BattleId

	// check that the battle exists
	exists, err := conn.Cmd("EXISTS", redisBattleId + ":teams").Int()
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	if exists != 1 {
		json.NewEncoder(w).Encode(struct {
			ErrorMessage 			string	 				`json:"errorMessage"`
		}{
			"Sorry, but it seems this battle no longer exists",
		})
		return
	}

	teamsList, err := conn.Cmd("LRANGE", redisBattleId + ":teams", 0, -1).List()
	if len(teamsList) >= 4 {
		json.NewEncoder(w).Encode(struct {
			ErrorMessage 			string	 				`json:"errorMessage"`
		}{
			"Sorry, but this battle is full",
		})
		return
	}

	teamName := "team_" + strconv.Itoa(len(teamsList) + 1)
	err = conn.Cmd(
		"RPUSH",
		redisBattleId + ":teams",
		teamName,
	).Err
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	// adding new team's members
	numAllies := len(battleConfiguration.Allies)
	var generatedAllyLocations []Character
	xAlternator := 1
	lastX := 5
	for i := 0; i < numAllies; i++ {
		currAlly := allies[i]
		// TODO: have the reference be the user's unique username or email
		allyRef := redisBattleId + ":" + teamName + ":" + strconv.Itoa(i)
		newX := lastX + (i * xAlternator)
		newY := int(0 + (math.Floor(float64(i / 4.0))))

		lastX = newX
		xAlternator = xAlternator * -1

		err = conn.Cmd(
			"HMSET",
			allyRef,
			"id",
			currAlly.Id,
			"index",
			i,
			"x",
			newX,
			"y",
			newY,
			"directionFacing",
			"S",
		).Err
		if err != nil {
			fmt.Println("adding allies error")
			log.Fatalln(err)
		}
		// Adds location to a list of occupied spaces as "x:y"
		err = conn.Cmd(
			"SADD",
			redisBattleId + ":" + teamName + "_spaces",
			strconv.Itoa(currAlly.Id) + ":" + strconv.Itoa(newX) + ":" + strconv.Itoa(newY),
			// strconv.Itoa(battleConfiguration.AllyLocations[i].X) + ":" + strconv.Itoa(battleConfiguration.AllyLocations[i].Y),
		).Err
		if err != nil {
			log.Fatalln(err)
		}

		currAlly.X = newX
		currAlly.Y = newY
		generatedAllyLocations = append(generatedAllyLocations, currAlly)
	}

	var teamRosters [][]Character
	for _, team := range teamsList {
		unitLocations, err := conn.Cmd("SMEMBERS", redisBattleId + ":" + team + "_spaces").List()
		if err != nil {

		}

		var teamUnits []Character
		var character Character
		for _, unitLocation := range unitLocations {
			err := bmController.DB.Get(
				&character,
				"SELECT id, name, wounds FROM units WHERE id=$1",
				strings.Split(unitLocation, ":")[0],
			)
			if err != nil {
				log.Fatalln(err)
			}

			teamUnits = append(teamUnits, character)
		}

		teamRosters = append(teamRosters, teamUnits)
	}

	teamData, _ := json.Marshal(teamRosters)

	bmController.GameMove <- types.Message{
		UserId: "server",
		RoomId: redisBattleId[7:],
		Type: "BOARD_UPDATE",
		Email: "server",
		Username: "server",
		Message: "",
		Recipient: "",
		Data: teamData,
	}

	json.NewEncoder(w).Encode(struct {
		BattleId 			string	 				`json:"battleId"`
		Allies				[]Character				`json:"allies"`
		Enemies				[]Character				`json:"enemies"`
	}{
		// battleConfiguration.BattleId,

	})
}