# Perun Node - Go implementation [![join the chat][rocketchat-image]][rocketchat-url]

[rocketchat-url]: https://chat.hyperledger.org/channel/perun
[rocketchat-image]: https://open.rocket.chat/images/join-chat.svg

| Develop | Master |
| :----: | :-----: |
| [![CircleCI](https://circleci.com/gh/hyperledger-labs/perun-node/tree/develop.svg?style=shield)](https://circleci.com/gh/hyperledger-labs/perun-node/tree/develop) | [![CircleCI](https://circleci.com/gh/hyperledger-labs/perun-node/tree/master.svg?style=shield)](https://circleci.com/gh/hyperledger-labs/perun-node/tree/master) |

Perun is an open source project that aims to increase blockchain transaction
throughput by using just a handful of main chain transactions to move an entire
peer-to-peer network of activity off the main chain.  After an initial setup of
a set of basic transaction channels, this network lets any participant transact
with any other participant via virtual channels which do not require additional
on-chain setup.  We do this by implementing the 
[Perun protocol](https://perun.network/), which has been formally proven to
allow for secure off-chain transactions.


## Project Status

At the moment the perun-node is neither ready for production nor does it
implement the complete perun protocol yet. But with basic features available,
the project is at a stage where you can try it out and start to get involved.

This is a complete re-implementation of the previous version (available under
the previous name of the project: dst-go) in branch
[legacy/master](https://github.com/hyperledger-labs/perun-node/tree/legacy/master).
This version builds on top of the [go-perun](https://github.com/hyperledger-labs/go-perun) SDK that implements a state
channel client based on perun protocol. See Description for more details.

## Description

The perun-node is multi-user state channel node that can be used for opening,
transacting on and closing state channels. It builds on the state channel
client implemented by
[go-perun](https://github.com/hyperledger-labs/go-perun) and implements the
following functionalities:

1. Payment App: For using perun protocol to establish and use bi-directional
payment channels.
2. ID Provider: For the user to define a list of known participants in
the off-chain network.
3. Key management: For managing the cryptographic keys of the user.
4. User session: For allowing multiple users to use a single node, each with
a dedicated key manager and ID provider.
5. User API interface: For the user to interact with the perun-node.

The current version provides the following features:

|Feature | Implementation |
|:--|:--|
|Blockchain Backend|Ethereum|
|Key management|Ethereum keystore |
|ID Provider|Local |
|User API|Two Party Payment API |
|User API Adapter|gRPC |
|Persistence|LevelDB|

This project currently contains two executable packages located in the `cmd` directory.

- `perunnode`: An app for starting a running instance of perun node. It can
  also generate configuration artifacts for trying out the node.

- `perunnodecli` is an app with interactive CLI interface that serves two purposes:
    - easy way to try out payment channel API.
    - reference implementation for using the generated grpc client stubs for
      payment channel API.

For detailed information on the features offered by these two applications and
steps on how to try them out, see the
[tutorial section](https://labs.hyperledger.org/perun-doc/node/introduction.html#user-guide)
on the project documentation website.

## Getting Started

Install the following pre-requisites.

    1. Go (v1.14 or later).
    2. ganache-cli (v6.9.1 or later).

Clone the project and sync the dependencies:

```bash
git clone https://github.com/hyperledger-labs/perun-node.git
cd perun-node
go mod tidy
```

Start the ganache-cli node for running integration tests:

```bash
# These funded accounts will be used in tests. "-b 1" configures the
# ganache-cli node to mine one block every second. It is required as our
# contracts use blockchain based timeout for settling a state channel on-chain.
ganache-cli -b 1 \
--account="0x1fedd636dbc7e8d41a0622a2040b86fea8842cef9d4aa4c582aad00465b7acff,100000000000000000000" \
--account="0xb0309c60b4622d3071fad3e16c2ce4d0b1e7758316c187754f4dd0cfb44ceb33,100000000000000000000"
```

Run the linter and tests from the project root directory:

```bash
# Lint
golangci-lint run ./...

# Test
go test -tags=integration -count=1 ./...

# Build peurnnode and perunnodecli binaries
make
```


## License

perun-node is open-sourced under the Apache-2.0 license. See the
[LICENSE](LICENSE) file for details.
