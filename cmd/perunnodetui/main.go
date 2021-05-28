// Copyright (c) 2021 - for information on the respective copyright owner
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

package main

import (
	"flag"
	"fmt"
	"time"

	"github.com/mum4k/termdash"
	"github.com/mum4k/termdash/cell"
	"github.com/mum4k/termdash/container"
	"github.com/mum4k/termdash/keyboard"
	"github.com/mum4k/termdash/terminal/tcell"
	"github.com/mum4k/termdash/terminal/terminalapi"
	"github.com/mum4k/termdash/widgets/text"
	"github.com/pkg/errors"

	pwallet "perun.network/go-perun/wallet"

	"github.com/hyperledger-labs/perun-node/api/grpc/pb"
	"github.com/hyperledger-labs/perun-node/blockchain/ethereum"
	"github.com/hyperledger-labs/perun-node/blockchain/ethereum/ethereumtest"
	"github.com/hyperledger-labs/perun-node/log"
)

const (
	terminalMinX = 106
	terminalMinY = 29

	logFile  = "perunnodetui.log"
	logLevel = "debug"

	challengeDurSecs uint64 = 10 // standard value all outgoing channel open requests.

	rootContainerID = "root"
)

var (
	// Data for an instance of client. Once initialized, these will only be read.
	sessionID   string
	userName    string
	onChainAddr pwallet.Address
	chainURL    string

	// Size of the terminal.
	x, y int

	// Properties of the log container used across both screens.
	logsHeight        = 8
	logBoxBorderColor = cell.Color(1)
	logBoxTitleColor  = cell.Color(6)

	// Singleton instances set during init and used across the program.
	// Safe for concurrent use.
	client pb.Payment_APIClient  // Connection to the perun node.
	logBox *text.Text            // Text box for logging. Common across two screen.
	logger log.Logger            // Logger for logging to file.
	errs   = make(chan error, 5) // Channel for functions to send errors to the main event loop.

)

func main() {
	var err error
	var isDeployRequested bool
	isDeployRequested, userName = parseFlags()

	deployContractsIfRequested(isDeployRequested)

	initLoggers(logLevel, logFile)

	// Initialize termdash container.
	t, err := tcell.New()
	if err != nil {
		logger.Fatal(errors.Wrap(err, "initializing tcell"))
	}
	defer t.Close()
	// After this point, call only return (not panic/os.Exit),
	// so that t.Close is invoked to return the terminal to sane state.

	if x, y = t.Size().X, t.Size().Y; x < terminalMinX || y < terminalMinY {
		logErrorf("Terminal window size is %d x %d. Should be at least %d x %d", x, y, terminalMinX, terminalMinY)
		return
	}
	logInfof("Terminal window size: %d x %d", x, y)
	c, err := container.New(t, container.ID(rootContainerID))
	if err != nil {
		logError(errors.Wrap(err, "initializing container"))
		return
	}

	// Construct event loop.
	quitter := make(chan bool) // For sending close signal to all any go routine.
	keyboardEvents := make(chan *terminalapi.Keyboard, 5)
	keyboardEventsHandler := func(k *terminalapi.Keyboard) { keyboardEvents <- k }
	errHandler := func(e error) { errs <- e }
	eventLoop := contructEventLoop(keyboardEvents, errs, quitter)

	// Initialize and render connect screen.
	connectScreen, err := newConnectScreen(userName)
	if err != nil {
		logError(errors.Wrap(err, "initializing connect screen"))
		return
	}
	connectScreen.connectBtn.SetCallback(connectScreen.getConnectFn(c, quitter))
	connectScreen.quitBtn.SetCallback(connectScreen.getQuitFn(quitter))
	if err = renderConnectScreen(c, connectScreen); err != nil {
		panic(err)
	}

	// renderDummyDashboard(c) // For testing dashboard view. This should normally be commented out.

	// Run termdash controller.
	controller, err := termdash.NewController(t, c,
		termdash.KeyboardSubscriber(keyboardEventsHandler), termdash.ErrorHandler(errHandler))
	if err != nil {
		logError(errors.Wrap(err, "initializing controller"))
		return
	}
	defer controller.Close()

	eventLoop()
}

func parseFlags() (bool, string) {
	aliceFlag := flag.Bool("alice", false, "load alice defaults")
	bobFlag := flag.Bool("bob", false, "load bob defaults")
	deployFlag := flag.Bool("deploy", false, "deploy contracts on blockchain")

	flag.Parse()
	user := ""
	switch {
	case *aliceFlag && !*bobFlag:
		user = alice
	case !*aliceFlag && *bobFlag:
		user = bob
	}
	return *deployFlag, user
}

func deployContractsIfRequested(requested bool) {
	if requested {
		_, _, err := ethereumtest.SetupContracts(defaultChainURL, ethereumtest.ChainID, ethereumtest.OnChainTxTimeout)
		if err != nil {
			panic(err)
		}
	}
}

func initLoggers(logLevel, logFile string) {
	err := log.InitLogger(logLevel, logFile)
	if err != nil {
		panic(err)
	}
	logger = log.NewLogger()

	logBox, err = text.New(text.RollContent(), text.WrapAtWords())
	if err != nil {
		panic(err)
	}
}

func contructEventLoop(keyboardEvents chan *terminalapi.Keyboard, errs chan error, quitter chan bool) func() {
	return func() {
		for {
			select {
			case e := <-errs:
				if e != nil {
					logError(e)
				}

			case k := <-keyboardEvents:
				if k.Key == keyboard.KeyEsc || k.Key == keyboard.KeyCtrlC {
					logErrorf("Received %s. Closing the client.", k.Key)
					close(quitter) // signal other go-routines to shut down.
					time.Sleep(1 * time.Second)
					return
				}

			case <-quitter:
				logError("User pressed quit button. Closing the client.")
				close(quitter) // signal other go-routines to shut down.
				time.Sleep(1 * time.Second)
				return
			}
		}
	}
}

// Handy functions for logging to both file and log box.
// These assume the logger and logBox instances are initialized.

func logErrorf(format string, a ...interface{}) {
	msg := fmt.Sprintf(format, a...)
	logger.Error(msg)
	logBox.Write("Error: "+msg+"\n", text.WriteCellOpts(cell.FgColor(cell.ColorRed))) // nolint: errcheck, gosec
}

func logInfof(format string, a ...interface{}) {
	msg := fmt.Sprintf(format, a...)
	logger.Info(msg)
	logBox.Write(msg + "\n") // nolint: errcheck, gosec
}

func logError(a ...interface{}) {
	msg := fmt.Sprint(a...)
	logger.Error(msg)
	logBox.Write("Error: "+msg+"\n", text.WriteCellOpts(cell.FgColor(cell.ColorRed))) // nolint: errcheck, gosec
}

func logInfo(a ...interface{}) {
	msg := fmt.Sprint(a...)
	logger.Error(msg)
	logBox.Write(msg + "\n") // nolint: errcheck, gosec
}

// renderDummyDashboard is used to fill dummy data in the table for testing.
// call this immediately after renderDashboard in main.
func renderDummyDashboard(c *container.Container) { // nolint: unused,deadcode
	var err error
	chainURL = defaultChainURL
	onChainAddr, err = ethereum.NewWalletBackend().ParseAddr(defaultOnChainAddrs[alice])
	if err != nil {
		panic(err)
	}

	quitter := make(chan bool)
	d, err := newDashboardScreen()
	if err != nil {
		panic(err)
	}
	if err := renderDashboardView(c, d); err != nil {
		fmt.Printf("%+v", err)
		panic(err)
	}
	go updateOnChainBalNTime(d.onChainBalText, d.timeText, quitter)

	dummyBalInfo := balInfo{
		ours:    "1.000000",
		theirs:  "2.6700000",
		version: "1",
	}
	p1, _ := newIncomingChannel("", "peer", "", "", time.Now().Format("15:04:05")) // nolint: errcheck
	p1.status = waitingForPeer
	p2, _ := newIncomingChannel("", "peer", "", "", time.Now().Format("15:04:05")) // nolint: errcheck
	p2.status = accepted
	p2.phase = transact
	p3, _ := newIncomingChannel("", "peer", "", "", time.Now().Format("15:04:05")) // nolint: errcheck
	p3.status = responding
	p3.phase = register
	p4, _ := newIncomingChannel("", "peer", "", "", time.Now().Format("15:04:05")) // nolint: errcheck
	p4.status = peerRejected
	p4.phase = settle
	p5, _ := newIncomingChannel("", "peer", "", "", time.Now().Format("15:04:05")) // nolint: errcheck
	p5.phase = closed
	p6, _ := newIncomingChannel("", "peer", "", "", time.Now().Format("15:04:05")) // nolint: errcheck
	p6.phase = errorPhase
	p1.current, p1.proposed = dummyBalInfo, dummyBalInfo
	p2.current, p2.proposed = dummyBalInfo, dummyBalInfo
	p3.current, p3.proposed = dummyBalInfo, dummyBalInfo
	p3.current.theirs = "2.032"

	p1.refreshEntry()
	p2.refreshEntry()
	p3.refreshEntry()
	p4.refreshEntry()
	p5.refreshEntry()
	p6.refreshEntry()
}
