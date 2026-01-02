package tools

import (
	"fmt"
)

func BookHotel(args map[string]interface{}) string {
	hotelName, _ := args["hotel_name"].(string)
	return fmt.Sprintf("成功预定酒店: %s", hotelName)
}

var BookHotelTool = &Tool{
	Name:        "book_hotel",
	Description: "预定酒店，需要提供酒店名称",
	Parameters: map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"hotel_name": map[string]interface{}{
				"type":        "string",
				"description": "酒店名称",
			},
		},
		"required": []string{"hotel_name"},
	},
	Func: BookHotel,
}

func BookFlight(args map[string]interface{}) string {
	from, _ := args["from_airport"].(string)
	to, _ := args["to_airport"].(string)
	return fmt.Sprintf("成功预定机票: %s → %s", from, to)
}

var BookFlightTool = &Tool{
	Name:        "book_flight",
	Description: "预定机票，需要提供出发地和目的地",
	Parameters: map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"from_airport": map[string]interface{}{
				"type":        "string",
				"description": "出发机场",
			},
			"to_airport": map[string]interface{}{
				"type":        "string",
				"description": "到达机场",
			},
		},
		"required": []string{"from_airport", "to_airport"},
	},
	Func: BookFlight,
}
