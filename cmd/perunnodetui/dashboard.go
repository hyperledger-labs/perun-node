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
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/mum4k/termdash/align"
	"github.com/mum4k/termdash/cell"
	"github.com/mum4k/termdash/container"
	"github.com/mum4k/termdash/container/grid"
	"github.com/mum4k/termdash/linestyle"
	"github.com/mum4k/termdash/widgets/text"
	"github.com/mum4k/termdash/widgets/textinput"
	"github.com/pkg/errors"
	pwallet "perun.network/go-perun/wallet"

	"github.com/hyperledger-labs/perun-node/blockchain/ethereum"
	"github.com/hyperledger-labs/perun-node/blockchain/ethereum/ethereumtest"
	"github.com/hyperledger-labs/perun-node/currency"
)

var (
	// T is the package level instance of table. Used for getting new rows for
	// new incoming or outgoing channels.
	T *table

	// R is the package level instance of registry. It is used to keep track of
	// mapping betweeen channel IDs and row numbers in the table.
	R *registry
)

var (

	// Address: 0x8450c0055cB180C7C37A25866132A740b812937B      Balance: 69.936479 ETH    |    00:09:32    |.

	// Colors and templates used in dashboard screen.
	onChainBalTemplate = []string{"\nAddress: %-50sBalance:", " %s ETH "}
	timeTemplate       = "\n%13s"
	addressColor       = cell.Color(7)
	borderColor        = cell.Color(1)
	borderTitleColor   = cell.Color(5)
	placeHolderColor   = cell.Color(8)
	commandTextColor   = cell.Color(1)
	balanceColor       = cell.Color(7)

	// Table title color is different, because it includes dashes for the full
	// length of the table to position the words channels and updates.
	tableTitleColor = cell.Color(1)

	// Options used for highlight the ch status field when it is waiting for user response.
	forUserWriteOpts = text.WriteCellOpts(cell.Blink(), cell.FgColor(cell.Color(4)))

	onChainBalNTimeUpdateInterval = time.Second * 1
)

type dashboardScreen struct {
	elementHeight    int
	onChainBalHeight int
	onChainBalWidth  int
	timeWidth        int
	tableHeight      int
	commandHeight    int

	onChainBalText *text.Text
	timeText       *text.Text
	commandTBox    *textinput.TextInput

	table *table
}

// table to display channels in the dashboard.
type table struct {
	sync.Mutex
	rows        []*text.Text
	rowElements []grid.Element
	width       int

	headerRow    string
	bodyTemplate []string

	offset       int
	currentIndex int
	lastIndex    int
}

// Computations for table: Column width calculation
// Title 1:						       Current Balance									Updated Balance
// Title 2:		S.No	Phase	Peer	Self	Peer's	Ver		‖   Status	Timeout		Self	Peer's	Ver
// MaxLen:		4		8		6		4		6		3		1	8		8			4		6		3
// W 4 Space:	2+6		12		10		8		10		5		3	12		12			8		10		5
// Channels = 53	Update = 50		Total = 103

func newTable(h int) (*table, error) {
	cntRows := h
	rows := make([]*text.Text, cntRows)
	rowElements := make([]grid.Element, cntRows)

	channelsLength := 53
	updatesLength := 50
	tableWidth := channelsLength + updatesLength
	title1 := "                             Current Balance                                   Proposed Balance       "
	title2Template := "  %-6s%-12s%-10s%-8s%-10s%-5s│  %-12s%-12s%-8s%-10s%-5s"                 // "S.No" is also text.
	bodyTemplate := []string{"  %-6d%-12s%-10s%-8s%-10s%-5s│  ", "%-12s", "%-12s%-8s%-10s%-5s"} // S.No is integer.

	// Sections are 53, 50 chars long. Channels is 8 chars, Updates is 7 chars.
	chDashes := strings.Repeat("―", 21)
	updateDashes := strings.Repeat("―", 23)
	headerRow := chDashes + "Channels―" + chDashes + updateDashes + "Updates" + updateDashes

	// Fill in title row and empty data for rest of the rows.
	var err error
	for i := range rows {
		if rows[i], err = text.New(text.DisableScrolling()); err != nil {
			return nil, errors.Wrap(err, fmt.Sprintf("initializing row #%2d", i+1))
		}
		rowElements[i] = grid.Widget(rows[i])
	}
	noData := strings.Repeat(" ", channelsLength) + "│" + strings.Repeat(" ", updatesLength-1)
	for i := 0; i < cntRows; i++ {
		if err = rows[i].Write(noData); err != nil {
			return nil, errors.Wrap(err, fmt.Sprintf("writing row #%2d", i+1))
		}
	}
	rows[0].Reset()
	err = rows[0].Write(title1, text.WriteCellOpts(cell.Bold(), cell.FgColor(borderTitleColor)))
	if err != nil {
		return nil, errors.Wrap(err, "writing title row 1")
	}
	title := fmt.Sprintf(title2Template,
		"S.No", "Phase", "Peer", "Self", "Peer's", "Ver", "Status", "Timeout", "Self", "Peer's", "Ver")
	rows[1].Reset()
	err = rows[1].Write(title, text.WriteCellOpts(cell.Bold(), cell.FgColor(borderTitleColor)))
	if err != nil {
		return nil, errors.Wrap(err, "writing title row 2")
	}

	return &table{
		rows:        rows,
		rowElements: rowElements,
		width:       tableWidth,

		bodyTemplate: bodyTemplate,
		headerRow:    headerRow,

		offset:       3,
		currentIndex: 1,
		lastIndex:    cntRows - 3, // Two title rows and leave last row empty.
	}, nil
}

func (t *table) getRow() (int, *text.Text, bool) {
	t.Lock()
	defer t.Unlock()
	if t.currentIndex == t.lastIndex {
		return 0, nil, false
	}

	row := t.rows[t.currentIndex-1+t.offset]
	index := t.currentIndex

	t.currentIndex++
	return index, row, true
}

func newDashboardScreen() (*dashboardScreen, error) {
	elementHeight := 2
	onChainBalHeight := elementHeight * 2
	commandTBoxHeight := elementHeight * 2
	tableHeight := y - (onChainBalHeight + commandTBoxHeight + logsHeight)
	timeWidth := 18 // Computed for table width of 103.

	table, err := newTable(tableHeight)
	if err != nil {
		return nil, errors.Wrap(err, "initializing table")
	}
	T = table
	R = newRegistry(T.lastIndex - T.offset)

	onChainBalWidth := T.width - timeWidth
	onChainBalText, err := text.New(text.DisableScrolling())
	if err != nil {
		return nil, errors.Wrap(err, "initializing on-chain balance text")
	}
	timeText, err := text.New(text.DisableScrolling())
	if err != nil {
		return nil, errors.Wrap(err, "initializing time text")
	}
	commandTBox, err := textinput.New(
		textinput.PlaceHolder(commandHelp), textinput.PlaceHolderColor(placeHolderColor),
		textinput.FillColor(cell.Color(0)), textinput.TextColor(commandTextColor),
		textinput.ExclusiveKeyboardOnFocus(),
		textinput.ClearOnSubmit(), textinput.OnSubmit(func(input string) error {
			logInfof("Received command input: " + input)
			return errors.WithMessage(handleCmd(input), "processing command")
		}))
	if err != nil {
		return nil, errors.Wrap(err, "initializing commands text")
	}

	return &dashboardScreen{
		elementHeight:    elementHeight,
		onChainBalHeight: onChainBalHeight,
		onChainBalWidth:  onChainBalWidth,
		timeWidth:        timeWidth,
		tableHeight:      tableHeight,
		commandHeight:    commandTBoxHeight,

		onChainBalText: onChainBalText,
		timeText:       timeText,
		commandTBox:    commandTBox,
		table:          table,
	}, nil
}

func renderDashboardView(c *container.Container, d *dashboardScreen) error {
	builder := grid.New()

	tableRows := make([]grid.Element, d.table.lastIndex+1)
	for i := range tableRows {
		tableRows[i] = grid.RowHeightFixed(1, d.table.rowElements[i])
	}

	builder.Add(
		grid.RowHeightFixed(d.onChainBalHeight,
			grid.ColWidthFixedWithOpts(d.onChainBalWidth,
				[]container.Option{
					container.BorderTitle(userName),
					container.TitleColor(addressColor), container.TitleFocusedColor(addressColor),
					container.BorderTitleAlignCenter(),
					container.Border(linestyle.Round),
					container.FocusedColor(borderColor),
				}, grid.ColWidthFixedWithOpts(d.onChainBalWidth, horizontalLeft,
					grid.Widget(d.onChainBalText)),
			),
			grid.ColWidthFixedWithOpts(d.timeWidth,
				[]container.Option{
					container.BorderTitle("Time"),
					container.TitleColor(borderTitleColor), container.TitleFocusedColor(borderTitleColor),
					container.BorderTitleAlignCenter(),
					container.Border(linestyle.Round),
					container.FocusedColor(borderColor),
				}, grid.ColWidthFixedWithOpts(d.timeWidth, horizontalCenter, grid.Widget(d.timeText)),
			),
		),
		grid.RowHeightFixedWithOpts(d.tableHeight,
			[]container.Option{
				container.BorderTitle(d.table.headerRow),
				container.TitleColor(tableTitleColor), container.TitleFocusedColor(tableTitleColor),
				container.Border(linestyle.Round),
				container.FocusedColor(borderColor),
			}, tableRows...),
		grid.RowHeightFixedWithOpts(d.commandHeight,
			[]container.Option{
				container.BorderTitle("Command Box"),
				container.TitleColor(borderTitleColor), container.TitleFocusedColor(borderTitleColor),
				container.Border(linestyle.Light),
				container.FocusedColor(borderColor),
			}, grid.Widget(d.commandTBox)),
		grid.RowHeightFixedWithOpts(logsHeight,
			[]container.Option{
				container.BorderTitle("Logs"),
				container.TitleColor(logBoxTitleColor), container.TitleFocusedColor(logBoxTitleColor),
				container.Border(linestyle.Round),
				container.AlignVertical(align.VerticalTop), container.AlignHorizontal(align.HorizontalLeft),
				container.FocusedColor(logBoxBorderColor),
			}, grid.Widget(logBox),
		),
	)
	gridOpts, err := builder.Build()
	if err != nil {
		return errors.Wrap(err, "building dashboard screen")
	}
	return errors.Wrap(c.Update(rootContainerID, gridOpts...), "updating dashboard screen")
}

// updateOnChainBalNTime updates the time and balance in the text boxes.
// should be invoked as a go-routine.
// the function logs the errors to error box and returns when quitter channel is closed.
//
// on-chain addresse and chain URL are taken from the package level variables.
func updateOnChainBalNTime(onChainBalText, timeText *text.Text, quitter chan bool) {
	onChainBal1WriteOpts := text.WriteCellOpts(cell.FgColor(balanceColor))
	handleError := func(err error) {
		if err != nil {
			logError(err)
		}
	}

	ticker := time.NewTicker(onChainBalNTimeUpdateInterval)
	for {
		select {
		case t := <-ticker.C:
			onChainBal, err := readOnChainBal(chainURL, onChainAddr)
			handleError(errors.WithMessage(err, "reading on-chain balance"))

			onChainBalText.Reset()
			err = onChainBalText.Write(fmt.Sprintf(onChainBalTemplate[0], onChainAddr.String()))
			handleError(errors.WithMessage(err, "writing on-chain balance part 1"))

			if onChainBal != "" {
				err = onChainBalText.Write(fmt.Sprintf(onChainBalTemplate[1], onChainBal), onChainBal1WriteOpts)
				handleError(errors.WithMessage(err, "writing on-chain balance part 2"))
			}

			timeText.Reset()
			err = timeText.Write(fmt.Sprintf(timeTemplate, t.Format("15:04:05")))
			handleError(errors.WithMessage(err, "writing time"))

		case <-quitter:
			return
		}
	}
}

func readOnChainBal(chainURL string, onChainAddr pwallet.Address) (string, error) {
	onChainBalVal, err := ethereum.BalanceAt(chainURL,
		ethereumtest.ChainConnTimeout, ethereumtest.OnChainTxTimeout, onChainAddr)
	if err != nil {
		return "", errors.WithMessage(err, "reading on-chain balance")
	}
	return currency.NewParser(currency.ETH).Print(onChainBalVal), nil
}
