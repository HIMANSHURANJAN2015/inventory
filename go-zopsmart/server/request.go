package server

import (
	"net/http"
	"strings"
	"strconv"
	"../appError"
	"../utility"
)

const (
	Int = "integer"
	Float = "float"
	Bool = "bool"
	IntArray = "intarray"
)

type ParsedParams struct {
	IntegerParams map[string]int
	FloatParams map[string]float64
	BoolParams map[string]bool
	IntegerArrayParams map[string][]int
}

// Will return ["", "entity", "<id>"]. After removing query params
func GetURIParts(r *http.Request) []string {
	url := r.RequestURI
	uriParts := strings.Split(url, "?")
	return strings.Split(uriParts[0], "/")
}


func fetchId(r *http.Request) (int) {
	parts := GetURIParts(r)
	if len(parts) > 2 {
		id := parts[2]
		intId, err := strconv.Atoi(id)
		if err == nil {
			return intId
		}
	}
	return 0
}

/*
1. If mandatory Paramters are not present and not in correct format then it will give th error
2. Optional Paramters will be initialized to there empty value, if not present, in each of the "Type maps" to which it belongs. 
   Hence the key will always be present
3. If value of a parameter can be multi type, then the parameter key will be present in all the type Maps initialized with
	empty value. Hower the actual group wil have the actual data passed.
*/
func GetRequestParams(r *http.Request, mandatoryFields, optionalFields map[string][]string) (ParsedParams) {
	r.ParseForm()
	var parsedParams ParsedParams

	integerParams := make(map[string]int)
	floatParams := make(map[string]float64)
	boolParams := make(map[string]bool)
	integerArrayParams := make(map[string][]int)

	for key, keyTypes := range mandatoryFields {
		valueFromRequest := r.Form[key]
		if len(valueFromRequest) == 0 {
			panic(appError.NewValidationError("Missing Required Field : "+key))		
		}
		parseParams(key, valueFromRequest, keyTypes,  integerParams, floatParams, boolParams, integerArrayParams)
	}
	for key, keyTypes := range optionalFields {
		valueFromRequest := r.Form[key]
		parseParams(key, valueFromRequest, keyTypes,  integerParams, floatParams, boolParams, integerArrayParams)
	}
	// Fetching id from request params
	integerParams["id"] = fetchId(r)

	parsedParams.IntegerParams = integerParams
	parsedParams.FloatParams = floatParams
	parsedParams.BoolParams = boolParams
	parsedParams.IntegerArrayParams = integerArrayParams
	return parsedParams
}

func parseParams(key string, valueFromRequest []string, keyTypes []string, integerParams map[string]int, floatParams map[string]float64, boolParams map[string]bool, integerArrayParams map[string][]int) {
	var err error
	switch {
		case utility.StringInSlice(Int, keyTypes) && len(valueFromRequest) == 1: // If more than 1 value is there, then it will go to Array
			val := 0
			if len(valueFromRequest) == 1 {
				val, err = strconv.Atoi(valueFromRequest[0])
				if err != nil {
					panic(appError.NewValidationError(key+" must be an integer"))
				}
			}
			integerParams[key] = val
		case utility.StringInSlice(Float, keyTypes) && len(valueFromRequest) == 1:
			var val float64 = 0
			if len(valueFromRequest) == 1 {
				val, err = strconv.ParseFloat(valueFromRequest[0], 64)
				if err != nil {
					panic(appError.NewValidationError(key+" must be an float"))
				}
			}
			floatParams[key] = val
		case utility.StringInSlice(Bool, keyTypes) && len(valueFromRequest) == 1:
			val := false
			if len(valueFromRequest) == 1 { 
				val, err = strconv.ParseBool(valueFromRequest[0])
				if err != nil {
					panic(appError.NewValidationError(key+" must be an boolean"))	
				}	
			}	
			boolParams[key] = val
		case utility.StringInSlice(IntArray, keyTypes) && len(valueFromRequest) > 1:
			var temp []int //nil
			for _,str := range valueFromRequest {
				val, err := strconv.Atoi(str)
				if err != nil {
					panic(appError.NewValidationError(key+" must be an integer"))
				}
				temp = append(temp, val)
			}
			integerArrayParams[key] = temp	
	}
}