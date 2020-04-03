#!/bin/bash
RESTART=$1

ADDR0="0xb79749F25Ef64F9AC277A4705887101D3311A0F4"
ADDR1="0x5E3230019fEd7aB462e3AC277E7709B9b2716b4F"
ADDR2="0x515B385bDc89bCc29077f2B00a88622883bfb498"
ADDR3="0xC927A0CF2d4a1B59775B5D0A35ec76d099e1FaD4"
ADDR4="0x48aFf0622a866d77651eAaA462Ea77b5F39D0ae1"
ADDR5="0xb715125A08140AEA83588a4b569599cde4a0a336"
ADDR6="0x499De281cd965781F1422b7cB73367C15DC416D2"
ADDR7="0xaA60af9BD19dc7438fd19457955C52982D070D27"
ADDR8="0x37da08b6Cd15c3aE905A25Df57B6841A5D80aC93"
ADDR9="0xec4A610a07e81264e8f7F1CAeAe522fEdD7e59c1"

KEY0="2628ca66087c6bc7f9eff7d70db7413d435e170040e8342e67b3db4e55ce752f"
KEY1="86e60281da515184c825c3f46c7ec490b075af1e74607e2e9a66e3df0fa22122"
KEY2="b77b291fab2b0a9e03b5ee0fb0f1140ff41780e93a39e534d54a05ccfad3eead"
KEY3="54a93b74538a7ab51062c7314ea9838519acae6b4ea3d47a7f367e866010364d"
KEY4="434e494f59f6228481256c0c88a375eef2c57be70e612576f302337f48a4634b"
KEY5="c85ab6a568ce788082664c0c17f86e332793895750455090f30f4578e4d20f9a"
KEY6="83d58f7a18e85b728bf5b00ce92d0d8491ae51a962331c8626e51ac32ba8b5f7"
KEY7="85a7751420007fba52e23eca493ac40c770b63c7a16f27ffec39fa01061bc435"
KEY8="5c148c5ba69b7b5c4e53d222e74e6edbbea72f3744fe2ab770320ae70b8d42c0"
KEY9="65d2ecce5d466cb3f9e0ca9acdf53575047ca71527f7c2ed2ab0de620918b2e7"

DATADIR=~/.root-chain-node-1

HTTP_PORT=8545
WS_PORT=8546

PERIOD=1

rootchain_running() {
  nc -z localhost "$HTTP_PORT"
}

start_rootchain() {
  echo "start root chain node"

  rm -rf $DATADIR

  make geth && build/bin/geth \
    --datadir $DATADIR \
    --dev \
    --dev.period $PERIOD \
    --dev.faucetkey $KEY0,$KEY1,$KEY2,$KEY3,$KEY4,$KEY5,$KEY6,$KEY7,$KEY8,$KEY9 \
    --miner.gastarget 100000000 \
    --miner.gasprice "10" \
    --rpc \
    --rpcport $HTTP_PORT  \
    --rpcapi eth,debug,net \
    --ws \
    --wsport $WS_PORT
}

restart_rootchain() {
  echo "restart root chain node"
  rm -rf $DATADIR/keystore
  make geth && build/bin/geth \
    --datadir $DATADIR \
    --dev \
    --dev.period $PERIOD \
    --dev.faucetkey $KEY0,$KEY1,$KEY2,$KEY3,$KEY4,$KEY5,$KEY6,$KEY7,$KEY8,$KEY9 \
    --miner.gastarget 100000000 \
    --miner.gasprice "10" \
    --rpc \
    --rpcport $HTTP_PORT \
    --rpcapi eth,debug,net \
    --ws \
    --wsport $WS_PORT
}

if [[ -z $RESTART ]]; then
  start_rootchain
else
  restart_rootchain
fi
