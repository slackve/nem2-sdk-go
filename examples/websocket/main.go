// Copyright 2018 ProximaX Limited. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package main

import (
	"context"
	"fmt"
	"github.com/proximax-storage/nem2-sdk-go/sdk"
	"time"
)

const (
	baseUrl     = "http://localhost:3000"
	networkType = sdk.MijinTest
	privateKey  = "0F3CC33190A49ABB32E7172E348EA927F975F8829107AAA3D6349BB10797D4F6"
)

// WebSockets make possible receiving notifications when a transaction or event occurs in the blockchain.
// The notification is received in real time without having to poll the API waiting for a reply.
func main() {

	conf, err := sdk.LoadTestnetConfig(baseUrl)
	if err != nil {
		panic(err)
	}

	acc, err := sdk.NewAccount(privateKey, networkType)

	ws, err := sdk.NewConnectWs(baseUrl)
	if err != nil {
		panic(err)
	}

	fmt.Println("websocket negotiated uid:", ws.Uid)

	// The UnconfirmedAdded channel notifies when a transaction related to an
	// address is in unconfirmed state and waiting to be included in a block.
	// The message contains the transaction.
	a, _ := ws.Subscribe.UnconfirmedAdded(acc.Address.Address)
	go func() {
		for {
			data := <-a.ChIn
			ch := data.(sdk.Transaction)
			fmt.Printf("UnconfirmedAdded Tx Hash: %v \n", ch.GetAbstractTransaction().Hash)
			a.Unsubscribe()
		}
	}()

	// The confirmedAdded channel notifies when a transaction related to an
	// address is included in a block. The message contains the transaction.
	b, _ := ws.Subscribe.ConfirmedAdded(acc.Address.Address)
	go func() {
		for {
			data := <-b.ChIn
			ch := data.(sdk.Transaction)
			fmt.Printf("ConfirmedAdded Tx Hash: %v \n", ch.GetAbstractTransaction().Hash)
			b.Unsubscribe()
			fmt.Println("Successful transfer!")

		}
	}()

	//The status channel notifies when a transaction related to an address rises an error.
	//The message contains the error message and the transaction hash.
	c, _ := ws.Subscribe.Status(acc.Address.Address)

	go func() {
		for {
			data := <-c.ChIn
			ch := data.(sdk.StatusInfo)
			c.Unsubscribe()
			fmt.Printf("Hash: %v \n", ch.Hash)
			panic(fmt.Sprint("Status: ", ch.Status))
		}
	}()

	time.Sleep(time.Second * 5)
	// Use the default http client
	client := sdk.NewClient(nil, conf)

	ttx, err := sdk.NewTransferTransaction(
		sdk.NewDeadline(time.Hour*1),
		sdk.NewAddress("SBILTA367K2LX2FEXG5TFWAS7GEFYAGY7QLFBYKC", networkType),
		sdk.Mosaics{sdk.Xem(10000000)},
		sdk.NewPlainMessage(""),
		networkType,
	)

	stx, err := acc.Sign(ttx)
	if err != nil {
		panic(fmt.Errorf("TransaferTransaction signing returned error: %s", err))
	}

	// Get the chain height
	restTx, resp, err := client.Transaction.Announce(context.Background(), stx)
	if err != nil {
		panic(err)
	}
	fmt.Printf("%s\n", restTx)
	fmt.Printf("Response Status Code == %d\n\n", resp.StatusCode)
	fmt.Printf("Hash: \t\t%v\n", stx.Hash)
	fmt.Printf("Signer: \t%X\n\n", acc.KeyPair.PublicKey.Raw)

	// The block channel notifies for every new block.
	// The message contains the block information.
	d, _ := ws.Subscribe.Block()

	for {
		data := <-d.ChIn
		ch := data.(*sdk.BlockInfo)
		fmt.Printf("Block received with height: %v \n", ch.Height)
	}
}
