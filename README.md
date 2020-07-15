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


## Project Status

A first version of dst-go is available in branch
[legacy/master](https://github.com/direct-state-transfer/dst-go/tree/legacy/master).
It is neither ready for production use nor does it implement the complete
Perun protocol yet. But with the basic features available it is at a stage
where you could try it out.

Now dst-go will be re-implemented from scratch building upon the
[go-perun](https://github.com/direct-state-transfer/go-perun). This is
happening on new
[master](https://github.com/direct-state-transfer/dst-go/tree/master) and
[develop](https://github.com/direct-state-transfer/dst-go/tree/develop)
branches.


## License

dst-go is open-sourced under the Apache-2.0 license. See the
[LICENSE](LICENSE) file for details.
