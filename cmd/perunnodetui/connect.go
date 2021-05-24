package main

import (
	"github.com/mum4k/termdash/align"
	"github.com/mum4k/termdash/cell"
	"github.com/mum4k/termdash/container"
	"github.com/mum4k/termdash/container/grid"
	"github.com/mum4k/termdash/linestyle"
	"github.com/mum4k/termdash/widgets/button"
	"github.com/mum4k/termdash/widgets/textinput"
	"github.com/pkg/errors"

	"github.com/hyperledger-labs/perun-node/blockchain/ethereum"
)

var (
	// Two default roles and their corresponding data.
	alice               = "Alice"
	bob                 = "Bob"
	defaultOnChainAddrs = map[string]string{
		alice: "0x8450c0055cB180C7C37A25866132A740b812937B",
		bob:   "0xc4bA4815c82727554e4c12A07a139b74c6742322",
	}
	defaultConfigFileURL = map[string]string{
		alice: "alice/session.yaml",
		bob:   "bob/session.yaml",
	}
	defaultNodeURL  = "localhost:50001"
	defaultChainURL = "ws://127.0.0.1:8545"

	// Commonly used alignment options.
	verticalMiddle   = []container.Option{container.AlignVertical(align.VerticalMiddle)}
	horizontalCenter = []container.Option{container.AlignHorizontal(align.HorizontalCenter)}
	horizontalLeft   = []container.Option{container.AlignHorizontal(align.HorizontalLeft)}

	// Colors used in connect screen.
	formFieldColor             = cell.Color(16)
	formButtonTextColor        = cell.Color(16)
	formButtonPressedFillColor = cell.Color(16)
	formConnectButtonColor     = cell.Color(3)
	formQuitButtonColor        = cell.Color(2)
)

type connectScreen struct {
	width         int
	elementHeight int
	buttonHeight  int
	cntFields     int

	sessionNameTBox   *textinput.TextInput
	perunNodeURLTBox  *textinput.TextInput
	configFileURLTBox *textinput.TextInput
	chainURLTBox      *textinput.TextInput
	onChainAddrTBox   *textinput.TextInput

	sessionNameTBoxWidget   grid.Element
	perunNodeURLTBoxWidget  grid.Element
	configFileURLTBoxWidget grid.Element
	chainURLTBoxWidget      grid.Element
	onChainAddrTBoxWidget   grid.Element

	connectBtn *button.Button
	quitBtn    *button.Button
}

func newConnectScreen(role string) (*connectScreen, error) {
	width := 65
	elementHeight := 3
	buttonHeight := 1
	// Spaces in text based on above width.
	field1 := "Name              "
	field2 := "Perun Node URL    "
	field3 := "Config file URL   "
	field4 := "Chain URL         "
	field5 := "On-Chain Addr     "
	sessionNameTBox, sessionNameTBoxWidget, err := newField(field1, role)
	if err != nil {
		return nil, errors.WithMessage(err, "initializing session name field")
	}
	perunNodeURLTBox, perunNodeURLTBoxWidget, err := newField(field2, defaultNodeURL)
	if err != nil {
		return nil, errors.WithMessage(err, "initializing perun node URL field")
	}
	configFileURLTBox, configFileURLTBoxWidget, err := newField(field3, defaultConfigFileURL[role])
	if err != nil {
		return nil, errors.WithMessage(err, "initializing config URL field")
	}
	chainURLTBox, chainURLTBoxWidget, err := newField(field4, defaultChainURL)
	if err != nil {
		return nil, errors.WithMessage(err, "initializing chain URL field")
	}
	onChainAddrTBox, onChainAddrTBoxWidget, err := newField(field5, defaultOnChainAddrs[role])
	if err != nil {
		return nil, errors.WithMessage(err, "initializing on-chain address field")
	}

	connectBtn, err := newButton("Connect", formConnectButtonColor, buttonHeight)
	if err != nil {
		return nil, errors.WithMessage(err, "initializing connect button")
	}
	quitBtn, err := newButton("Quit", formQuitButtonColor, buttonHeight)
	if err != nil {
		return nil, errors.WithMessage(err, "initializing quit button")
	}

	return &connectScreen{
		width:         width,
		elementHeight: elementHeight,
		buttonHeight:  buttonHeight,
		cntFields:     5,

		sessionNameTBox:         sessionNameTBox,
		sessionNameTBoxWidget:   sessionNameTBoxWidget,
		perunNodeURLTBox:        perunNodeURLTBox,
		perunNodeURLTBoxWidget:  perunNodeURLTBoxWidget,
		configFileURLTBox:       configFileURLTBox,
		configFileURLTBoxWidget: configFileURLTBoxWidget,

		chainURLTBox:          chainURLTBox,
		chainURLTBoxWidget:    chainURLTBoxWidget,
		onChainAddrTBox:       onChainAddrTBox,
		onChainAddrTBoxWidget: onChainAddrTBoxWidget,
		connectBtn:            connectBtn,
		quitBtn:               quitBtn,
	}, nil
}

func newField(label, defaultValue string) (*textinput.TextInput, grid.Element, error) {
	tBox, err := textinput.New(
		textinput.Label(label),
		textinput.Border(linestyle.Round),
		textinput.FillColor(formFieldColor),
		textinput.DefaultText(defaultValue),
		textinput.ExclusiveKeyboardOnFocus(),
	)
	if err != nil {
		return nil, nil, errors.WithStack(err)
	}
	return tBox, grid.Widget(tBox), nil
}

func newButton(text string, color cell.Color, buttonHeight int) (*button.Button, error) {
	button, err := button.New(text,
		func() error { return nil },
		button.TextHorizontalPadding(1),
		button.TextColor(formButtonTextColor),
		button.FillColor(color),
		button.PressedFillColor(formButtonPressedFillColor),
		button.DisableShadow(),
		button.Height(buttonHeight))
	return button, errors.WithStack(err)
}

// getConnectFn returns the call back for connect button.
// It is a closure over connect screen, passed container, quitter and package level variables.
func (s *connectScreen) getConnectFn(c *container.Container, quitter chan bool) func() error {
	return func() error {
		var err error

		// Read & validate fields; set package level variables.
		perunNodeURL := s.perunNodeURLTBox.Read()
		configFileURL := s.configFileURLTBox.Read()

		userName = s.sessionNameTBox.Read()
		chainURL = s.chainURLTBox.Read()
		onChainAddr, err = ethereum.NewWalletBackend().ParseAddr(s.onChainAddrTBox.Read())
		if err != nil {
			return errors.Wrap(err, "parsing on-chain address")
		}

		// Connect to perun node.
		sessionID, client, err = connectToNode(perunNodeURL, configFileURL)
		if err != nil {
			return errors.WithMessage(err, "connecting to perun node")
		}
		logInfof("Session established. Session ID: %v", sessionID)

		err = subProposals()
		if err != nil {
			return errors.WithMessage(err, "subscribing to proposals")
		}
		logInfo("Subscribed to incoming proposals")

		// Render dashboard screen.
		d, err := newDashboardScreen()
		if err != nil {
			return errors.WithMessage(err, "initializing dashboard screen")
		}
		err = renderDashboardView(c, d)
		if err != nil {
			return errors.WithMessage(err, "rendering dashboard screen")
		}
		go updateOnChainBalNTime(d.onChainBalText, d.timeText, quitter)
		return nil
	}
}

// getQuitFn returns the call back for quit button.
// It sends a message over the passed channel.
func (s *connectScreen) getQuitFn(quitter chan bool) func() error {
	return func() error {
		quitter <- true
		return nil
	}
}

func renderConnectScreen(c *container.Container, s *connectScreen) error {
	totalHeight := y - logsHeight
	verticalFreeSpace := totalHeight - (s.elementHeight*s.cntFields + s.buttonHeight + 1)
	// +1 for space b/w form, btn.

	totalWidth := s.width
	horizontalFreeSpace := x - totalWidth

	builder := grid.New()
	builder.Add(
		grid.RowHeightFixedWithOpts(totalHeight,
			[]container.Option{
				container.PaddingLeft(horizontalFreeSpace / 2),
				container.PaddingRight(horizontalFreeSpace / 2),
			},
			grid.ColWidthFixedWithOpts(totalWidth,
				[]container.Option{
					container.AlignHorizontal(align.HorizontalCenter),
					container.AlignVertical(align.VerticalMiddle),
					container.PaddingTop(verticalFreeSpace / 2),
					container.PaddingBottom(verticalFreeSpace / 2),
				},
				grid.RowHeightFixedWithOpts(s.elementHeight, verticalMiddle, s.sessionNameTBoxWidget),
				grid.RowHeightFixedWithOpts(s.elementHeight, verticalMiddle, s.perunNodeURLTBoxWidget),
				grid.RowHeightFixedWithOpts(s.elementHeight, verticalMiddle, s.configFileURLTBoxWidget),
				grid.RowHeightFixedWithOpts(s.elementHeight, verticalMiddle, s.chainURLTBoxWidget),
				grid.RowHeightFixedWithOpts(s.elementHeight, verticalMiddle, s.onChainAddrTBoxWidget),
				grid.RowHeightFixed(s.buttonHeight+3, // 3 for padding.
					grid.ColWidthFixedWithOpts(totalWidth/2, horizontalCenter, grid.Widget(s.connectBtn)),
					grid.ColWidthFixedWithOpts(totalWidth/2, horizontalCenter, grid.Widget(s.quitBtn))))),
		grid.RowHeightFixedWithOpts(logsHeight,
			[]container.Option{
				container.BorderTitle("Logs"),
				container.Border(linestyle.Round),
				container.BorderColor(logBoxBorderColor),
				container.FocusedColor(logBoxBorderColor),
				container.TitleColor(logBoxTitleColor),
				container.TitleFocusedColor(logBoxTitleColor),
				container.AlignVertical(align.VerticalTop),
				container.AlignHorizontal(align.HorizontalLeft),
			},
			grid.Widget(logBox)))

	gridOpts, err := builder.Build()
	if err != nil {
		return errors.Wrap(err, "building connect screen")
	}

	return errors.Wrap(c.Update(rootContainerID, gridOpts...), "updating connect screen")
}
