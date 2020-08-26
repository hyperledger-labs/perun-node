# Glossary

#### Addresses and Accounts

| | |
| :--| :--|
| On-chain Account|Account used for signing on-chain transactions and funding the ledger channels.|
| Off-chain Account| Account used for signing off-chain transactions and authentication messages in the off-chain network .|
| On-chain Address |Address corresponding to the on-chain account.|
| Off-chain Address | Address corresponding to the off-chain account. This is the permanent identity of a participant in the off-chain network.|
| Participant Address | Address used for representing a user in a channel. Currently, the off-chain address is reused as the participant address.|
|Comm Address|Address for establishing physical network connection for off-chain communication.|
|Comm Type|Type of physical network connection protocol for off-chain communication.|
|Alias| A user-friendly used for referring to a peer in user API calls. It is assigned by the user in his/her contacts file. |
| Adjudicator Address | Address of the adjudicator contract. (See Configuration and Timeouts.) |
| Asset Address | Address of the asset contract. (See Configuration and Timeouts.) |

#### IDs

| | |
| :--| :--|
|Session ID|Uniquely identifies a session on a perun node. It is derived by combining user id with some random data.|
|Proposal ID|Uniquely identifies a channel proposal. It is derived from the parameters in the channel proposal.|
|Channel ID|Uniquely identifies a channel in the off-chain network. It is derived from the parameters of a channel.|
|Update ID|Uniquely identifies a channel update. It is derived from a combination of the channel ID and its current version.|

#### Configuration and Timeouts

| | |
| :--| :--|
|Chain URL|Address of the blockchain node used by the perun node for sending and receiving on-chain transactions.|
| Adjudicator contract | The on-chain smart contract used for resolving disputes over the channel state, and handles closing of channels.|
| Asset contract |The on-chain smart contract used by adjudicator smart contract and the Perun SDK to interact with a certain type of asset/currency when funding a channel or withdrawing funds from it.
| Chain connection timeout | The blockchain node is considered unreachable if a connection could not be established within this timeout.|
| On-chain transaction timeout | The maximum time to wait for an on-chain transaction to be considered final. It this expires, the transaction is considered as failed.|
| Response timeout | The maximum time to wait for receiving a response to an off-chain message such as response for channel proposal or state update. If this expires, the peer is considered as unreachable. For proper operation of the off-chain network, all perun nodes should use the same response timeout.|
| Expiry | Timeout specified in each notification to the user. The user should respond to the notification before this timeout expires.|
| Challenge Duration |The period within which any participant must refute (with the latest state), if he/she finds a registered state is not the latest. The refute is automatically done by the perun node. It is a specified during a channel proposal and agreed upon by all participants.|

#### Channel Life cycle

| | |
| :--| :--|
| Open/Propose | Proposes the parameters for a new channel to the involved participants. If all participants accept the proposal, the channel is funded on-chain. |
| Update | Progress the state of an open channel to a new version. This is done by sending an update proposal to all participants and getting it accepted by all of them.|
| Close | Finalize the state of an open channel, register it on the blockchain and withdraw funds to the user's on-chain account.|

#### Channel Parameters


| | |
| :--| :--|
|Balances|Balance of each participant in the channel. It can be change during each update, but the sum of balances should always be the same.|
|Final|Final indicates if this is a final update. Once a final update it accepted by all participants, no further updates are possible.|
|Version|Version indicates the current version of the state in the channel. The initial state has version 0. All state updates increase the version by 1.|
