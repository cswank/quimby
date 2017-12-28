package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/cswank/brewery/recipes"
)

func GetRecipe(w http.ResponseWriter, req *http.Request) error {
	args := GetArgs(req)
	recipe, err := recipes.NewRecipe(args.Vars["name"])
	if err != nil {
		return err
	}
	temp := getTemperature(args)
	method := recipe.GetMethod(temp)
	return json.NewEncoder(w).Encode(method)
}

func getTemperature(args *Args) float64 {
	t, _ := strconv.ParseFloat(args.Args.Get("grain_temperature"), 64)
	return t
}
