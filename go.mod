module github.com/hyperledger-labs/perun-node

go 1.14

replace perun.network/go-perun => ./go-perun

require (
	github.com/abiosoft/ishell v2.0.0+incompatible
	github.com/abiosoft/readline v0.0.0-20180607040430-155bce2042db // indirect
	github.com/ethereum/go-ethereum v1.10.1
	github.com/fatih/color v1.7.0
	github.com/flynn-archive/go-shlex v0.0.0-20150515145356-3f9db97f8568 // indirect
	github.com/gdamore/tcell/v2 v2.3.3 // indirect
	github.com/golang/protobuf v1.4.3
	github.com/gorilla/websocket v1.4.2
	github.com/kylelemons/godebug v1.1.0
	github.com/lucasb-eyer/go-colorful v1.2.0 // indirect
	github.com/magiconair/properties v1.8.4 // indirect
	github.com/mattn/go-colorable v0.1.7 // indirect
	github.com/mitchellh/mapstructure v1.3.3 // indirect
	github.com/mum4k/termdash v0.16.0
	github.com/otiai10/copy v1.2.0
	github.com/pelletier/go-toml v1.8.1 // indirect
	github.com/phayes/freeport v0.0.0-20180830031419-95f893ade6f2
	github.com/pkg/errors v0.9.1
	github.com/rivo/uniseg v0.2.0 // indirect
	github.com/shopspring/decimal v1.2.0
	github.com/sirupsen/logrus v1.8.1
	github.com/spf13/afero v1.4.1 // indirect
	github.com/spf13/cast v1.3.1 // indirect
	github.com/spf13/cobra v1.1.1
	github.com/spf13/jwalterweatherman v1.1.0 // indirect
	github.com/spf13/pflag v1.0.5
	github.com/spf13/viper v1.7.1
	github.com/stretchr/testify v1.7.0
	golang.org/x/net v0.0.0-20210226172049-e18ecbb05110
	golang.org/x/term v0.0.0-20210220032956-6a3ed077a48d // indirect
	google.golang.org/grpc v1.36.0
	google.golang.org/protobuf v1.25.0
	gopkg.in/ini.v1 v1.62.0 // indirect
	gopkg.in/yaml.v3 v3.0.0-20210107192922-496545a6307b
	gotest.tools v2.2.0+incompatible
	perun.network/go-perun v0.6.1-0.20210630194328-c1cbad083c2b
)
