package data

import (
	"strconv"

	"github.com/golang-jwt/jwt/v5"
)

type TokenData map[string]string

func (td TokenData) ToMap() jwt.MapClaims {
	res := jwt.MapClaims{}

	for key, value := range td {
		res[key] = value
	}

	return res
}

func TokenDataFromMap(mapClaims jwt.MapClaims) TokenData {
	res := TokenData{}

	for key, value := range mapClaims {
		switch v := value.(type) {
		case float64:
			res[key] = strconv.FormatInt(int64(v), 10)
		case string:
			res[key] = v
		}
	}

	return res
}
