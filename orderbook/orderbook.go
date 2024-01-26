package orderbook

import (
	"fmt"
	"math/rand"
	"sort"
	"time"
)

type Match struct {
	Ask        *Order
	Bid        *Order
	SizeFilled float64
	Price      float64
}

type Order struct {
	ID        int64
	UserID    int64
	Size      float64
	Bid       bool
	Limit     *Limit
	Timestamp int64
}
type Orders []*Order

func (o Orders) Len() int           { return len(o) }
func (o Orders) Swap(i, j int)      { o[i], o[j] = o[j], o[i] }
func (o Orders) Less(i, j int) bool { return o[i].Timestamp < o[j].Timestamp }

func NewOrder(bid bool, size float64, userID int64) *Order {
	return &Order{
		UserID:    userID,
		ID:        int64(rand.Intn(1000000000000)),
		Size:      size,
		Bid:       bid,
		Timestamp: time.Now().UnixNano(),
	}

}
func (o *Order) String() string {
	return fmt.Sprintf("[size: %.2f]", o.Size)

}

func (o *Order) isFilled() bool {
	return o.Size == 0.0
}

type Limit struct {
	Price       float64
	Orders      Orders
	TotalVolume float64
}

type Limits []*Limit
type ByBestAsk struct{ Limits }

func (a ByBestAsk) Len() int           { return len(a.Limits) }
func (a ByBestAsk) Swap(i, j int)      { a.Limits[i], a.Limits[j] = a.Limits[j], a.Limits[i] }
func (a ByBestAsk) Less(i, j int) bool { return a.Limits[i].Price < a.Limits[j].Price }

type ByBestBid struct{ Limits }

func (b ByBestBid) Len() int           { return len(b.Limits) }
func (b ByBestBid) Swap(i, j int)      { b.Limits[i], b.Limits[j] = b.Limits[j], b.Limits[i] }
func (b ByBestBid) Less(i, j int) bool { return b.Limits[i].Price < b.Limits[j].Price }

func NewLimit(price float64) *Limit {
	return &Limit{
		Price:  price,
		Orders: []*Order{},
	}
}

func (l *Limit) String() string {
	return fmt.Sprintf("[price: %.2f | volume: %.2f]", l.Price, l.TotalVolume)
}

func (l *Limit) AddOrder(o *Order) {
	o.Limit = l
	l.Orders = append(l.Orders, o)
	l.TotalVolume += o.Size

	sort.Sort(Orders{})

}
func (l *Limit) DeleteOrder(o *Order) {
	for i := 0; i < len(l.Orders); i++ {
		if l.Orders[i] == o {
			l.Orders[i] = l.Orders[len(l.Orders)-1]
			l.Orders = l.Orders[:len(l.Orders)-1]
		}
	}
	o.Limit = nil
	l.TotalVolume -= o.Size
}

func (l *Limit) Fill(o *Order) []Match {
	var (
		matches        []Match
		ordersToDelete []*Order
	)

	for _, order := range l.Orders {
		match := l.fillOrder(order, o)
		matches = append(matches, match)

		l.TotalVolume -= match.SizeFilled

		if order.isFilled() {
			ordersToDelete = append(ordersToDelete, order)
		}

		if o.isFilled() {

			break
		}
	}
	for _, order := range ordersToDelete {
		l.DeleteOrder(order)
	}
	return matches
}

func (l *Limit) fillOrder(a, b *Order) Match {
	var (
		bid        *Order
		ask        *Order
		sizeFilled float64
	)
	if a.Bid {
		bid = a
		ask = b
	} else {
		bid = b
		ask = a
	}
	if a.Size >= b.Size {
		a.Size -= b.Size
		sizeFilled = b.Size
		b.Size = 0

	} else {
		b.Size -= a.Size
		sizeFilled = a.Size
		a.Size = 0.0
	}
	return Match{
		Bid:        bid,
		Ask:        ask,
		SizeFilled: sizeFilled,
		Price:      l.Price,
	}
}

type OrderBook struct {
	ask []*Limit
	bid []*Limit

	AskLimits map[float64]*Limit
	BidLimits map[float64]*Limit
	Orders    map[int64]*Order
}

func NewOrderBook() *OrderBook {
	return &OrderBook{
		ask:       []*Limit{},
		bid:       []*Limit{},
		AskLimits: make(map[float64]*Limit),
		BidLimits: make(map[float64]*Limit),
		Orders:    make(map[int64]*Order),
	}
}
func (ab *OrderBook) Placeholder(price float64, o *Order) []Match {
	if o.Size > 0.0 {
		ab.add(price, o)

	}
	return []Match{}
}

func (ab *OrderBook) add(price float64, o *Order) {

}

func (ab *OrderBook) PlaceLimitOrder(price float64, o *Order) {
	var limit *Limit

	if o.Bid {
		limit = ab.BidLimits[price]
	} else {
		limit = ab.AskLimits[price]
	}

	if limit == nil {
		limit = NewLimit(price)
		if o.Bid {
			ab.bid = append(ab.bid, limit)
			ab.BidLimits[price] = limit
		} else {
			ab.ask = append(ab.ask, limit)
			ab.BidLimits[price] = limit
		}
	} else {
		ab.ask = append(ab.ask, limit)
		ab.AskLimits[price] = limit
	}
	ab.Orders[o.ID] = o
	limit.AddOrder(o)
}

func (ab *OrderBook) Asks() []*Limit {
	sort.Sort(ByBestAsk{ab.ask})
	return ab.ask

}

func (ab *OrderBook) Bids() []*Limit {
	sort.Sort(ByBestBid{ab.bid})
	return ab.bid

}

func (ob *OrderBook) PlaceMarketOrder(o *Order) []Match {
	matches := []Match{}

	if o.Bid {
		if o.Size > ob.AskTotalVolume() {
			panic(fmt.Errorf("not enough volume[size: %.2f] for market order[%.2f]", ob.AskTotalVolume(), o.Size))
		}
		for _, limit := range ob.Asks() {
			limitMatches := limit.Fill(o)
			matches = append(matches, limitMatches...)
			if len(limit.Orders) == 0 {
				ob.clearLimit(true, limit)
			}
		}
	} else {
		if o.Size > ob.BidTotalVolume() {
			panic(fmt.Errorf("not enough volume[size: %.2f] for market order[%.2f]", ob.BidTotalVolume(), o.Size))
		}
		for _, limit := range ob.Bids() {
			limitMatches := limit.Fill(o)
			matches = append(matches, limitMatches...)

			if len(limit.Orders) == 0 {
				ob.clearLimit(true, limit)
			}

		}
	}

	return matches
}

func (ob *OrderBook) BidTotalVolume() float64 {
	totalVolume := 0.0

	for i := 0; i < len(ob.bid); i++ {
		totalVolume += ob.bid[i].TotalVolume
	}

	return totalVolume
}

func (ob *OrderBook) AskTotalVolume() float64 {
	totalVolume := 0.0

	for i := 0; i < len(ob.ask); i++ {
		totalVolume += ob.ask[i].TotalVolume
	}

	return totalVolume
}

func (ob *OrderBook) clearLimit(bid bool, l *Limit) {
	if bid {
		delete(ob.BidLimits, l.Price)
		for i := 0; i < len(ob.bid); i++ {
			if ob.bid[i] == l {
				ob.bid[i] = ob.bid[len(ob.bid)-1]
				ob.bid = ob.bid[:len(ob.bid)-1]
			}
		}
	} else {
		if bid {
			delete(ob.AskLimits, l.Price)
			for i := 0; i < len(ob.ask); i++ {
				if ob.ask[i] == l {
					ob.ask[i] = ob.ask[len(ob.ask)-1]
					ob.ask = ob.ask[:len(ob.ask)-1]
				}

			}
		}
	}
}

func (ob *OrderBook) CancelOrder(o *Order) {
	limit := o.Limit
	limit.DeleteOrder(o)
	delete(ob.Orders, o.ID)

}
