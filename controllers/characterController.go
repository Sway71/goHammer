package controllers

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"

	"github.com/husobee/vestigo"
	"github.com/jmoiron/sqlx"
	"github.com/mediocregopher/radix.v2/pool"
)

type Character struct {
	Id 				int				`json:"id"`
	Name 			string			`json:"name"`
	//Level			int				`json:"level"`
	//Exp				int				`json:"exp"`
	Wounds			int				`json:"wounds"`
	MaxWounds		int				`json:"maxWounds"`
	WeaponSkill		int				`json:"weaponSkill"`
	BallisticSkill	int				`json:"ballisticSkill"`
	Strength		int				`json:"strength"`
	Toughness		int				`json:"toughness"`
	Initiative		int				`json:"initiative"`
	Move			int				`json:"move"`
	Attacks			int				`json:"attacks"`
	Leadership		int				`json:"leadership"`
	X				int				`json:"x"`
	Y				int				`json:"y"`
}

type CharacterController struct {
	RedisPool		*pool.Pool
	DB				*sqlx.DB
}

func (c *CharacterController) GetCharacters(w http.ResponseWriter, r *http.Request) {
	var characters []Character
	err := c.DB.Select(&characters, "SELECT * FROM character ORDER BY id")
	if err != nil {
		log.Fatalln(err)
	}

	json.NewEncoder(w).Encode(characters)
}

func (c *CharacterController) GetCharacter(w http.ResponseWriter, r *http.Request) {
	var character Character
	id := vestigo.Param(r, "id")

	err := c.DB.Get(&character, "SELECT * FROM character WHERE id=$1", id)
	if err != nil {
		log.Fatalln(err)
	}

	json.NewEncoder(w).Encode(character)
}

func (c *CharacterController) CreateCharacter(w http.ResponseWriter, r *http.Request) {
	b, err := ioutil.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	var character Character
	err = json.Unmarshal(b, &character)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	var lastInsertId int;
	characterInsert := `
		INSERT INTO character (
			name,
			level,
			exp,
			hp,
			mp,
			move,
			jump,
			speed,
			x,
			y
		  ) VALUES (
		  	$1,
		  	$2,
		  	$3,
		  	$4,
		  	$5,
		  	$6,
		  	$7,
		  	$8,
		  	$9,
		  	$10,
			$11,
			$12,
			$13,
			$14
		  ) RETURNING id
	`
	hp := rand.Intn(10) + 25
	mp := rand.Intn(5) + 8
	err = c.DB.QueryRow(
		characterInsert,
		character.Name,
		0,
		0,
		hp,
		mp,
		3,
		3,
		5,
		0,
		0,
	).Scan(&lastInsertId)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	json.NewEncoder(w).Encode(struct {
		Id		int		`json:"id"`
	}{
		lastInsertId,
	})
}