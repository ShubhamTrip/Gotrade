package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/adshao/go-binance/v2"
)

var (
	apiKey    = os.Getenv("BINANCE_API_KEY")
	apiSecret = os.Getenv("BINANCE_API_SECRET")
	symbol    = "BTCUSDT"
	spread    = 20.0    // USD spread
	orderQty  = "0.001" // BTC amount
	sleepTime = 30 * time.Second
)

func getMidPrice(client *binance.Client) (float64, error) {
	price, err := client.NewAveragePriceService().Symbol(symbol).Do(context.Background())
	if err != nil {
		return 0, err
	}
	return strconv.ParseFloat(price.Price, 64)
}

func placeLimitOrders(client *binance.Client, midPrice float64) ([]int64, error) {
	bid := fmt.Sprintf("%.2f", midPrice-spread/2)
	ask := fmt.Sprintf("%.2f", midPrice+spread/2)

	log.Printf("Placing Buy @ %s, Sell @ %s\n", bid, ask)

	buyResp, err := client.NewCreateOrderService().
		Symbol(symbol).
		Side(binance.SideTypeBuy).
		Type(binance.OrderTypeLimit).
		TimeInForce(binance.TimeInForceTypeGTC).
		Quantity(orderQty).
		Price(bid).
		Do(context.Background())
	if err != nil {
		return nil, err
	}

	sellResp, err := client.NewCreateOrderService().
		Symbol(symbol).
		Side(binance.SideTypeSell).
		Type(binance.OrderTypeLimit).
		TimeInForce(binance.TimeInForceTypeGTC).
		Quantity(orderQty).
		Price(ask).
		Do(context.Background())
	if err != nil {
		return nil, err
	}

	return []int64{buyResp.OrderID, sellResp.OrderID}, nil
}

func cancelOrders(client *binance.Client, orderIDs []int64) {
	for _, orderID := range orderIDs {
		_, err := client.NewCancelOrderService().
			Symbol(symbol).
			OrderID(orderID).
			Do(context.Background())
		if err != nil {
			log.Printf("Error cancelling order %d: %v", orderID, err)
		} else {
			log.Printf("Cancelled order %d", orderID)
		}
	}
}

func main() {

	client := binance.NewClient(apiKey, apiSecret)
	client.BaseURL = "https://testnet.binance.vision"

	for {
		mid, err := getMidPrice(client)
		if err != nil {
			log.Printf("Failed to fetch price: %v", err)
			time.Sleep(5 * time.Second)
			continue
		}

		orderIDs, err := placeLimitOrders(client, mid)
		if err != nil {
			log.Printf("Failed to place orders: %v", err)
			time.Sleep(5 * time.Second)
			continue
		}

		time.Sleep(sleepTime)
		cancelOrders(client, orderIDs)
	}
}
