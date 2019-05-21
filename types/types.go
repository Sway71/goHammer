package types

type Message struct {
	UserId				string				`json:"userId"`
	RoomId				string				`json:"roomId"`
	Type				string				`json:"type"`
	Email    			string 				`json:"email"`
	Username 			string 				`json:"username"`
	Message  			string 				`json:"message"`
	Recipient			string				`json:"recipient"'`
	Data				[]byte				`json:"data"`
}

type Location struct {
	X int	`json:"x"`
	Y int	`json:"y"`
}

type MapTile struct {
	Height			int					`json:"height"`
	Terrain			string				`json:"terrain"`
	UnitIndex		int					`json:"unitIndex"`
	UnitId			int					`json:"unitId"`
}
