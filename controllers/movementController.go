package controllers

import (
	"fmt"

	"goHammer/types"
	"goHammer/utils"
)

var directions = []types.Location{
	{0, 1},
	{1, 0},
	{0, -1},
	{-1, 0},
}


func ContainsPoint(pointsSlice []types.Location, pointToCheck types.Location) bool {
	result := false
	for _, point := range pointsSlice {
		if point.X == pointToCheck.X && point.Y == pointToCheck.Y {
			result = true
		}
	}
	return result
}

func getNeighbors(
	move int,
	jump int,
	playerLocation types.Location,
	terrainMap [][]types.MapTile,
	movableLocations *[]types.Location,
	allySpaces []types.Location,
	enemySpaces []types.Location,
) {

	var currentTile = terrainMap[playerLocation.X][playerLocation.Y]

	for _, direction := range directions {
		newX, newY := playerLocation.X+direction.X, playerLocation.Y+direction.Y
		var neighbor = types.Location{
			newX,
			newY,
		}

		if neighbor.X >= 0 && neighbor.X <= len(terrainMap)-1 && neighbor.Y >= 0 && neighbor.Y <= len(terrainMap[0])-1 {
			neighborTile := terrainMap[neighbor.X][neighbor.Y]

			// in reality the following items still need to be taken into account:
			// 1. a character can jump down further than they can jump up, ?(jump vs. jump - 1)?	#needsTesting
			// 2. the cost for jumping up is higher than jumping down (2 vs 1)						#needsTesting
			// 3. certain types of terrain cost more to move across (probably only matters for water)
			// 4. jumping across gaps or over enemies (long term goal)
			// 5. occupied (thus not movable) spaces for allies (passable) and enemies (blocked)	#needsTesting
			cost := 1

			if neighborTile.Height-currentTile.Height > jump/2 {
				cost = 2
			} else if neighborTile.Terrain == "deep_water" {
				cost = move
			}
			//if math.Abs(float64(neighborTile.Height-currentTile.Height)) < float64(jump) {
			if (neighborTile.Height-currentTile.Height < jump ||
				currentTile.Height-neighborTile.Height <= jump) &&
				!ContainsPoint(enemySpaces, neighbor) {

				// recursively look for other movable locations if player still has moves left
				if move > cost {
					getNeighbors(
						move-cost,
						jump,
						neighbor,
						terrainMap,
						movableLocations,
						allySpaces,
						enemySpaces,
					)
				}

				if !ContainsPoint(*movableLocations, neighbor) && !ContainsPoint(allySpaces, neighbor) {
					*movableLocations = append(*movableLocations, neighbor)
				}
			}
		}

	}
}

func GetMovableSpaces(
	move int,
	jump int,
	playerLocation types.Location,
	terrainMap [][]types.MapTile,
	allySpaces []string,
	enemySpaces []string,
) []types.Location {
	var movableLocations = []types.Location{playerLocation}
	allySpacesClean := utils.ConvertStringToCoords(allySpaces)
	enemySpacesClean := utils.ConvertStringToCoords(enemySpaces)

	getNeighbors(move, jump, playerLocation, terrainMap, &movableLocations, allySpacesClean, enemySpacesClean)

	return movableLocations
}

func GetPath(
	move int,
	jump int,
	playerLocation types.Location,
	goalLocation types.Location,
	terrainMap [][]types.MapTile,
	enemySpaces []string,
) []types.Location {
	bestPath := []types.Location{goalLocation}

	// TODO: a lot of the conditionals below are repeats of algorithm above. extract into function if applicable
	// TODO: should I use some of the logic below to completely replace the algorithm above?
	var currentTile = terrainMap[playerLocation.X][playerLocation.Y]

	frontier := []types.Location{playerLocation}
	cameFromMap := map[types.Location]types.Location{playerLocation: playerLocation}
	costSoFarMap := map[types.Location]int{playerLocation: 0}

	// TODO: the below method is only partially done, I need to use cameFromMap and costSoFarMap to reverse-lookup the path
	for i := 1; i <= move; i++ {
		for _, frontierPoint := range frontier {
			for _, direction := range directions {

				// creating the location, making sure it exists, then adding it to the correct group based on cost
				newLocation := types.Location{frontierPoint.X+direction.X, frontierPoint.Y+direction.Y}
				if newLocation.X >= 0 && newLocation.X <= len(terrainMap)-1 &&
					newLocation.Y >= 0 && newLocation.Y <= len(terrainMap[0])-1 {

					// TODO: populate maps and do early return if at goal unless we want to randomize paths
					newTile := terrainMap[newLocation.X][newLocation.Y]
					if !ContainsPoint(frontier, newLocation) &&
						!ContainsPoint(utils.ConvertStringToCoords(enemySpaces), newLocation)  {
						frontier = append(frontier, newLocation)
						cameFromMap[newLocation] = frontierPoint
						if newTile.Height-currentTile.Height > jump/2 {
							costSoFarMap[newLocation] = costSoFarMap[frontierPoint] + 2
						} else if newTile.Terrain == "deep_water" {
							costSoFarMap[newLocation] = move
						} else {
							costSoFarMap[newLocation] = costSoFarMap[frontierPoint] + 1
						}
					}
				}

			}
		}
	}

	fmt.Println(frontier)
	fmt.Printf("%v \n", costSoFarMap)
	fmt.Printf("%v \n", cameFromMap)

	currLocation := goalLocation
	var nextLocation types.Location
	for currLocation.X != playerLocation.X && currLocation.Y != playerLocation.Y {
		nextLocation = cameFromMap[currLocation]
		bestPath = append(bestPath, nextLocation)
		currLocation = nextLocation
	}
	fmt.Printf("bestPath: %v \n", bestPath)
	return bestPath
}