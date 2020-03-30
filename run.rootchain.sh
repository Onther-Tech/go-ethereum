ADDR0="0x71562b71999873DB5b286dF957af199Ec94617F7";
ADDR1="0x3cd9f729c8d882b851f8c70fb36d22b391a288cd";
ADDR2="0x57ab89f4eabdffce316809d790d5c93a49908510";
ADDR3="0x6c278df36922fea54cf6f65f725267e271f60dd9";

KEY0="b71c71a67e1177ad4e901695e1b4b9ee17ae16c6668d313eac2f96dbcda3f291";
KEY1="bfaa65473b85b3c33b2f5ddb511f0f4ef8459213ada2920765aaac25b4fe38c5";
KEY2="067394195895a82e685b000e592f771f7899d77e87cc8c79110e53a2f0b0b8fc";
KEY3="ae03e057a5b117295db86079ba4c8505df6074cdc54eec62f2050e677e5d4e66";

make geth && build/bin/geth \
  --dev \
  --dev.period 1 \
  --dev.faucetkey "$KEY0,$KEY1,$KEY2,$KEY3" \
  --rpc \
  --rpcport 8545 \
  --rpcapi eth,debug,net \
  --rpcaddr 0.0.0.0 \
  --ws \
  --wsport 8546 \
  --wsaddr 0.0.0.0 \
  --wsapi eth,debug,net \
  --miner.gastarget 7500000 \
  --miner.gasprice "10"
