package utility

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"strconv"
	"strings"
	"../appError"
)

// Target key should be a pointer to struct
func LoadJsonFromFile(path string, targetKey interface{}) {
	// Read it from file
	raw, err := ioutil.ReadFile("../config/config.json")
	if err != nil {
		log.Println(err)
	}

	err = json.Unmarshal(raw, targetKey)
	if err!=nil {
		log.Println(err)
	}
}

type Params struct  {
	IntegerParams []string
	FloatParams []string
	IntegerArrayParams []string
}

type ParsedParams struct{
	IntegerParams map[string]int
	FloatParams map[string]float64
	IntegerArrayParams map[string][]int
}

// Will convert string to int/float if value for that field exists in data
func ReadValueFromStringMap(args Params, data map[string]string) (ParsedParams) {
	var result ParsedParams
	integerMap := make(map[string]int)
	integerArrayMap := make(map[string][]int)
	// Parsing integer params
	for _,key := range args.IntegerParams {
		integerMap[key] = 0
		str,present := data[key]
		log.Println("key,present",key,present)
		if !present || str == "" {
			continue
		}
		val, err := strconv.Atoi(str)
		if err != nil {
			panic(appError.NewValidationError(key+" must be an integer"))
		} 
		integerMap[key] = val
	}
	result.IntegerParams = integerMap

	// Parsing IntegerArray params
	for _,key := range args.IntegerArrayParams {
		var temp []int
		serializeString, present := data[key]
		if !present || serializeString == "" {
			continue
		}
		valuePassed := strings.Split(serializeString, "@")
		for _,str := range valuePassed {
			strInt, err := strconv.Atoi(str)
			if err != nil {
				panic(appError.NewValidationError(key+" not in correct format"))
			}
			temp = append(temp, strInt)	
		}
		integerArrayMap[key] = temp
	}
	result.IntegerArrayParams = integerArrayMap
	return result
}