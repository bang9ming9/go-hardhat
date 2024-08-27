# go-hardhat
golang 으로 개발된 컨트랙트 개발 도구(tool) 입니다.<br>
컨트랙트 개발을 위한 프로젝트 생성을 도와주며, 테스트를 위한 시뮬레이션 과 유틸 함수를 제공합니다.

# Build
```bash
# go 1.22
git clone https://github.com/bang9ming9/go-hardhat && cd go-hardhat

### bms: bang9ming9 solidity ###

# case1: /usr/local/bin 에 빌드하기
sudo go build -o /usr/local/bin/bms .

# case2: 원하는 경로 빌드 후 PATH 등록하기
go build -o <YOUEPATH>/bms .
echo 'export PATH=$PATH:<YOUEPATH>/bms' >> ~/.zshrc # or ~/.bashrc
```

# Basic Useage
## 프로젝트 생성
```bash
mkdir my-contract && cd my-contract
bms init my-contract
```
*go.mod* 파일 생성과 동시에 `contracts`, `test` 디렉토리를 생성합니다.<br>

`contracts` 디렉토리에는 *remapping.txt*, *.prettierrc* 파일을 생성합니다.<br>
해당 파일은 VSCode 의 [Solidity Extention](https://marketplace.visualstudio.com/items?itemName=JuanBlanco.solidity) 에서 유용하게 동작합니다.<br>
> <br>
> VSCode에서 editor.formatOnSave를 true로 설정하면 파일을 저장할 때 자동으로 코드가 포맷팅됩니다.<br>
> <br>
***setting.json***
```
{
    ...
    "editor.formatOnSave": true,
    ...
}
```

`test` 디렉토리에는 **bms** 명령어와 **go-hardhat** 코드 버전을 일치시키기 위해 *base.go* 파일이 생성됩니다.
<br>

## 컨트랙트 컴파일
프로젝트의 `contracts` 디렉토리에 있는 모든 *.sol* 파일을 찾아 컴파일한 후, **golang** 으로 바인딩 합니다.
```bash
bms compile <solc-version>
```
>
> `compile` 명령어에는 여러 가지 옵션이 있으며, `bms compile -h`를 통해 확인할 수 있습니다.
>
> 예를 들어, [OpenZeppelin](https://github.com/OpenZeppelin/openzeppelin-contracts) 코드를 사용하고 있고 해당 디렉토리가 `contracts` 폴더에 포함되어 있다면, <br>
> `--exclude ./contracts/openzeppelin-contracts` 옵션을 사용하여 컴파일 대상에서 제외할 수 있습니다.
>


## 테스트 코드
```go
import (
    ...
	"testing"
	"github.com/bang9ming9/go-hardhat/bms"
	utils "github.com/bang9ming9/go-hardhat/bms/utils"
    ...
)

// 시뮬레이션 백앤드 생성
func ... (t *testing.T){
    ...
    backend := bms.NewBacked(t)
    ...
}

// 테스트 지갑 가져오기
func ... (t *testing.T){
    ...
    eoas := bms.GetEOAs(t, 10)
    ...
}

// 트랜잭션 실행하기
func ...(t *testing.T) {
    ...
    backend := bms.NewBacked(t)
    eoas := bms.GetEOAs(t, 10)
    txs := utils.NewTxPool(backend)

    // case: 단일 트랜잭션 성공
    require.NoError(t, txs.Exec(utils.SendDynamicTx(backend, owner, &eoa.From, []byte{})))

    // case: 트랜잭션 성공 
    for _, eoa := range eoas {
        // TxPool.Exec 을 통해 트랜잭션 목록 캐싱
        require.NoError(t, txs.Exec(utils.SendDynamicTx(backend, owner, &eoa.From, []byte{})))
    }
    // TxPool.AllReceiptStatusSuccessful 에서 에러가 리턴되지 않았다면 모두 성공한 트랜잭션
    require.NoError(t, txs.AllReceiptStatusSuccessful(ctx))

    // case: 예상되는 트랜잭션 실패
    require.Error(t, txs.Exec(utils.SendDynamicTx(backend, owner, &eoa.From, []byte{})))
    _, err := txs.Exec(utils.SendDynamicTx(backend, owner, &eoa.From, []byte{}))
    t.Log(err) // err 을 통해서 revert message 확인 가능
}
```

### 다양한 기능은 [bm-governance/test/b9m9_test.go](https://github.com/bang9ming9/bm-governance/blob/main/test/b9m9_test.go) 을 참고해 주세요.