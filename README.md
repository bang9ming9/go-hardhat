go build -o ~/usr/local/bin/bms . (bang9ming9solidity)

example
```
cd .....
bsm init
bsm compile --tidy (go mod tidy)
bsm setsolc --version=0.8.25 --remappings=@openzeppelin/=...../node_modules/@openzeppelin,@other=...../
```

require
https://github.com/ethereum/go-ethereum/tree/master/cmd/abigen
```
cd go/pkg/mod/github.com/ethereum/go-ethereum@v1.13.12/cmd/abigen
go build -o /usr/local/bin/abigen .
```