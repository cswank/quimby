package handlers

import (
	"encoding/json"
	"strconv"

	"github.com/cswank/brewery/recipes"
)

func GetRecipe(args *Args) error {
	recipe, err := recipes.NewRecipe(args.Vars["name"])
	if err != nil {
		return err
	}
	temp := getTemperature(args)
	method := recipe.GetMethod(temp)
	enc := json.NewEncoder(args.W)
	return enc.Encode(method)
}

func getTemperature(args *Args) float64 {
	t, _ := strconv.ParseFloat(args.Args.Get("grain_temperature"), 64)
	return t
}
