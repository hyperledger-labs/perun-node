# Trying out Perun Node and Perun Node CLI


#### perunnode

An application for starting a running version of perun node that
serves payment channel API on using grpc protocol. Any application can connect
with the perunnode by directly integrating the grpc client stubs in their
project. The proto file for generating the grpc stub can be found at
`api/grpc/pb/api.proto`.

#### perunnodecli

An application with interactive CLI interface that uses generated grpc client
stubs to interact with a running instance of perun node. Besides this, it does
not share any code with rest of the project.

It provides different sets of commands: chain, node, session, contact, channel,
payment. Each of these commands have sub-commands to do specific operations.
In the app, type `help` to get information on each of these commands.

The chain command provides options to directly interact with blockchain such
as for deploying contracts, while the rest of the commands are for accessing
the API of the node.


## Pre-requisites

To try out the perunnode-cli app on your machine, you would require the
following things to be running:

1. A running blockchain node with perun contracts deployed on it.
2. Running perunnode connected to the blockchain node.
3. Configuration files for session, in a path accessible by perunnode.

Follow the below steps to setup the pre-requisites:

1. Start a blockchain node:

```
ganache-cli -b 1 --account="0x1fedd636dbc7e8d41a0622a2040b86fea8842cef9d4aa4c582aad00465b7acff,100000000000000000000" --account="0xb0309c60b4622d3071fad3e16c2ce4d0b1e7758316c187754f4dd0cfb44ceb33,100000000000000000000"
```

Once the ganache-cli is started, you will see a line 
"Listening on 127.0.0.1:8545" with a blinking cursor. Leave this running in
this terminal.

2. In another terminal, clone the project in any location, build perunnode and
   perunnode-cli:

```
git clone https://github.com/hyperledger-labs/perun-node.git

cd perun-node

make
```

3. Generate configuration artifacts for demo artifacts for node and session
   configuration:

```
./perunnode generate
```

This will generate the following artifacts:

- Node: node.yaml file.
- Session: Two directories (alice and bob) each containing session.yaml file,
  idprovider.yaml file and keystore directory with keys corresponding to the
  on-chain and off-chain accounts.


4. Set the blockchain address. Besides of commands to use API of perunnode,
   this application also has a few helper commands to directly interact with
   blockchain that are organized as sub-commands of chain command. To use these
   commands, you need to first set the address of the blockchain node:

```
chain set-blockchain-address ws://127.0.0.1:8545
```

5. Deploy perun contracts on the newly started `ganache-cli` node using the
   chain command.

```
chain deploy-perun-contracts
```

6. Reading on-chain balance. For reading the on-chain balance, you need to
   specify the address as a hex string with a prefix of '0x'. For convenience,
   the on-chain addresses that were funded when starting the ganache-cli node as
   described in step 1 are already added to autocomplete. However, this command
   can also be used to read on-chain balance for other addresses.

First address is that of alice and second is that of bob. This can be verified
by looking at the `alice/session.yaml` and `bob/session.yaml` files.

```
chain get-on-chain-balance 0x8450c0055cB180C7C37A25866132A740b812937B
chain get-on-chain-balance 0xc4bA4815c82727554e4c12A07a139b74c6742322
```

You can use these commands at any time before opening, while open or after
closing a payment channel.

7. Start the perunnode:

```
./perunnode run
```

This will start the perunnode using the config file located at default path
`./node.yaml` that was generated in step 3. You will see a line
"Serving payment channel API via grpc at port :50001" with a blinking cursor.
Leave this running in this terminal.

Now all the pre-requisites for `perunnode-cli` are setup.

## Opening a session, opening channel within it, making payments & closing it

1. Open two new terminals side by side, one each for alice and bob roles
   respectively. In both the terminal, start the perunnode-cli app using below
command:

```
./perunnodecli
```

This will bring up an interactive shell with auto-completion support. Type
`help` to see a list of commands and their help message. Typing one of those
commands without any arguments will print the help message for that command,
including the list of sub-commands.

2. In any one of the terminals, deploy perun contracts on the ganache-cli node
   using the below commands. Just a reminder for one last time, you can almost
   get every value by using auto-completion (by pressing TAB) and get away
   without typing.

```
chain deploy-perun-contracts ws://127.0.0.1:8545
```

From here on, choose one terminal for alice role and one for bob role. In each
step, the role will be the enclosed in square brackets before description.

3. Opening a session and reading contact.

- (a) [Alice] Start the session and get the contact of bob to check if it is
   present. Getting the contact will also add the peer alias to auto-completion
   list. The alias will then suggested, wherever a peer alias is expected. Two
   exceptions where peer alias is not auto-completed are `contact add` and
   `contact get` commands, because these commands are designed to add/get
    contacts for unknown aliases.

```
# [Alice]
node connect :50001
session open alice/session.yaml
contact get bob
```

- (b) [Bob] Repeat step 3 for bob using below commands:


```
# [Bob]
node connect :50001
session open bob/session.yaml
contact get alice
```

4. Sending a request to open a payment channel and accepting it.

- (a) [Alice] Send a request to open a channel with bob:

```
# [Alice]
channel send-opening-request bob 1 2
```

- (b) [Bob] Receives a channel opening request notification that includes
  request ID. Type the command to accept the channel opening request directly after receiving the notification:

```
# [Bob]
channel accept-opening-request request_1_alice
```

- Once successfully accepted, information on the opened channel is printed in
  both terminals.

5. Listing out open channels. In any of the terminals, type the below command
   to see the list of open channels:

```
channel list-open-channels
```

6. Sending a request to open a payment channel and rejecting it.

- (a) [Bob] Send a request to open a channel with bob:

```
# [Bob]
channel send-opening-request alice 3 4
```

- (b) [Alice] Receives a channel opening request notification that includes
  request ID. Reject it:

```
# [Alice]
channel reject-opening-request request_1_bob
```

- Once successfully accepted, information on the opened channel is printed in
  both terminals.

7. Sending a payment on the open channel and accepting it.

- (a) [Alice] Send a payment to bob on an open channel:

```
# [Alice]
payment send-to-peer ch_1_bob 0.1
```

- (b) [Bob] Receives a payment notification that includes the channel alias.
  Accept it:

```
# [Bob]
payment accept-payment-update-from-peer ch_1_alice
```

Once payment is accepted, the updated information is printed on both terminals.

8. Sending a payment on the open channel and rejecting it.

- (a) [Bob] Send a payment to bob on an open channel:

```
# [Bob]
payment send-to-peer ch_1_alice 0.2
```

- (b) [Alice] Receives a payment notification that includes the channel alias.
  Reject it:

```
# [Alice]
payment reject-payment-update-from-peer ch_1_bob
```
Once payment is rejected, green message is printed on alice terminal for
successfully rejecting the payment. Red error message is printed on bob
terminal as the payment was rejected by user.

9. Try to close the session will return error when there are open channels. Run
   the below command in any or both of the terminals and they should return an
   error.

```
session close no-force
```

10. Close the channel.

- (a) [Alice] Close the channel with the below command.

```
# [Alice]
channel close-n-settle-on-chain ch_1_bob
```

- (b) [Bob] Receives a finalizing update when alice sends close command. This
  is to finalize the channel off-chain, so that it can be collaboratively
  closed on the blockchain without waiting for challenge duration to expire.
  However, due to an issue (that will be fixed in next updated), the
  collaborative close will not work as expected. So reject the finalizing
  update:

```
# [Bob]
payment reject-payment-update-from-peer ch_1_alice
```

Now the program will opt for non-collaborative close by registering the state
on the blockchain, waiting for the challenge duration to expire and then
withdrawing the funds.

Even if Bob doesn't respond, Alice's request will wait until response timeout
expires (in this demo it is 10s) and then proceed with non-collaborative
close. Bob's node on the other hand will receive a notification when the
channel is finalized on the blockchain and funds will be withdrawn
automatically. A channel closed notification will be printed.

11. Close the session:

Since the open channels are closed, the session can be closed with the same
command as in step 8, but without any error.

```
# [Alice]
session close no-force
```

```
# [Bob]
session close no-force
```

12. To try out persistence of channels:

- (a) Open a session for alice, bob and then open a few channels using commands
  described in step 4.

- (b) Close the session using force option:

```
# [Alice]
session close force
```

```
# [Bob]
session close force
```

- (c) Open sessions for alice and bob again using the commands in step 3. Once
  the session is opened, the channels restored from persistence will be printed
  along with their aliases. You can send payments on these channels and close
  them as before. There is no difference between a channel opened in current
  session and a channel restored from persistence.


## Remarks

- You can try to open as many channels as desired using the commands as
  described in step 5. Each channel is addressed by its alias (that will be
  suggested in auto-complete).

- You can also try and send as many payments as desired using the commands as
  described in step 7. However, whenever a new payment notification is
  received, the previous one is automatically dropped. This however, is not a
  feature of payment channel API, where you can respond to any of the
  notifications as long as they have not expired. It was just a feature in the
  perunnode-cli app to make it simpler.

- The purpose of the perunnode-cli software is to demo the payment channel API
  and also as a reference implementation for using the grpc client stubs.
