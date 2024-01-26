package main

import (
	"GoCrpto/orderbook"
	"fmt"
	"reflect"
	"testing"
)

func assert(t *testing.T, a any, b any) {
	if !reflect.DeepEqual(a, b) {
		t.Errorf("%+v != %+v", a, b)
	}
}

func TestLimit(t *testing.T) {

	l := orderbook.NewLimit(10_000)
	buyOrder := orderbook.NewOrder(true, 5, 0)
	buyOrderA := orderbook.NewOrder(true, 8, 0)
	buyOrderB := orderbook.NewOrder(true, 10, 0)
	buyOrderC := orderbook.NewOrder(true, 15, 0)

	l.AddOrder(buyOrder)
	l.AddOrder(buyOrderA)
	l.AddOrder(buyOrderB)
	l.AddOrder(buyOrderC)

	l.DeleteOrder(buyOrderA)

	fmt.Println(l)

}

func TestOrderBook(t *testing.T) {
	ab := orderbook.NewOrderBook()
	buyOrder := orderbook.NewOrder(true, 10, 0)
	buyOrderA := orderbook.NewOrder(true, 2000, 0)

	ab.Placeholder(18_000, buyOrder)
	ab.Placeholder(19_000, buyOrderA)

	fmt.Printf("%+v", ab)
}
func TestPlaceLimitOrder(t *testing.T) {
	ab := orderbook.NewOrderBook()
	sellOrderA := orderbook.NewOrder(false, 10, 0)
	sellOrderB := orderbook.NewOrder(false, 5, 0)
	ab.PlaceLimitOrder(10000, sellOrderA)
	ab.PlaceLimitOrder(9000, sellOrderB)
	assert(t, len(ab.Orders), 2)
	//	assert(t,len(ab.Orders[sellOrderA.ID],sellOrderA)
	assert(t, len(ab.Asks()), 2)

}

func TestPlaceMarketOrder(t *testing.T) {
	ob := orderbook.NewOrderBook()

	sellOrder := orderbook.NewOrder(false, 20, 0)
	ob.PlaceLimitOrder(10_000, sellOrder)

	buyOrder := orderbook.NewOrder(true, 10, 0)
	matches := ob.PlaceMarketOrder(buyOrder)

	assert(t, len(matches), 1)
	assert(t, len(ob.Asks()), 1)
	assert(t, ob.AskTotalVolume(), 10.0)
	assert(t, matches[0].Ask, sellOrder)
	assert(t, matches[0].Bid, buyOrder)
	assert(t, matches[0].SizeFilled, 10.0)
	assert(t, matches[0].Price, 10_000.0)
	//	assert(t, buyOrder.isFilled(), true)

	//	fmt.Printf("%+v", matches)
}

func TestPlaceMarketOrderMultiFill(t *testing.T) {
	ob := orderbook.NewOrderBook()
	buyOrderA := orderbook.NewOrder(true, 5, 0)
	buyOrderB := orderbook.NewOrder(true, 8, 0)
	buyOrderC := orderbook.NewOrder(true, 10, 0)
	buyOrderD := orderbook.NewOrder(true, 1, 0)

	ob.PlaceLimitOrder(5_000, buyOrderA)
	ob.PlaceLimitOrder(5_000, buyOrderB)
	ob.PlaceLimitOrder(9_000, buyOrderC)
	ob.PlaceLimitOrder(10_000, buyOrderD)

	assert(t, ob.BidTotalVolume(), 24.00)

	sellOrder := orderbook.NewOrder(false, 20, 0)
	matches := ob.PlaceMarketOrder(sellOrder)
	assert(t, ob.BidTotalVolume(), 4.0)
	assert(t, len(matches), 4)
	assert(t, len(ob.Bids()), 2)

	//fmt.Printf("%+v", matches)
}

func TestCancelOrder(t *testing.T) {
	ob := orderbook.NewOrderBook()

	buyOrder := orderbook.NewOrder(true, 4, 0)
	ob.PlaceLimitOrder(10000.0, buyOrder)

	assert(t, ob.BidTotalVolume(), 4.0)
	ob.CancelOrder(buyOrder)
	assert(t, ob.BidTotalVolume(), 0.0)

	_, ok := ob.Orders[buyOrder.ID]
	assert(t, ok, false)
}
