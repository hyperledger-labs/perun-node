module.exports = {
    plugins: ["truffle-security", "solidity-coverage"],

    mocha: {
        reporter: 'eth-gas-reporter',
        reporterOptions: {
            // See https://www.npmjs.com/package/eth-gas-reporter
            // gasPrice: 300,
            onlyCalledMethods: false
        }
    },

    compilers: {
        solc: {
            version: "^0.7.0",
            settings: {
                optimizer: {
                    enabled: true,
                    runs: 200,
                },
            },
        }
    }
}
