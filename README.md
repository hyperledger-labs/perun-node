# Direct State Transfer - Go implementation (DST-GO)

| Develop | Master |
| :----: | :-----: |
| [![CircleCI](https://circleci.com/gh/direct-state-transfer/dst-go/tree/develop.svg?style=shield)](https://circleci.com/gh/direct-state-transfer/dst-go/tree/develop) | [![CircleCI](https://circleci.com/gh/direct-state-transfer/dst-go/tree/master.svg?style=shield)](https://circleci.com/gh/direct-state-transfer/dst-go/tree/master) |

Direct State Transfer (DST) is an open source project that aims to
increase blockchain transaction throughput by using just a handful of
main chain transactions to move an entire peer-to-peer network of
activity off the main chain.  After an initial setup of a set of basic
transaction channels, this network lets any participant transact with
any other participant via virtual channels which do not require
additional on-chain setup.  We do this by implementing the [Perun
protocol](https://perun.network/), which has been formally proven to
allow for secure off-chain transactions.

Please refer to the [Developer
Guide](https://github.com/direct-state-transfer/dst-doc/blob/master/source/developer_guide.rst)
and the rest of the
**[dst-doc](https://github.com/direct-state-transfer/dst-doc)** project
for more details on the architecture and currently implemented features
of **dst-go**, the [Go](https://golang.org/) implementation of the
Direct State Transfer project.

## Project Status

At the moment dst-go is neither ready for production use nor does it
implement the complete Perun protocol yet. But with the basic features
available, the project is at a stage where you can try it out and start
to get involved.

## Getting Started

This below document explains how to start using direct state transfer software.

### Prerequisites

1. Install Go (v1.10 or later)
2. Install Geth (v1.8.20-stable) (Optional; required only for running walkthrough)

### Build from source

The following commands can be used to build and install dst-go from its source.
Once the GOPATH and GOBIN are properly set in the local environment, following commands can be executed.

```bash
cd $GOPATH
cd src

#create directory
mkdir -p github.com/direct-state-transfer
cd github.com/direct-state-transfer

#clone dst-go repo
git clone https://github.com/direct-state-transfer/dst-go.git
cd dst-go

#install govendor tool to sync vendored dependencies
go get -u -v github.com/kardianos/govendor

#sync dependencies
#should be run inside the cloned root of the dst-go repo (~/PATH/dst-go)
govendor sync -v

#Now navigate to dst-go module inside the repo
cd dst-go
#install dst-go
go install -v
```

dst-go is installed in the local machine and the binary will be available at the set GOBIN path.

### Build using Make

From the local workspace run the following command to clone the dst-go project.

```bash
git clone https://github.com/direct-state-transfer/dst-go.git
```

The available make file can be used to build dst-go from source.
The following commands can be executed from the root repository of the project.
(~/LOCAL_WORKSPACE/dst-go/)

```bash
#To get a list of available build targets
make help
#To fetch dependencies and install dst-go
make install
```

On successful run, the binary will be available at ~/LOCAL_WORKSPACE/dst-go/build/workspace_/bin

### Build and run walkthrough

The walkthrough runs a complete sample sequence showing how a state channel is intialized and transactions are made outside blockchain using the Perun protocol.
This sequence contains the following steps:

* Initializing a state channel between two users by blocking Ether in Perun smart contracts
* Transfer of 10 Ethers one by one from one account to the other
* Finally settles it by updating the balances in Blockchain network once the transactions are done.

This can be run with a geth node (real backend) or with the simulated backend (from go-ethereum project).

**If real backend is used**, the following points need to be considered.

1. The geth node should be configured to use port number 8546 for
   websocket connection or the geth nodes websocket port number should
   be updated in `~/LOCAL_WORKSPACE/dst-go/testdata/test_addresses.json`
   in the place of ethereum node url.

    The following shows the default test_addresses.json with some sample keys already present in the keystore which is already present in the project. Similarly the ports and ethereum addresses of alice and bob can be updated.

    ```json
    {
        "ethereum_node_url" : "ws://localhost:8546",
        "alice_password"   : "",
        "bob_password"     : "",

        "alice_id" : {
            "on_chain_id": "0x932a74da117eb9288ea759487360cd700e7777e1",
            "listener_ip_addr": "localhost:9605",
            "listener_endpoint":"/"
        },
        "bob_id" :{
            "on_chain_id": "0x815430d6ea7275317d09199a5a5675f017e011ef",
            "listener_ip_addr": "localhost:9604",
            "listener_endpoint":"/"
        }
    }
    ```

2. The key files of the keys mentioned in the test_addresses.json for alice and bob should be present in testdata/test-keystore and the geth's keystore.
3. Sample configuration is available at testdata with default keys and key files. To use these, simply add the key files from testdata/test-keystore directory to geth's keystore.
4. Both Alice's and Bob's account should have minimum of 10 Ethers each to run this walkthrough. (It is currently not yet tested on mainnet.)
5. If a local geth node is used to run the walkthrough, the mining should be activated to execute transactions.
6. Make sure that the vendored dependencies using govendor tool are synchronized.

#### Build and run from source

Use the below commands to build and run the walkthrough from source.

```bash
cd $GOPATH/src/github.com/direct-state-transfer/dst-go/walkthrough
#build walkthrough
go build -v

#Run walkthrough
#Initialize bob's node (should be started first) and let it run..
./walkthrough --real_backend_bob

#Open a new terminal and go to walkthrough dir
cd $GOPATH/src/github.com/direct-state-transfer/dst-go/walkthrough

#Initialize alice's node and commence walkthrough
./walkthrough --real_backend_alice
```

Now the complete transaction sequence between two parties will be run and displayed in the terminal. Run `walkthrough -h` to see the available options.

#### Using Make

The following make commands are available to run the walkthrough sequence.

```bash
#To run walkthrough with real backend
make runWalkthrough BUILDOPTS="--real_backend"

#To run with simulated backend
#This will work with the default configuration in testdata
make runWalkthrough BUILDOPTS="--simulated_backend"

#Multiple options can also be passed as shown below.
make runWalkthrough BUILDOPTS="--simulated_backend --dispute --ch_message_print"
```

### Testing

#### Run the tests of all packages

In test mode short, all the unit tests will be performed using the simulated backend.

```bash
# To run all the tests
make test
# All options supported by go test command can be passed via testOpts
# option as shown below.
# Some flags will be enabled by default (-cover and cache=1). Output is
# non verbose by default.
make test BUILDOPTS="-v -short"
```

#### Performing lint

Run the below command from the root of the project to perform lint on the project.

```bash
make lint
```

## License

dst-go is open-sourced under the Apache-2.0 license. See the
[LICENSE](LICENSE) file for details.
