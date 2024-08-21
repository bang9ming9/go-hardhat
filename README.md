### build
```bash
go build -o /usr/local/bin/bms . # bms: bang9ming9 solidity
```

example
```
cd .....
bms init
bms compile --tidy (go mod tidy)
```

require
https://github.com/ethereum/go-ethereum/tree/master/cmd/abigen
```
cd go/pkg/mod/github.com/ethereum/go-ethereum@v1.13.12/cmd/abigen
go build -o /usr/local/bin/abigen .
```
