// Copyright (c) 2022 - for information on the respective copyright owner
// see the NOTICE file and/or the repository at
// https://github.com/hyperledger-labs/perun-node
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package grpc

import (
	"fmt"

	"github.com/hyperledger-labs/perun-node/currency"
)

var curr = currency.ETHSymbol

var D Dashboard

type Dashboard struct {
	length                int
	leftAlignLeadingSpace int
}

func InitDashboard() Dashboard {

	d := Dashboard{
		length:                134,
		leftAlignLeadingSpace: 2,
	}

	// d.PrintTestChars()

	d.PrintDashes()
	d.PrintStringCenterAligned("Car owner's perun dashboard")
	d.PrintDashes()
	return d
}

func (d Dashboard) PrintTestChars() {
	fmt.Printf(".")
	for i := 0; i < d.length-2; i++ {
		fmt.Printf(" ")
	}
	fmt.Printf(".")
	fmt.Printf("\n")
}

func (d Dashboard) PrintDashes() {
	for i := 0; i < d.length; i++ {
		fmt.Printf("―")
	}
	fmt.Printf("\n")
}

func (d Dashboard) PrintBlank() {
	for i := 0; i < d.length; i++ {
		fmt.Printf(" ")
	}
	fmt.Printf("\n")
}

func (d Dashboard) PrintStringCenterAligned(s string) {
	cntBlanks := d.length - len(s)

	for i := 0; i < cntBlanks/2; i++ {
		fmt.Printf(" ")
	}
	fmt.Printf(s)
	// If count of blanks has odd numbered value, print a extra space in right half.
	if cntBlanks%2 == 1 {
		fmt.Printf(" ")
	}
	for i := 0; i < cntBlanks/2; i++ {
		fmt.Printf(" ")
	}
	fmt.Printf("\n")
}

func (d Dashboard) PrintStringLeftAligned(s string) {
	cntBlanks := d.length - len(s)

	for i := 0; i < d.leftAlignLeadingSpace; i++ {
		fmt.Printf(" ")
	}

	fmt.Printf(s)

	for i := 0; i < cntBlanks-d.leftAlignLeadingSpace; i++ {
		fmt.Printf(" ")
	}
	fmt.Printf("\n")
}

func (d Dashboard) SessionOpened() {
	d.PrintStringLeftAligned(fmt.Sprintf("Car connected. Services included: funding, watching"))
}
func (d Dashboard) FundingRequest(parts, balances []string) {
	d.PrintStringLeftAligned(fmt.Sprintf("Funding request: %v %v %s", parts, balances, curr))
}
func (d Dashboard) FundingSuccessful() {
	d.PrintStringLeftAligned(fmt.Sprintf("Funding successful"))
}

func (d Dashboard) WatchingRequest() {
	d.PrintStringLeftAligned(fmt.Sprintf("Watching request for channel with charger"))
}
func (d Dashboard) WatchingSuccessful() {
	d.PrintStringLeftAligned(fmt.Sprintf("Started watching"))
}

func (d Dashboard) ChannelUpdated(parts, balances []string) {
	d.PrintStringLeftAligned(fmt.Sprintf("Channel balance updated: %v %v %s", parts, balances, curr))
}

func (d Dashboard) ChannelFinalized(parts, balances []string) {
	d.PrintStringLeftAligned(fmt.Sprintf("Channel finalized off the chain: %v %v %s", parts, balances, curr))
}

func (d Dashboard) ChannelRegistered(parts, balances []string) {
	d.PrintStringLeftAligned(fmt.Sprintf("Channel registered on the chain: %v %v %s", parts, balances, curr))
}

func (d Dashboard) ChannelConcluded() {
	d.PrintStringLeftAligned(fmt.Sprintf("Channel concluded on the chain"))
}

func (d Dashboard) WithdrawRequest(parts, balances []string) {
	d.PrintStringLeftAligned(fmt.Sprintf("Withdraw request: %v %v %s", parts, balances, curr))
}
func (d Dashboard) WithdrawSuccessful() {
	d.PrintStringLeftAligned(fmt.Sprintf("Withdraw successful"))
}
