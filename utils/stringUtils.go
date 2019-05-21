package utils

import (
	"strconv"
	"strings"

	"goHammer/types"
)

func ConvertStringToCoords(stringArray []string) []types.Location {
	output := make([]types.Location, len(stringArray))
	for i, value := range stringArray {
		currPoint := strings.Split(value, ":")
		// TODO: error handling below...
		xPoint, _ := strconv.Atoi(currPoint[0])
		yPoint, _ := strconv.Atoi(currPoint[1])
		output[i] = types.Location{xPoint, yPoint}
	}
	return output
}
