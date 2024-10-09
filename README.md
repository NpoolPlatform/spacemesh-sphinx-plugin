[![Test](https://github.com/NpoolPlatform/sphinx-plugin-p2/actions/workflows/main.yml/badge.svg?branch=master)](https://github.com/NpoolPlatform/sphinx-plugin-p2/actions/workflows/main.yml)

支持币种：Spacemesh Chia

详细信息：[please see it](https://github.com/NpoolPlatform/sphinx-plugin/blob/master/README.md)

## 币种特性

### Chia

Chia在浏览器或节点上查询交易时，只能查到硬币的信息。

而过程中产生所谓的transaction_id、tx_id只是用来查看节点是否已经处理了这个交易，一旦交易被链处理，交易ID就没用了，区块链也不会记录。

如何确认Chia交易是否有效，第一记录的交易ID在节点的内存池（MemoryPool）已经无法查到，第二记录的CoinID被查询到已经被用掉了。

在数据库中的Payload字段里可以查询到tx_id和SpentCoinIDs，可以用SpentCoinIDs去浏览器查询硬币是否被花掉，以及是如何花掉的。
```Info
payload: {"tx_id":"5cfcd1a5a4c660ffdd5de4faa6fdc368a775d22f2a4aa3b551035d7ec45317db","SpentCoinIDs":["0xafdf0ab5dee1a5db9e3979a2fb6195e7b5b4dd5a7306f1ffcd56b205340b0a01"]}
```
