# 伤害事件

## 事件样例

```json
{
    "timestamp": 3195901,
    "type": "calculateddamage",
    "sourceID": 11,
    "targetID": 6,
    "ability": {
        "name": "攻击",
        "guid": 39147,
        "type": 1024,
        "abilityIcon": "000000-000101.png"
    },
    "fight": 7,
    "buffs": "1001191.1001362.1001457.1001918.1002680.1000297.",
    "hitType": 1,
    "amount": 0,
    "multiplier": 0.8,
    "packetID": 36630,
    "sourceResources": {
        "hitPoints": 114511339,
        "maxHitPoints": 114525943,
        "mp": 10000,
        "maxMP": 10000,
        "tp": 0,
        "maxTP": 0,
        "x": 10000,
        "y": 9000,
        "facing": -472
    },
    "targetResources": {
        "hitPoints": 246664,
        "maxHitPoints": 224240,
        "mp": 10000,
        "maxMP": 10000,
        "tp": 0,
        "maxTP": 0,
        "x": 10001,
        "y": 9540,
        "facing": -158,
        "absorb": 109
    }
},
{
    "timestamp": 3196701,
    "type": "damage",
    "sourceID": 11,
    "targetID": 6,
    "ability": {
        "name": "攻击",
        "guid": 39147,
        "type": 1024,
        "abilityIcon": "000000-000101.png"
    },
    "fight": 7,
    "hitType": 1,
    "amount": 0,
    "absorbed": 82120,
    "packetID": 36630,
    "multiplier": 0.8,
    "targetResources": {
        "hitPoints": 224240,
        "maxHitPoints": 224240,
        "mp": 10000,
        "maxMP": 10000,
        "tp": 0,
        "maxTP": 0,
        "x": 9999,
        "y": 9538,
        "facing": -158,
        "absorb": 109
    }
},
```

如上示例，伤害事件包含两个事件，一个是`calculateddamage`，一个是`damage`。`calculateddamage`事件是伤害计算事件，`damage`事件是伤害实际生效事件；这两个事件由 `PacketID` 关联。

> [!NOTE]
> 要注意的是一个 PacketID 下可能由多组对应的 `calculateddamage` 和 `damage` 事件（例如全屏大伤害），所以在解析时应找出 PacketID 下的所有事件后，根据 `sourceID` 和 `targetID` 这两者进行再次匹配。

在增伤/减伤分析功能中，既需要相关的 buffs，有需要具体的伤害数据，因此必须将这两个事件进行匹配合并；考虑到这两者的关联性，我们在日志预处理阶段就将这两个事件合并为一个事件（将 `calculateddamage` 事件中的 `buffs` 字段合并到 `damage` 事件中），便于后续的分析工作；后续在 ApplyLogs 阶段，我们会忽略 `calculateddamage` 事件，只处理 `damage` 事件。

要注意的是 FFLogs GraphQL 给出的数据中，`dataType=All` 的结果中，`calculateddamage` 事件中是不包含 `buffs` 字段的，所以在预处理阶段，我们除了拉取常规事件外，还需要额外查询一次 `dataType=DamageTaken` 获取 `calculateddamage` 事件，以便获取 `buffs` 字段。
