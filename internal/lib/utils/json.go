package utils

import (
	"encoding/json"
	"fmt"
)

func PrintAsJSON(data interface{}) (*[]byte, error) {
	//    var err := error
	p, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	return &p, nil
}
