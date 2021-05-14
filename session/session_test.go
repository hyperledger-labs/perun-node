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
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	pethchannel "perun.network/go-perun/backend/ethereum/channel"
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
		assertAPIError(t, err, perun.ClientError, perun.ErrV2InvalidArgument, wantMessage)
		assertErrV2InfoInvalidArgument(t, err.AddInfo(), "peer alias", peer1WithAlias0.Alias)
	})

	t.Run("peerID_already_registered", func(t *testing.T) {
		err := openSession.AddPeerID(peerIDs[0])
		require.Error(t, err)

		wantMessage := idprovider.ErrPeerIDAlreadyRegistered.Error()
		assertAPIError(t, err, perun.ClientError, perun.ErrV2ResourceExists, wantMessage)
		assertErrV2InfoResourceExists(t, err.AddInfo(), "peer alias", peerIDs[0].Alias)
	})

	t.Run("peerID_address_string_too_long", func(t *testing.T) {
		peer1WithInvalidAddrString := peerIDs[2]
		peer1WithInvalidAddrString.OffChainAddrString = "0x931d387731BBbc988B31221D387706c74f77d004d6b84b"
		err := openSession.AddPeerID(peer1WithInvalidAddrString)
		require.Error(t, err)

		wantMessage := idprovider.ErrParsingOffChainAddress.Error()
		assertAPIError(t, err, perun.ClientError, perun.ErrV2InvalidArgument, wantMessage)
		argumentValue := peer1WithInvalidAddrString.OffChainAddrString
		assertErrV2InfoInvalidArgument(t, err.AddInfo(), "off-chain address string", argumentValue)
	})

	t.Run("session_closed", func(t *testing.T) {
		err := closedSession.AddPeerID(peerIDs[0])
		require.Error(t, err)

		wantMessage := session.ErrSessionClosed.Error()
		assertAPIError(t, err, perun.ClientError, perun.ErrV2FailedPreCondition, wantMessage)
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

		wantMessage := session.ErrUnknownPeerAlias.Error()
		assertAPIError(t, err, perun.ClientError, perun.ErrV2ResourceNotFound, wantMessage)
		assertErrV2InfoResourceNotFound(t, err.AddInfo(), "peer alias", unknownAlias)
	})

	t.Run("session_closed", func(t *testing.T) {
		_, err := closedSession.GetPeerID(peerIDs[0].Alias)
		require.Error(t, err)

		wantMessage := session.ErrSessionClosed.Error()
		assertAPIError(t, err, perun.ClientError, perun.ErrV2FailedPreCondition, wantMessage)
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
	var chID [32]byte
	rand.Read(chID[:])
	ch := &mocks.Channel{}
	ch.On("ID").Return(chID)
	ch.On("State").Return(makeState(t, openingBalInfo, false))
	watcherSignal := make(chan time.Time)
	ch.On("Watch", mock.Anything).WaitUntil(watcherSignal).Return(nil)
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
		sess, chClient := newSessionWMockChClient(t, false, peerIDs...)
		chClient.On("ProposeChannel", mock.Anything, mock.Anything).Return(ch, nil)
		chClient.On("Register", mock.Anything, mock.Anything).Return()

		_, err := sess.OpenCh(context.Background(), validOpeningBalInfo, app, 10)
		require.Error(t, err)

		wantMessage := session.ErrSessionClosed.Error()
		assertAPIError(t, err, perun.ClientError, perun.ErrV2FailedPreCondition, wantMessage)
		assert.Nil(t, err.AddInfo())
	})

	t.Run("one_unknown_peer_alias", func(t *testing.T) {
		unknownAlias := "unknown-alias"
		invalidOpeningBalInfo := validOpeningBalInfo
		invalidOpeningBalInfo.Parts = []string{perun.OwnAlias, unknownAlias}
		sess, _ := newSessionWMockChClient(t, true, peerIDs...)

		_, err := sess.OpenCh(context.Background(), invalidOpeningBalInfo, app, 10)
		require.Error(t, err)

		wantMessage := session.ErrUnknownPeerAlias.Error()
		assertAPIError(t, err, perun.ClientError, perun.ErrV2ResourceNotFound, wantMessage)
		assertErrV2InfoResourceNotFound(t, err.AddInfo(), "peer alias", unknownAlias)
	})

	t.Run("two_unknown_peer_aliases", func(t *testing.T) {
		unknownAlias1 := "unknown-alias-1"
		unknownAlias2 := "unknown-alias-2"
		invalidOpeningBalInfo := validOpeningBalInfo
		invalidOpeningBalInfo.Parts = []string{unknownAlias1, unknownAlias2}
		sess, _ := newSessionWMockChClient(t, true, peerIDs...)

		_, err := sess.OpenCh(context.Background(), invalidOpeningBalInfo, app, 10)
		require.Error(t, err)

		wantMessage := session.ErrUnknownPeerAlias.Error()
		assertAPIError(t, err, perun.ClientError, perun.ErrV2ResourceNotFound, wantMessage)
		resourceType := "peer alias"
		resourceID := fmt.Sprintf("%s,%s", unknownAlias1, unknownAlias2)
		assertErrV2InfoResourceNotFound(t, err.AddInfo(), resourceType, resourceID)
	})

	t.Run("repeated_peer_aliases", func(t *testing.T) {
		invalidOpeningBalInfo := validOpeningBalInfo
		invalidOpeningBalInfo.Parts = []string{peerIDs[0].Alias, peerIDs[0].Alias}
		sess, _ := newSessionWMockChClient(t, true, peerIDs...)

		_, err := sess.OpenCh(context.Background(), invalidOpeningBalInfo, app, 10)
		require.Error(t, err)

		wantMessage := session.ErrRepeatedPeerAlias.Error()
		assertAPIError(t, err, perun.ClientError, perun.ErrV2InvalidArgument, wantMessage)
		resourceType := "peer alias"
		assertErrV2InfoInvalidArgument(t, err.AddInfo(), resourceType, peerIDs[0].Alias)
	})

	t.Run("missing_own_alias", func(t *testing.T) {
		invalidOpeningBalInfo := validOpeningBalInfo
		invalidOpeningBalInfo.Parts = []string{peerIDs[0].Alias, peerIDs[1].Alias}
		sess, _ := newSessionWMockChClient(t, true, peerIDs...)

		_, err := sess.OpenCh(context.Background(), invalidOpeningBalInfo, app, 10)
		require.Error(t, err)

		wantMessage := session.ErrNoEntryForSelf.Error()
		assertAPIError(t, err, perun.ClientError, perun.ErrV2InvalidArgument, wantMessage)
		argumentName := "peer alias"
		argumentValue := fmt.Sprintf("%s,%s", peerIDs[0].Alias, peerIDs[1].Alias)
		assertErrV2InfoInvalidArgument(t, err.AddInfo(), argumentName, argumentValue)
	})

	t.Run("unsupported_currency", func(t *testing.T) {
		invalidOpeningBalInfo := validOpeningBalInfo
		invalidOpeningBalInfo.Currency = "unsupported-currency"
		sess, chClient := newSessionWMockChClient(t, true, peerIDs...)
		chClient.On("Register", mock.Anything, mock.Anything).Return()

		_, err := sess.OpenCh(context.Background(), invalidOpeningBalInfo, app, 10)
		require.Error(t, err)

		wantMessage := session.ErrUnknownCurrency.Error()
		argumentName := "currency"
		assertAPIError(t, err, perun.ClientError, perun.ErrV2InvalidArgument, wantMessage)
		assertErrV2InfoInvalidArgument(t, err.AddInfo(), argumentName, invalidOpeningBalInfo.Currency)
	})

	t.Run("invalid_amount", func(t *testing.T) {
		invalidOpeningBalInfo := validOpeningBalInfo
		invalidOpeningBalInfo.Bal = []string{"abc", "gef"}
		sess, chClient := newSessionWMockChClient(t, true, peerIDs...)
		chClient.On("Register", mock.Anything, mock.Anything).Return()

		_, err := sess.OpenCh(context.Background(), invalidOpeningBalInfo, app, 10)
		require.Error(t, err)

		wantMessage := session.ErrInvalidAmountInBalance.Error()
		argumentName := "amount"
		assertAPIError(t, err, perun.ClientError, perun.ErrV2InvalidArgument, wantMessage)
		assertErrV2InfoInvalidArgument(t, err.AddInfo(), argumentName, invalidOpeningBalInfo.Bal[0])
	})

	t.Run("chClient_proposeChannel_AnError", func(t *testing.T) {
		anError := assert.AnError
		ch, _ := newMockPCh(t, validOpeningBalInfo)
		sess, chClient := newSessionWMockChClient(t, true, peerIDs...)
		chClient.On("Register", mock.Anything, mock.Anything).Return()
		chClient.On("ProposeChannel", mock.Anything, mock.Anything).Return(ch, anError)

		_, err := sess.OpenCh(context.Background(), validOpeningBalInfo, app, 10)
		require.Error(t, err)

		wantMessage := "proposing channel"
		assertAPIError(t, err, perun.InternalError, perun.ErrV2UnknownInternal, wantMessage)
	})

	t.Run("chClient_proposeChannel_PeerRequestTimedOut", func(t *testing.T) {
		timeout := sessiontest.ResponseTimeout.String()
		peerRequestTimedOutError := pclient.RequestTimedOutError("some-error")
		ch, _ := newMockPCh(t, validOpeningBalInfo)
		sess, chClient := newSessionWMockChClient(t, true, peerIDs...)
		chClient.On("Register", mock.Anything, mock.Anything).Return()
		chClient.On("ProposeChannel", mock.Anything, mock.Anything).Return(ch, peerRequestTimedOutError)

		_, err := sess.OpenCh(context.Background(), validOpeningBalInfo, app, 10)
		require.Error(t, err)

		wantMessage := peerRequestTimedOutError.Error()
		peerAlias := peerIDs[0].Alias // peer in validOpeningBal is peerIDs[0].
		assertAPIError(t, err, perun.ParticipantError, perun.ErrV2PeerRequestTimedOut, wantMessage)
		assertErrV2InfoPeerRequestTimedOut(t, err.AddInfo(), peerAlias, timeout)
		assert.Contains(t, err.Message(), "proposing channel")
	})

	t.Run("chClient_proposeChannel_PeerRejected", func(t *testing.T) {
		reason := "some random reason"
		peerRejectedError := pclient.PeerRejectedError{
			ItemType: "channel proposal",
			Reason:   reason,
		}
		ch, _ := newMockPCh(t, validOpeningBalInfo)
		sess, chClient := newSessionWMockChClient(t, true, peerIDs...)
		chClient.On("Register", mock.Anything, mock.Anything).Return()
		chClient.On("ProposeChannel", mock.Anything, mock.Anything).Return(ch, peerRejectedError)

		_, err := sess.OpenCh(context.Background(), validOpeningBalInfo, app, 10)
		require.Error(t, err)

		wantMessage := peerRejectedError.Error()
		peerAlias := peerIDs[0].Alias // peer in validOpeningBal is peerIDs[0].
		assertAPIError(t, err, perun.ParticipantError, perun.ErrV2PeerRejected, wantMessage)
		assertErrV2InfoPeerRejected(t, err.AddInfo(), peerAlias, reason)
		assert.Contains(t, err.Message(), "proposing channel")
	})

	t.Run("chClient_proposeChannel_PeerNotFunded", func(t *testing.T) {
		var peerIdx uint16 = 1 // Index of peer (proposee) is always 1.
		// pointer to the error is used as go-perun returns this error as pointer.
		fundingTimeoutError := &pchannel.FundingTimeoutError{
			Errors: []*pchannel.AssetFundingError{{
				Asset:         pchannel.Index(0),
				TimedOutPeers: []pchannel.Index{peerIdx},
			}},
		}
		ch, _ := newMockPCh(t, validOpeningBalInfo)
		sess, chClient := newSessionWMockChClient(t, true, peerIDs...)
		chClient.On("Register", mock.Anything, mock.Anything).Return()
		chClient.On("ProposeChannel", mock.Anything, mock.Anything).Return(ch, fundingTimeoutError)

		_, err := sess.OpenCh(context.Background(), validOpeningBalInfo, app, 10)
		require.Error(t, err)

		wantMessage := fundingTimeoutError.Error()
		peerAlias := peerIDs[0].Alias // peer in validOpeningBal is peerIDs[0].
		assertAPIError(t, err, perun.ParticipantError, perun.ErrV2PeerNotFunded, wantMessage)
		assertErrV2InfoPeerNotFunded(t, err.AddInfo(), peerAlias)
		assert.Contains(t, err.Message(), "proposing channel")
	})

	t.Run("chClient_proposeChannel_FundingTxTimedOut", func(t *testing.T) {
		fundingTxTimedoutError := pclient.TxTimedoutError{
			TxType: pethchannel.Fund.String(),
			TxID:   "0xabcd",
		}
		ch, _ := newMockPCh(t, validOpeningBalInfo)
		sess, chClient := newSessionWMockChClient(t, true, peerIDs...)
		chClient.On("Register", mock.Anything, mock.Anything).Return()
		chClient.On("ProposeChannel", mock.Anything, mock.Anything).Return(ch, fundingTxTimedoutError)

		_, err := sess.OpenCh(context.Background(), validOpeningBalInfo, app, 10)
		require.Error(t, err)

		wantMessage := fundingTxTimedoutError.Error()
		assertAPIError(t, err, perun.ProtocolFatalError, perun.ErrV2TxTimedOut, wantMessage)
		txType := fundingTxTimedoutError.TxType
		txID := fundingTxTimedoutError.TxID
		txTimeout := ethereumtest.OnChainTxTimeout.String()
		assertErrV2InfoTxTimedOut(t, err.AddInfo(), txType, txID, txTimeout)
		assert.Contains(t, err.Message(), "proposing channel")
	})

	t.Run("chClient_proposeChannel_ChainNotReachable", func(t *testing.T) {
		chainURL := ethereumtest.ChainURL
		chainNotReachableError := pclient.ChainNotReachableError{}
		ch, _ := newMockPCh(t, validOpeningBalInfo)
		sess, chClient := newSessionWMockChClient(t, true, peerIDs...)
		chClient.On("Register", mock.Anything, mock.Anything).Return()
		chClient.On("ProposeChannel", mock.Anything, mock.Anything).Return(ch, chainNotReachableError)

		_, err := sess.OpenCh(context.Background(), validOpeningBalInfo, app, 10)
		require.Error(t, err)

		wantMessage := chainNotReachableError.Error()
		assertAPIError(t, err, perun.ProtocolFatalError, perun.ErrV2ChainNotReachable, wantMessage)
		assertErrV2InfoChainNotReachable(t, err.AddInfo(), chainURL)
		assert.Contains(t, err.Message(), "proposing channel")
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

	proposal, err := pclient.NewLedgerChannelProposal(10, ownAddr.OffChainAddr, allocation,
		[]pwire.Address{peer.OffChainAddr, ownAddr.OffChainAddr},
		pclient.WithApp(pchannel.NoApp(), pchannel.NoData()), pclient.WithRandomNonce())
	require.NoError(t, err)
	return proposal
}

func newSessionWChProposal(t *testing.T, peerIDs []perun.PeerID) (
	*session.Session, pclient.ChannelProposal, string) {
	session, _ := newSessionWMockChClient(t, true, peerIDs...)
	ownPeerID, err := session.GetPeerID(perun.OwnAlias)
	require.NoError(t, err)
	chProposal := newChProposal(t, ownPeerID, peerIDs[0])
	chProposalID := fmt.Sprintf("%x", chProposal.ProposalID())
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
	wantMessage := session.ErrSubAlreadyExists.Error()
	assertAPIError(t, err, perun.ClientError, perun.ErrV2ResourceExists, wantMessage)
	assertErrV2InfoResourceExists(t, err.AddInfo(), "subscription to channel proposals", openSession.ID())

	// == SubTest 3: Unsub successfully ==
	err = openSession.UnsubChProposals()
	require.NoError(t, err)

	// == SubTest 4: Unsub again, should error ==
	err = openSession.UnsubChProposals()
	require.Error(t, err)
	wantMessage = session.ErrNoActiveSub.Error()
	assertAPIError(t, err, perun.ClientError, perun.ErrV2ResourceNotFound, wantMessage)
	assertErrV2InfoResourceNotFound(t, err.AddInfo(), "subscription to channel proposals", openSession.ID())

	t.Run("Sub_sessionClosed", func(t *testing.T) {
		err = closedSession.SubChProposals(dummyNotifier)
		require.Error(t, err)
		wantMessage := session.ErrSessionClosed.Error()
		assertAPIError(t, err, perun.ClientError, perun.ErrV2FailedPreCondition, wantMessage)
		assert.Nil(t, err.AddInfo())
	})

	t.Run("Unsub_sessionClosed", func(t *testing.T) {
		err = closedSession.UnsubChProposals()
		require.Error(t, err)
		wantMessage := session.ErrSessionClosed.Error()
		assertAPIError(t, err, perun.ClientError, perun.ErrV2FailedPreCondition, wantMessage)
		assert.Nil(t, err.AddInfo())
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

	t.Run("happy_accept", func(t *testing.T) {
		session, chProposal, chProposalID := newSessionWChProposal(t, peerIDs)

		ch, _ := newMockPCh(t, openingBalInfo)
		responder := &mocks.ChProposalResponder{}
		responder.On("Accept", mock.Anything, mock.Anything).Return(ch, nil)
		session.HandleProposalWInterface(chProposal, responder)

		gotChInfo, err := session.RespondChProposal(context.Background(), chProposalID, true)
		require.NoError(t, err)
		assert.Equal(t, gotChInfo.ChID, fmt.Sprintf("%x", ch.ID()))
	})

	t.Run("happy_reject", func(t *testing.T) {
		session, chProposal, chProposalID := newSessionWChProposal(t, peerIDs)

		responder := &mocks.ChProposalResponder{}
		responder.On("Reject", mock.Anything, mock.Anything).Return(nil)
		session.HandleProposalWInterface(chProposal, responder)

		_, err := session.RespondChProposal(context.Background(), chProposalID, false)
		assert.NoError(t, err)
	})

	t.Run("respond_session_closed", func(t *testing.T) {
		sess, _ := newSessionWMockChClient(t, false, peerIDs...)

		chProposalID := "any-proposal-id" // A closed session returns error irrespective of proposal id.
		_, err := sess.RespondChProposal(context.Background(), chProposalID, true)
		require.Error(t, err)

		wantMessage := session.ErrSessionClosed.Error()
		assertAPIError(t, err, perun.ClientError, perun.ErrV2FailedPreCondition, wantMessage)
		assert.Nil(t, err.AddInfo())
	})

	t.Run("respond_unknown_proposalID", func(t *testing.T) {
		session, _ := newSessionWMockChClient(t, true, peerIDs...)

		unknownProposalID := "unknown-proposal-id"
		_, err := session.RespondChProposal(context.Background(), unknownProposalID, true)
		require.Error(t, err)

		assertAPIError(t, err, perun.ClientError, perun.ErrV2ResourceNotFound, "proposal")
		assertErrV2InfoResourceNotFound(t, err.AddInfo(), "proposal", unknownProposalID)
	})

	t.Run("response_timeout_expired", func(t *testing.T) {
		modifiedResponseTimeout := 1 * time.Second
		chClient := &mocks.ChClient{} // Dummy ChClient is sufficient as no methods on it will be invoked.
		prng := rand.New(rand.NewSource(ethereumtest.RandSeedForTestAccs))
		modifiedCfg := sessiontest.NewConfigT(t, prng, peerIDs...)
		modifiedCfg.ResponseTimeout = modifiedResponseTimeout
		session, err := session.NewSessionForTest(modifiedCfg, true, chClient)
		require.NoError(t, err)
		require.NotNil(t, session)

		ownPeerID, err := session.GetPeerID(perun.OwnAlias)
		require.NoError(t, err)
		chProposal := newChProposal(t, ownPeerID, peerIDs[0])
		chProposalID := fmt.Sprintf("%x", chProposal.ProposalID())
		responder := &mocks.ChProposalResponder{} // Dummy responder as no methods on it will be invoked.
		session.HandleProposalWInterface(chProposal, responder)
		time.Sleep(modifiedResponseTimeout + 1*time.Second) // Wait until the notification expires.
		_, apiErr := session.RespondChProposal(context.Background(), chProposalID, true)
		require.Error(t, apiErr)

		assertAPIError(t, apiErr, perun.ParticipantError, perun.ErrV2UserResponseTimedOut, "")
		assertErrV2InfoUserResponseTimedout(t, apiErr.AddInfo())
	})

	t.Run("respond_accept_AnError", func(t *testing.T) {
		session, chProposal, chProposalID := newSessionWChProposal(t, peerIDs)

		ch, _ := newMockPCh(t, openingBalInfo)
		responder := &mocks.ChProposalResponder{}
		responder.On("Accept", mock.Anything, mock.Anything).Return(ch, assert.AnError)
		session.HandleProposalWInterface(chProposal, responder)

		_, err := session.RespondChProposal(context.Background(), chProposalID, true)
		assertAPIError(t, err, perun.InternalError, perun.ErrV2UnknownInternal, assert.AnError.Error())
	})

	t.Run("respond_reject_AnError", func(t *testing.T) {
		session, chProposal, chProposalID := newSessionWChProposal(t, peerIDs)

		responder := &mocks.ChProposalResponder{}
		responder.On("Reject", mock.Anything, mock.Anything).Return(assert.AnError)
		session.HandleProposalWInterface(chProposal, responder)

		_, err := session.RespondChProposal(context.Background(), chProposalID, false)
		assertAPIError(t, err, perun.InternalError, perun.ErrV2UnknownInternal, assert.AnError.Error())
	})

	t.Run("respond_accept_PeerNotFunded", func(t *testing.T) {
		var peerIdx uint16 = 0 // Index of peer (proposer) is always 0.
		// pointer to the error is used as go-perun returns this error as pointer.
		fundingTimeoutError := &pchannel.FundingTimeoutError{
			Errors: []*pchannel.AssetFundingError{{
				Asset:         pchannel.Index(0),
				TimedOutPeers: []pchannel.Index{peerIdx},
			}},
		}
		session, chProposal, chProposalID := newSessionWChProposal(t, peerIDs)

		ch, _ := newMockPCh(t, openingBalInfo)
		responder := &mocks.ChProposalResponder{}
		responder.On("Accept", mock.Anything, mock.Anything).Return(ch, fundingTimeoutError)
		session.HandleProposalWInterface(chProposal, responder)

		_, err := session.RespondChProposal(context.Background(), chProposalID, true)
		peerAlias := peerIDs[0].Alias // peer in validOpeningBal is peerIDs[0].
		assertAPIError(t, err, perun.ParticipantError, perun.ErrV2PeerNotFunded, fundingTimeoutError.Error())
		assertErrV2InfoPeerNotFunded(t, err.AddInfo(), peerAlias)
	})

	t.Run("respond_accept_FundingTxTimedOut", func(t *testing.T) {
		fundingTxTimedoutError := pclient.TxTimedoutError{
			TxType: pethchannel.Fund.String(),
			TxID:   "0xabcd",
		}
		session, chProposal, chProposalID := newSessionWChProposal(t, peerIDs)

		ch, _ := newMockPCh(t, openingBalInfo)
		responder := &mocks.ChProposalResponder{}
		responder.On("Accept", mock.Anything, mock.Anything).Return(ch, fundingTxTimedoutError)
		session.HandleProposalWInterface(chProposal, responder)

		_, err := session.RespondChProposal(context.Background(), chProposalID, true)
		assertAPIError(t, err, perun.ProtocolFatalError, perun.ErrV2TxTimedOut, fundingTxTimedoutError.Error())
		txType := fundingTxTimedoutError.TxType
		txID := fundingTxTimedoutError.TxID
		txTimeout := ethereumtest.OnChainTxTimeout.String()
		assertErrV2InfoTxTimedOut(t, err.AddInfo(), txType, txID, txTimeout)
	})

	t.Run("respond_accept_ChainNotReachable", func(t *testing.T) {
		chainURL := ethereumtest.ChainURL
		chainNotReachableError := pclient.ChainNotReachableError{}
		session, chProposal, chProposalID := newSessionWChProposal(t, peerIDs)

		ch, _ := newMockPCh(t, openingBalInfo)
		responder := &mocks.ChProposalResponder{}
		responder.On("Accept", mock.Anything, mock.Anything).Return(ch, chainNotReachableError)
		session.HandleProposalWInterface(chProposal, responder)

		_, err := session.RespondChProposal(context.Background(), chProposalID, true)
		wantMessage := chainNotReachableError.Error()
		assertAPIError(t, err, perun.ProtocolFatalError, perun.ErrV2ChainNotReachable, wantMessage)
		assertErrV2InfoChainNotReachable(t, err.AddInfo(), chainURL)
	})
}

func Test_ProposeCh_GetChsInfo(t *testing.T) {
	peerIDs := newPeerIDs(t, uint(2))
	setupSession := func() (perun.SessionAPI, *mocks.ChClient) {
		prng := rand.New(rand.NewSource(ethereumtest.RandSeedForTestAccs))
		cfg := sessiontest.NewConfigT(t, prng, peerIDs...)
		chClient := &mocks.ChClient{}
		session, err := session.NewSessionForTest(cfg, true, chClient)
		require.NoError(t, err)
		require.NotNil(t, session)
		return session, chClient
	}

	proposeCh := func(session perun.SessionAPI, chClient *mocks.ChClient) string {
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
		chClient.On("ProposeChannel", mock.Anything, mock.Anything).Return(ch, nil).Once()
		chClient.On("Register", mock.Anything, mock.Anything).Return().Once()

		chInfo, err := session.OpenCh(context.Background(), validOpeningBalInfo, app, 10)
		require.NoError(t, err)
		require.NotZero(t, chInfo)
		return chInfo.ChID
	}

	t.Run("happy_one_channel", func(t *testing.T) {
		sess, chClient := setupSession()
		chID := proposeCh(sess, chClient)
		chsInfo := sess.GetChsInfo()
		assert.Len(t, chsInfo, 1)
		assert.Equal(t, chsInfo[0].ChID, chID)
	})

	t.Run("happy_many_channels_ordered", func(t *testing.T) {
		sess, chClient := setupSession()
		cntChsToOpen := 20
		chIDs := make([]string, cntChsToOpen)
		for i := 0; i < cntChsToOpen; i++ {
			chIDs[i] = proposeCh(sess, chClient)
		}

		getChIDs := func() []string {
			chsInfo := sess.GetChsInfo()
			gotChIDs := make([]string, len(chsInfo))
			for i := range chsInfo {
				gotChIDs[i] = chsInfo[i].ChID
			}
			return gotChIDs
		}
		gotChIDs1 := getChIDs()
		gotChIDs2 := getChIDs()
		gotChIDs3 := getChIDs()

		assert.Equal(t, chIDs, gotChIDs1)
		assert.Equal(t, chIDs, gotChIDs2)
		assert.Equal(t, chIDs, gotChIDs3)
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
	peerIDs := newPeerIDs(t, uint(2))
	validOpeningBalInfo := perun.BalInfo{
		Currency: currency.ETH,
		Parts:    []string{perun.OwnAlias, peerIDs[0].Alias},
		Bal:      []string{"1", "2"},
	}
	updatedBalInfo := validOpeningBalInfo
	updatedBalInfo.Bal = []string{"0.5", "2.5"}

	currState := makeState(t, validOpeningBalInfo, false)
	chUpdate := &pclient.ChannelUpdate{
		State: makeState(t, updatedBalInfo, false),
	}

	t.Run("happy", func(t *testing.T) {
		session, _ := newSessionWMockChClient(t, true)
		responder := &mocks.ChUpdateResponder{}
		responder.On("Reject", mock.Anything, mock.Anything).Return(nil)
		session.HandleUpdateWInterface(currState, *chUpdate, responder)
	})
	t.Run("session_closed", func(t *testing.T) {
		session, _ := newSessionWMockChClient(t, false)
		session.HandleUpdate(currState, pclient.ChannelUpdate{}, new(pclient.UpdateResponder))
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

func assertAPIError(t *testing.T, e perun.APIErrorV2, category perun.ErrorCategory, code perun.ErrorCode, msg string) {
	t.Helper()

	assert.Equal(t, category, e.Category())
	assert.Equal(t, code, e.Code())
	assert.Contains(t, e.Message(), msg)
}

func assertErrV2InfoPeerRequestTimedOut(t *testing.T, info interface{}, peerAlias, timeout string) {
	t.Helper()

	addInfo, ok := info.(perun.ErrV2InfoPeerRequestTimedOut)
	require.True(t, ok)
	assert.Equal(t, peerAlias, addInfo.PeerAlias)
	assert.Equal(t, timeout, addInfo.Timeout)
}

func assertErrV2InfoPeerRejected(t *testing.T, info interface{}, peerAlias, reason string) {
	t.Helper()

	addInfo, ok := info.(perun.ErrV2InfoPeerRejected)
	require.True(t, ok)
	assert.Equal(t, peerAlias, addInfo.PeerAlias)
	assert.Equal(t, reason, addInfo.Reason)
}

func assertErrV2InfoPeerNotFunded(t *testing.T, info interface{}, peerAlias string) {
	t.Helper()

	addInfo, ok := info.(perun.ErrV2InfoPeerNotFunded)
	require.True(t, ok)
	assert.Equal(t, peerAlias, addInfo.PeerAlias)
}

func assertErrV2InfoUserResponseTimedout(t *testing.T, info interface{}) {
	t.Helper()

	addInfo, ok := info.(perun.ErrV2InfoUserResponseTimedOut)
	require.True(t, ok)
	assert.Less(t, addInfo.Expiry, time.Now().Unix())
}

func assertErrV2InfoResourceNotFound(t *testing.T, info interface{}, resourceType, resourceID string) {
	t.Helper()

	addInfo, ok := info.(perun.ErrV2InfoResourceNotFound)
	require.True(t, ok)
	assert.Equal(t, resourceType, addInfo.Type)
	assert.Equal(t, resourceID, addInfo.ID)
}

func assertErrV2InfoResourceExists(t *testing.T, info interface{}, resourceType, resourceID string) {
	t.Helper()

	addInfo, ok := info.(perun.ErrV2InfoResourceExists)
	require.True(t, ok)
	assert.Equal(t, resourceType, addInfo.Type)
	assert.Equal(t, resourceID, addInfo.ID)
}

func assertErrV2InfoInvalidArgument(t *testing.T, info interface{}, name, value string) {
	t.Helper()

	addInfo, ok := info.(perun.ErrV2InfoInvalidArgument)
	require.True(t, ok)
	assert.Equal(t, name, addInfo.Name)
	assert.Equal(t, value, addInfo.Value)
	t.Log("requirement:", addInfo.Requirement)
}

func assertErrV2InfoTxTimedOut(t *testing.T, info interface{}, txType, txID, txTimeout string) {
	t.Helper()

	addInfo, ok := info.(perun.ErrV2InfoTxTimedOut)
	require.True(t, ok)
	assert.Equal(t, txType, addInfo.TxType)
	assert.Equal(t, txID, addInfo.TxID)
	assert.Equal(t, txTimeout, addInfo.TxTimeout)
}

func assertErrV2InfoChainNotReachable(t *testing.T, info interface{}, chainURL string) {
	t.Helper()

	addInfo, ok := info.(perun.ErrV2InfoChainNotReachable)
	require.True(t, ok)
	assert.Equal(t, chainURL, addInfo.ChainURL)
}
