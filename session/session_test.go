// Copyright (c) 2020 - for information on the respective copyright owner
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

package session_test

import (
	"context"
	"fmt"
	"math/rand"
	"testing"
	"time"

	"github.com/phayes/freeport"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	pchannel "perun.network/go-perun/channel"
	pclient "perun.network/go-perun/client"
	pwire "perun.network/go-perun/wire"

	"github.com/hyperledger-labs/perun-node"
	"github.com/hyperledger-labs/perun-node/blockchain/ethereum/ethereumtest"
	"github.com/hyperledger-labs/perun-node/currency"
	"github.com/hyperledger-labs/perun-node/idprovider"
	"github.com/hyperledger-labs/perun-node/idprovider/local"
	"github.com/hyperledger-labs/perun-node/internal/mocks"
	"github.com/hyperledger-labs/perun-node/session"
	"github.com/hyperledger-labs/perun-node/session/sessiontest"
)

var RandSeedForNewPeerIDs int64 = 121

func init() {
	session.SetWalletBackend(ethereumtest.NewTestWalletBackend())
}

func Test_SessionAPI_Interface(t *testing.T) {
	assert.Implements(t, (*perun.SessionAPI)(nil), new(session.Session))
}

func newSessionWMockChClient(t *testing.T, isOpen bool, peerIDs ...perun.PeerID) (*session.Session, *mocks.ChClient) {
	prng := rand.New(rand.NewSource(ethereumtest.RandSeedForTestAccs))
	cfg := sessiontest.NewConfigT(t, prng, peerIDs...)
	chClient := &mocks.ChClient{}
	s, err := session.NewSessionForTest(cfg, isOpen, chClient)
	require.NoError(t, err)
	require.NotNil(t, s)
	return s, chClient
}

func Test_Session_AddPeerID(t *testing.T) {
	peerIDs := newPeerIDs(t, uint(3))
	// In openSession, peer0 is already present, peer1 can be added.
	openSession, _ := newSessionWMockChClient(t, true, peerIDs[0])
	closedSession, _ := newSessionWMockChClient(t, false, peerIDs[0])

	t.Run("happy_add_peerID", func(t *testing.T) {
		err := openSession.AddPeerID(peerIDs[1])
		assert.NoError(t, err)
		assert.Nil(t, err)
	})

	t.Run("alias_used_for_diff_peerID", func(t *testing.T) {
		peer1WithAlias0 := peerIDs[1]
		peer1WithAlias0.Alias = peerIDs[0].Alias
		err := openSession.AddPeerID(peer1WithAlias0)
		require.Error(t, err)

		wantMessage := idprovider.ErrPeerAliasAlreadyUsed.Error()
		wantRequirement := "peer alias should be unique for each peer ID"
		assert.Equal(t, perun.ClientError, err.Category())
		assert.Equal(t, perun.ErrV2InvalidArgument, err.Code())
		assert.Equal(t, wantMessage, err.Message())
		addInfo, ok := err.AddInfo().(perun.ErrV2InfoInvalidArgument)
		require.True(t, ok)
		assert.Equal(t, "peer alias", addInfo.Name)
		assert.Equal(t, peer1WithAlias0.Alias, addInfo.Value)
		assert.Equal(t, wantRequirement, addInfo.Requirement)
	})

	t.Run("peerID_already_registered", func(t *testing.T) {
		err := openSession.AddPeerID(peerIDs[0])
		require.Error(t, err)

		wantMessage := idprovider.ErrPeerIDAlreadyRegistered.Error()
		assert.Equal(t, perun.ClientError, err.Category())
		assert.Equal(t, perun.ErrV2ResourceExists, err.Code())
		assert.Equal(t, wantMessage, err.Message())
		addInfo, ok := err.AddInfo().(perun.ErrV2InfoResourceExists)
		require.True(t, ok)
		assert.Equal(t, "peer alias", addInfo.Type)
		assert.Equal(t, peerIDs[0].Alias, addInfo.ID)
	})

	t.Run("peerID_address_string_too_long", func(t *testing.T) {
		peer1WithInvalidAddrString := peerIDs[2]
		peer1WithInvalidAddrString.OffChainAddrString = "0x931d387731BBbc988B31221D387706c74f77d004d6b84b"
		err := openSession.AddPeerID(peer1WithInvalidAddrString)
		require.Error(t, err)

		wantMessage := idprovider.ErrParsingOffChainAddress.Error()
		assert.Equal(t, perun.ClientError, err.Category())
		assert.Equal(t, perun.ErrV2InvalidArgument, err.Code())
		assert.Contains(t, err.Message(), wantMessage)
		addInfo, ok := err.AddInfo().(perun.ErrV2InfoInvalidArgument)
		require.True(t, ok)
		assert.Equal(t, "off-chain address string", addInfo.Name)
		assert.Equal(t, peer1WithInvalidAddrString.OffChainAddrString, addInfo.Value)
	})

	t.Run("session_closed", func(t *testing.T) {
		err := closedSession.AddPeerID(peerIDs[0])
		require.Error(t, err)

		wantMessage := session.ErrSessionClosed.Error()
		assert.Equal(t, perun.ClientError, err.Category())
		assert.Equal(t, perun.ErrV2FailedPreCondition, err.Code())
		assert.Equal(t, wantMessage, err.Message())
		assert.Nil(t, err.AddInfo())
	})
}

func Test_Session_GetPeerID(t *testing.T) {
	peerIDs := newPeerIDs(t, uint(1))
	// In openSession, peer0 is present and peer1 is not present.
	openSession, _ := newSessionWMockChClient(t, true, peerIDs[0])
	closedSession, _ := newSessionWMockChClient(t, false, peerIDs[0])

	t.Run("happy_get_contact", func(t *testing.T) {
		peerID, err := openSession.GetPeerID(peerIDs[0].Alias)
		require.NoError(t, err)
		assert.True(t, local.PeerIDEqual(peerID, peerIDs[0]))
	})

	t.Run("peerID_not_found", func(t *testing.T) {
		unknownAlias := "unknown-alias"
		_, err := openSession.GetPeerID(unknownAlias)
		require.Error(t, err)

		wantMessage := session.ErrUnknownPeerAlias
		assert.Equal(t, perun.ClientError, err.Category())
		assert.Equal(t, perun.ErrV2ResourceNotFound, err.Code())
		assert.Contains(t, err.Message(), wantMessage)
		addInfo, ok := err.AddInfo().(perun.ErrV2InfoResourceNotFound)
		require.True(t, ok)
		assert.Equal(t, "peer alias", addInfo.Type)
		assert.Equal(t, unknownAlias, addInfo.ID)
	})

	t.Run("session_closed", func(t *testing.T) {
		_, err := closedSession.GetPeerID(peerIDs[0].Alias)
		require.Error(t, err)

		wantMessage := session.ErrSessionClosed.Error()
		assert.Equal(t, perun.ClientError, err.Category())
		assert.Equal(t, perun.ErrV2FailedPreCondition, err.Code())
		assert.Equal(t, wantMessage, err.Message())
		assert.Nil(t, err.AddInfo())
	})
}

func makeState(t *testing.T, balInfo perun.BalInfo, isFinal bool) *pchannel.State {
	allocation, err := session.MakeAllocation(balInfo, nil)
	require.NoError(t, err)
	return &pchannel.State{
		ID:         [32]byte{0},
		Version:    0,
		App:        pchannel.NoApp(),
		Allocation: *allocation,
		Data:       pchannel.NoData(),
		IsFinal:    isFinal,
	}
}

func newMockPCh(t *testing.T, openingBalInfo perun.BalInfo) (
	*mocks.Channel, chan time.Time) {
	ch := &mocks.Channel{}
	ch.On("ID").Return([32]byte{0, 1, 2})
	ch.On("State").Return(makeState(t, openingBalInfo, false))
	watcherSignal := make(chan time.Time)
	ch.On("Watch").WaitUntil(watcherSignal).Return(nil)
	return ch, watcherSignal
}

func Test_Session_OpenCh(t *testing.T) {
	peerIDs := newPeerIDs(t, uint(2))
	validOpeningBalInfo := perun.BalInfo{
		Currency: currency.ETH,
		Parts:    []string{perun.OwnAlias, peerIDs[0].Alias},
		Bal:      []string{"1", "2"},
	}
	app := perun.App{
		Def:  pchannel.NoApp(),
		Data: pchannel.NoData(),
	}

	t.Run("happy_1_own_alias_first", func(t *testing.T) {
		ch, _ := newMockPCh(t, validOpeningBalInfo)
		session, chClient := newSessionWMockChClient(t, true, peerIDs...)
		chClient.On("ProposeChannel", mock.Anything, mock.Anything).Return(ch, nil)
		chClient.On("Register", mock.Anything, mock.Anything).Return()

		chInfo, err := session.OpenCh(context.Background(), validOpeningBalInfo, app, 10)
		require.NoError(t, err)
		require.NotZero(t, chInfo)
	})

	t.Run("happy_2_own_alias_not_first", func(t *testing.T) {
		validOpeningBalInfo2 := validOpeningBalInfo
		validOpeningBalInfo2.Parts = []string{peerIDs[0].Alias, perun.OwnAlias}

		ch, _ := newMockPCh(t, validOpeningBalInfo2)
		session, chClient := newSessionWMockChClient(t, true, peerIDs...)
		chClient.On("ProposeChannel", mock.Anything, mock.Anything).Return(ch, nil)
		chClient.On("Register", mock.Anything, mock.Anything).Return()

		chInfo, err := session.OpenCh(context.Background(), validOpeningBalInfo2, app, 10)
		require.NoError(t, err)
		require.NotZero(t, chInfo)
	})

	t.Run("session_closed", func(t *testing.T) {
		ch, _ := newMockPCh(t, validOpeningBalInfo)
		session, chClient := newSessionWMockChClient(t, false, peerIDs...)
		chClient.On("ProposeChannel", mock.Anything, mock.Anything).Return(ch, nil)
		chClient.On("Register", mock.Anything, mock.Anything).Return()

		_, err := session.OpenCh(context.Background(), validOpeningBalInfo, app, 10)
		require.Error(t, err)
		t.Log(err)
	})

	t.Run("missing_parts", func(t *testing.T) {
		invalidOpeningBalInfo := validOpeningBalInfo
		invalidOpeningBalInfo.Parts = []string{perun.OwnAlias, "missing-part"}
		session, _ := newSessionWMockChClient(t, true, peerIDs...)

		_, err := session.OpenCh(context.Background(), invalidOpeningBalInfo, app, 10)
		require.Error(t, err)
		t.Log(err)
	})

	t.Run("repeated_parts", func(t *testing.T) {
		invalidOpeningBalInfo := validOpeningBalInfo
		invalidOpeningBalInfo.Parts = []string{peerIDs[0].Alias, peerIDs[0].Alias}
		session, _ := newSessionWMockChClient(t, true, peerIDs...)

		_, err := session.OpenCh(context.Background(), invalidOpeningBalInfo, app, 10)
		require.Error(t, err)
		t.Log(err)
	})

	t.Run("missing_own_alias", func(t *testing.T) {
		invalidOpeningBalInfo := validOpeningBalInfo
		invalidOpeningBalInfo.Parts = []string{peerIDs[0].Alias, peerIDs[1].Alias}
		session, _ := newSessionWMockChClient(t, true, peerIDs...)

		_, err := session.OpenCh(context.Background(), invalidOpeningBalInfo, app, 10)
		require.Error(t, err)
		t.Log(err)
	})

	t.Run("unsupported_currency", func(t *testing.T) {
		invalidOpeningBalInfo := validOpeningBalInfo
		invalidOpeningBalInfo.Currency = "unsupported-currency"
		session, chClient := newSessionWMockChClient(t, true, peerIDs...)
		chClient.On("Register", mock.Anything, mock.Anything).Return()

		_, err := session.OpenCh(context.Background(), invalidOpeningBalInfo, app, 10)
		require.Error(t, err)
		t.Log(err)
	})

	t.Run("invalid_amount", func(t *testing.T) {
		invalidOpeningBalInfo := validOpeningBalInfo
		invalidOpeningBalInfo.Bal = []string{"abc", "gef"}
		session, chClient := newSessionWMockChClient(t, true, peerIDs...)
		chClient.On("Register", mock.Anything, mock.Anything).Return()

		_, err := session.OpenCh(context.Background(), invalidOpeningBalInfo, app, 10)
		require.Error(t, err)
		t.Log(err)
	})

	t.Run("chClient_proposeChannel_AnError", func(t *testing.T) {
		ch, _ := newMockPCh(t, validOpeningBalInfo)
		session, chClient := newSessionWMockChClient(t, true, peerIDs...)
		chClient.On("ProposeChannel", mock.Anything, mock.Anything).Return(ch, assert.AnError)
		chClient.On("Register", mock.Anything, mock.Anything).Return()

		_, err := session.OpenCh(context.Background(), validOpeningBalInfo, app, 10)
		require.Error(t, err)
		t.Log(err)
	})

	t.Run("chClient_proposeChannel_PeerRejected", func(t *testing.T) {
		ch, _ := newMockPCh(t, validOpeningBalInfo)
		session, chClient := newSessionWMockChClient(t, true, peerIDs...)
		chClient.On("ProposeChannel", mock.Anything, mock.Anything).Return(ch, errors.New("channel proposal rejected"))
		chClient.On("Register", mock.Anything, mock.Anything).Return()

		_, err := session.OpenCh(context.Background(), validOpeningBalInfo, app, 10)
		require.Error(t, err)
		t.Log(err)
	})
}

func newChProposal(t *testing.T, ownAddr, peer perun.PeerID) pclient.ChannelProposal {
	prng := rand.New(rand.NewSource(121))
	chAsset := ethereumtest.NewRandomAddress(prng)

	openingBalInfo := perun.BalInfo{
		Currency: currency.ETH,
		Parts:    []string{peer.Alias, perun.OwnAlias},
		Bal:      []string{"1", "2"},
	}
	allocation, err := session.MakeAllocation(openingBalInfo, chAsset)
	require.NoError(t, err)

	return pclient.NewLedgerChannelProposal(10, ownAddr.OffChainAddr, allocation,
		[]pwire.Address{peer.OffChainAddr, ownAddr.OffChainAddr},
		pclient.WithApp(pchannel.NoApp(), pchannel.NoData()), pclient.WithRandomNonce())
}

func newSessionWChProposal(t *testing.T, peerIDs []perun.PeerID) (
	*session.Session, pclient.ChannelProposal, string) {
	session, _ := newSessionWMockChClient(t, true, peerIDs...)
	ownPeerID, err := session.GetPeerID(perun.OwnAlias)
	require.NoError(t, err)
	chProposal := newChProposal(t, ownPeerID, peerIDs[0])
	chProposalID := fmt.Sprintf("%x", chProposal.Base().ProposalID())
	return session, chProposal, chProposalID
}

func Test_Session_HandleProposalWInterface(t *testing.T) {
	peerIDs := newPeerIDs(t, uint(1))

	t.Run("happy", func(t *testing.T) {
		session, chProposal, _ := newSessionWChProposal(t, peerIDs)

		responder := &mocks.ChProposalResponder{}
		session.HandleProposalWInterface(chProposal, responder)
	})

	t.Run("unknown_peer", func(t *testing.T) {
		session, _ := newSessionWMockChClient(t, true) // Don't register any peer in ID provider.
		ownPeerID, err := session.GetPeerID(perun.OwnAlias)
		require.NoError(t, err)
		unknownPeerID := peerIDs[0]
		chProposal := newChProposal(t, ownPeerID, unknownPeerID)

		responder := &mocks.ChProposalResponder{}
		responder.On("Reject", mock.Anything, mock.Anything).Return(nil)
		session.HandleProposalWInterface(chProposal, responder)
	})

	t.Run("session_closed", func(t *testing.T) {
		session, _ := newSessionWMockChClient(t, false, peerIDs...)
		chProposal := &mocks.ChannelProposal{}

		responder := &mocks.ChProposalResponder{}
		session.HandleProposalWInterface(chProposal, responder)
	})
}

func Test_SubUnsubChProposal(t *testing.T) {
	dummyNotifier := func(notif perun.ChProposalNotif) {}
	openSession, _ := newSessionWMockChClient(t, true)
	closedSession, _ := newSessionWMockChClient(t, false)

	// Note: All sub tests are written at the same level because each sub test modifies the state of session
	// and the order of execution needs to be maintained.

	// == SubTest 1: Sub successfully ==
	err := openSession.SubChProposals(dummyNotifier)
	require.NoError(t, err)

	// == SubTest 2: Sub again, should error ==
	err = openSession.SubChProposals(dummyNotifier)
	require.Error(t, err)

	// == SubTest 3: Unsub successfully ==
	err = openSession.UnsubChProposals()
	require.NoError(t, err)

	// == SubTest 4: Unsub again, should error ==
	err = openSession.UnsubChProposals()
	require.Error(t, err)

	t.Run("Sub_sessionClosed", func(t *testing.T) {
		err = closedSession.SubChProposals(dummyNotifier)
		require.Error(t, err)
	})

	t.Run("Unsub_sessionClosed", func(t *testing.T) {
		err = closedSession.UnsubChProposals()
		require.Error(t, err)
	})
}

func Test_HandleProposalWInterface_Sub(t *testing.T) {
	peerIDs := newPeerIDs(t, uint(1)) // Aliases of peerIDs are their respective indices in the array.

	t.Run("happy_HandleSub", func(t *testing.T) {
		session, chProposal, _ := newSessionWChProposal(t, peerIDs)

		responder := &mocks.ChProposalResponder{}
		session.HandleProposalWInterface(chProposal, responder)
		notifs := make([]perun.ChProposalNotif, 0, 2)
		notifier := func(notif perun.ChProposalNotif) {
			notifs = append(notifs, notif)
		}

		err := session.SubChProposals(notifier)
		require.NoError(t, err)
		notifRecieved := func() bool {
			return len(notifs) == 1
		}
		assert.Eventually(t, notifRecieved, 2*time.Second, 100*time.Millisecond)
	})

	t.Run("happy_SubHandle", func(t *testing.T) {
		session, chProposal, _ := newSessionWChProposal(t, peerIDs)

		notifs := make([]perun.ChProposalNotif, 0, 2)
		notifier := func(notif perun.ChProposalNotif) {
			notifs = append(notifs, notif)
		}
		err := session.SubChProposals(notifier)
		require.NoError(t, err)
		responder := &mocks.ChProposalResponder{}

		session.HandleProposalWInterface(chProposal, responder)
		notifRecieved := func() bool {
			return len(notifs) == 1
		}
		assert.Eventually(t, notifRecieved, 2*time.Second, 100*time.Millisecond)
	})
}

func Test_HandleProposalWInterface_Respond(t *testing.T) {
	peerIDs := newPeerIDs(t, uint(1)) // Aliases of peerIDs are their respective indices in the array.

	openingBalInfo := perun.BalInfo{
		Currency: currency.ETH,
		Parts:    []string{peerIDs[0].Alias, perun.OwnAlias},
		Bal:      []string{"1", "2"},
	}

	t.Run("happy_Handle_Respond_Accept", func(t *testing.T) {
		session, chProposal, chProposalID := newSessionWChProposal(t, peerIDs)

		ch, _ := newMockPCh(t, openingBalInfo)
		responder := &mocks.ChProposalResponder{}
		responder.On("Accept", mock.Anything, mock.Anything).Return(ch, nil)
		session.HandleProposalWInterface(chProposal, responder)

		gotChInfo, err := session.RespondChProposal(context.Background(), chProposalID, true)
		require.NoError(t, err)
		assert.Equal(t, gotChInfo.ChID, fmt.Sprintf("%x", ch.ID()))
	})

	t.Run("happy_Handle_Respond_Reject", func(t *testing.T) {
		session, chProposal, chProposalID := newSessionWChProposal(t, peerIDs)

		responder := &mocks.ChProposalResponder{}
		responder.On("Reject", mock.Anything, mock.Anything).Return(nil)
		session.HandleProposalWInterface(chProposal, responder)

		_, err := session.RespondChProposal(context.Background(), chProposalID, false)
		assert.NoError(t, err)
	})

	t.Run("happy_Handle_Respond_Accept_Error", func(t *testing.T) {
		session, chProposal, chProposalID := newSessionWChProposal(t, peerIDs)

		ch, _ := newMockPCh(t, openingBalInfo)
		responder := &mocks.ChProposalResponder{}
		responder.On("Accept", mock.Anything, mock.Anything).Return(ch, assert.AnError)
		session.HandleProposalWInterface(chProposal, responder)

		_, err := session.RespondChProposal(context.Background(), chProposalID, true)
		assert.Error(t, err)
		t.Log(err)
	})

	t.Run("Handle_Respond_Reject_Error", func(t *testing.T) {
		session, chProposal, chProposalID := newSessionWChProposal(t, peerIDs)

		responder := &mocks.ChProposalResponder{}
		responder.On("Reject", mock.Anything, mock.Anything).Return(assert.AnError)
		session.HandleProposalWInterface(chProposal, responder)

		_, err := session.RespondChProposal(context.Background(), chProposalID, false)
		assert.Error(t, err)
		t.Log(err)
	})

	t.Run("Respond_Unknonwn_ProposalID", func(t *testing.T) {
		session, _ := newSessionWMockChClient(t, true, peerIDs...)

		_, err := session.RespondChProposal(context.Background(), "unknown-proposal-id", true)
		require.Error(t, err)
		t.Log(err)
	})

	t.Run("Handle_Respond_Timeout", func(t *testing.T) {
		chClient := &mocks.ChClient{} // Dummy ChClient is sufficient as no methods on it will be invoked.
		prng := rand.New(rand.NewSource(ethereumtest.RandSeedForTestAccs))
		modifiedCfg := sessiontest.NewConfigT(t, prng, peerIDs...)
		modifiedCfg.ResponseTimeout = 1 * time.Second
		session, err := session.NewSessionForTest(modifiedCfg, true, chClient)
		require.NoError(t, err)
		require.NotNil(t, session)

		ownPeerID, err := session.GetPeerID(perun.OwnAlias)
		require.NoError(t, err)
		chProposal := newChProposal(t, ownPeerID, peerIDs[0])
		chProposalID := fmt.Sprintf("%x", chProposal.Base().ProposalID())

		responder := &mocks.ChProposalResponder{} // Dummy responder is sufficient as no methods on it will be invoked.
		session.HandleProposalWInterface(chProposal, responder)
		time.Sleep(2 * time.Second) // Wait until the notification expires.
		_, err = session.RespondChProposal(context.Background(), chProposalID, true)
		assert.Error(t, err)
		t.Log(err)
	})

	t.Run("Respond_Session_Closed", func(t *testing.T) {
		session, _ := newSessionWMockChClient(t, false, peerIDs...)

		chProposalID := "any-proposal-id" // A closed session returns error irrespective of proposal id.
		_, err := session.RespondChProposal(context.Background(), chProposalID, true)
		assert.Error(t, err)
		t.Log(err)
	})
}

func Test_ProposeCh_GetChsInfo(t *testing.T) {
	peerIDs := newPeerIDs(t, uint(2))
	prng := rand.New(rand.NewSource(ethereumtest.RandSeedForTestAccs))
	cfg := sessiontest.NewConfigT(t, prng, peerIDs...)
	validOpeningBalInfo := perun.BalInfo{
		Currency: currency.ETH,
		Parts:    []string{perun.OwnAlias, peerIDs[0].Alias},
		Bal:      []string{"1", "2"},
	}
	app := perun.App{
		Def:  pchannel.NoApp(),
		Data: pchannel.NoData(),
	}
	ch, _ := newMockPCh(t, validOpeningBalInfo)
	chClient := &mocks.ChClient{}
	chClient.On("ProposeChannel", mock.Anything, mock.Anything).Return(ch, nil)
	chClient.On("Register", mock.Anything, mock.Anything).Return()
	session, err := session.NewSessionForTest(cfg, true, chClient)
	require.NoError(t, err)
	require.NotNil(t, session)

	chInfo, err := session.OpenCh(context.Background(), validOpeningBalInfo, app, 10)
	require.NoError(t, err)
	require.NotZero(t, chInfo)

	t.Run("happy", func(t *testing.T) {
		chID := fmt.Sprintf("%x", ch.ID())
		chsInfo := session.GetChsInfo()
		assert.Len(t, chsInfo, 1)
		assert.Equal(t, chsInfo[0].ChID, chID)
	})
}

func Test_ProposeCh_GetCh(t *testing.T) {
	peerIDs := newPeerIDs(t, uint(2))
	prng := rand.New(rand.NewSource(ethereumtest.RandSeedForTestAccs))
	cfg := sessiontest.NewConfigT(t, prng, peerIDs...)
	validOpeningBalInfo := perun.BalInfo{
		Currency: currency.ETH,
		Parts:    []string{perun.OwnAlias, peerIDs[0].Alias},
		Bal:      []string{"1", "2"},
	}
	app := perun.App{
		Def:  pchannel.NoApp(),
		Data: pchannel.NoData(),
	}
	ch, _ := newMockPCh(t, validOpeningBalInfo)
	chClient := &mocks.ChClient{}
	chClient.On("ProposeChannel", mock.Anything, mock.Anything).Return(ch, nil)
	chClient.On("Register", mock.Anything, mock.Anything).Return()
	session, err := session.NewSessionForTest(cfg, true, chClient)
	require.NoError(t, err)
	require.NotNil(t, session)

	chInfo, err := session.OpenCh(context.Background(), validOpeningBalInfo, app, 10)
	require.NoError(t, err)
	require.NotZero(t, chInfo)

	t.Run("happy", func(t *testing.T) {
		chID := fmt.Sprintf("%x", ch.ID())
		gotCh, err := session.GetCh(chID)
		require.NoError(t, err)
		assert.Equal(t, gotCh.ID(), chID)
	})

	t.Run("unknownChID", func(t *testing.T) {
		_, err := session.GetCh("unknown-ch-ID")
		require.Error(t, err)
	})
}

func newSessionWCh(t *testing.T, peerIDs []perun.PeerID, openingBalInfo perun.BalInfo,
	ch perun.Channel) *session.Session {
	app := perun.App{
		Def:  pchannel.NoApp(),
		Data: pchannel.NoData(),
	}
	session, chClient := newSessionWMockChClient(t, true, peerIDs...)
	chClient.On("ProposeChannel", mock.Anything, mock.Anything).Return(ch, nil)
	chClient.On("Register", mock.Anything, mock.Anything).Return()
	chClient.On("Close", mock.Anything).Return(nil)

	chInfo, err := session.OpenCh(context.Background(), openingBalInfo, app, 10)
	require.NoError(t, err)
	require.NotZero(t, chInfo)

	return session
}

func Test_ProposeCh_CloseSession(t *testing.T) {
	peerIDs := newPeerIDs(t, uint(2))
	validOpeningBalInfo := perun.BalInfo{
		Currency: currency.ETH,
		Parts:    []string{perun.OwnAlias, peerIDs[0].Alias},
		Bal:      []string{"1", "2"},
	}
	t.Run("happy_no_force", func(t *testing.T) {
		ch, _ := newMockPCh(t, validOpeningBalInfo)
		ch.On("Phase").Return(pchannel.Acting)
		session, chClient := newSessionWMockChClient(t, true, peerIDs...)
		chClient.On("ProposeChannel", mock.Anything, mock.Anything).Return(ch, nil)
		chClient.On("Register", mock.Anything, mock.Anything).Return()
		chClient.On("Close", mock.Anything).Return(nil)

		persistedChs, err := session.Close(false)
		require.NoError(t, err)
		assert.Len(t, persistedChs, 0)
	})
	t.Run("happy_force", func(t *testing.T) {
		ch, _ := newMockPCh(t, validOpeningBalInfo)
		ch.On("Phase").Return(pchannel.Acting)
		session := newSessionWCh(t, peerIDs, validOpeningBalInfo, ch)

		persistedChs, err := session.Close(true)
		require.NoError(t, err)
		assert.Len(t, persistedChs, 1)
	})
	t.Run("no_force_openChs", func(t *testing.T) {
		ch, _ := newMockPCh(t, validOpeningBalInfo)
		ch.On("Phase").Return(pchannel.Acting)
		session := newSessionWCh(t, peerIDs, validOpeningBalInfo, ch)

		_, err := session.Close(false)
		require.Error(t, err)
		t.Log(err)
	})
	t.Run("no_force_openChs", func(t *testing.T) {
		ch, _ := newMockPCh(t, validOpeningBalInfo)
		ch.On("Phase").Return(pchannel.Acting)
		session := newSessionWCh(t, peerIDs, validOpeningBalInfo, ch)

		_, err := session.Close(false)
		require.Error(t, err)
		t.Log(err)
	})

	t.Run("force_unexpectedPhaseChs", func(t *testing.T) {
		ch, _ := newMockPCh(t, validOpeningBalInfo)
		ch.On("Phase").Return(pchannel.Registering)
		session := newSessionWCh(t, peerIDs, validOpeningBalInfo, ch)

		_, err := session.Close(false)
		require.Error(t, err)
		t.Log(err)
	})
	t.Run("session_closed", func(t *testing.T) {
		session, _ := newSessionWMockChClient(t, false)

		_, err := session.Close(false)
		require.Error(t, err)
		t.Log(err)
	})
}

func Test_Session_HandleUpdateWInterface(t *testing.T) {
	t.Run("happy", func(t *testing.T) {
		peerIDs := newPeerIDs(t, uint(2))
		validOpeningBalInfo := perun.BalInfo{
			Currency: currency.ETH,
			Parts:    []string{perun.OwnAlias, peerIDs[0].Alias},
			Bal:      []string{"1", "2"},
		}
		updatedBalInfo := validOpeningBalInfo
		updatedBalInfo.Bal = []string{"0.5", "2.5"}

		chUpdate := &pclient.ChannelUpdate{
			State: makeState(t, updatedBalInfo, false),
		}
		session, _ := newSessionWMockChClient(t, true)
		responder := &mocks.ChUpdateResponder{}
		responder.On("Reject", mock.Anything, mock.Anything).Return(nil)
		session.HandleUpdateWInterface(*chUpdate, responder)
	})
	t.Run("session_closed", func(t *testing.T) {
		session, _ := newSessionWMockChClient(t, false)
		session.HandleUpdate(pclient.ChannelUpdate{}, new(pclient.UpdateResponder))
	})
}

func newPeerIDs(t *testing.T, n uint) []perun.PeerID {
	ethereumBackend := ethereumtest.NewTestWalletBackend()
	// Use same prng for each call.
	prng := rand.New(rand.NewSource(RandSeedForNewPeerIDs))
	peerIDs := make([]perun.PeerID, n)
	for i := range peerIDs {
		port, err := freeport.GetFreePort()
		require.NoError(t, err)
		peerIDs[i].Alias = fmt.Sprintf("%d", i)
		peerIDs[i].OffChainAddrString = ethereumtest.NewRandomAddress(prng).String()
		peerIDs[i].CommType = "tcp"
		peerIDs[i].CommAddr = fmt.Sprintf("127.0.0.1:%d", port)

		peerIDs[i].OffChainAddr, err = ethereumBackend.ParseAddr(peerIDs[i].OffChainAddrString)
		require.NoError(t, err)
	}
	return peerIDs
}
