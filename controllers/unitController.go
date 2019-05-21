package controllers

type Unit struct {
	Id 					int				`json:"id"`
	Faction 			string			`json:"faction"`
	Name 				string			`json:"name"`
	Wounds				int				`json:"wounds"`
	WeaponSkill			int				`json:"weaponSkill"`
	BallisticSkill		int				`json:"ballisticSkill"`
	Strength			int				`json:"strength"`
	Toughness			int				`json:"toughness"`
	Initiative			int				`json:"initiative"`
	Move				int				`json:"move"`
	Attacks				int				`json:"attacks"`
	Leadership			int				`json:"leadership"`
	Points				int				`json:"points"`
}