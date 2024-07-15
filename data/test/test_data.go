package main

import (
	"fmt"
	"ga4ct/data"
	"log"
)

func main() {
	es, err := data.NewEventStore()
	if err != nil {
		log.Fatal(err)
	}
	defer es.Close()

	// Test: No dimensions or measures
	// _, err = es.NewQuery().Dimensions().Measures().Query()
	// if err != nil {
	// 	fmt.Println("Test 1: Errored correctly")
	// }

	// Test: Only dimensions
	// res, err := es.NewQuery().Dimensions(data.EventName).Measures().Limit(10).Query()
	// if err != nil {
	// 	log.Fatal(err)
	// }

	// table, err := res.Table()
	// if err != nil {
	// 	log.Fatal(err)
	// }

	// fmt.Printf("%v\n", *table)

	// // Test: Only measures
	// res, err = es.NewQuery().Dimensions().Measures(data.EventCount).Limit(10).Query()
	// if err != nil {
	// 	log.Fatal(err)
	// }

	// table, err = res.Table()
	// if err != nil {
	// 	log.Fatal(err)
	// }

	// fmt.Printf("%v\n", *table)

	// Test: Both
	res, err := es.NewQuery().Dimensions(data.Date).Measures(data.EventCount).Limit(10).Query()
	if err != nil {
		log.Fatal(err)
	}

	chart, err := res.LineChart()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("%v\n", *chart)
}
