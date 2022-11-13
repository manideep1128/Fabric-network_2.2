#!/bin/bash

export PATH=${PWD}/../bin:${PWD}:$PATH
export FABRIC_CFG_PATH=${PWD}
export VERBOSE=false

# Generates Org certs using cryptogen tool
function generateCerts (){
  which cryptogen
  if [ "$?" -ne 0 ]; then
    echo "cryptogen tool not found. exiting"
    exit 1
  fi
  echo
  echo "##########################################################"
  echo "##### Generate certificates using cryptogen tool #########"
  echo "##########################################################"

  cryptogen generate --config=./crypto-config.yaml
  if [ "$?" -ne 0 ]; then
    echo "Failed to generate certificates..."
    exit 1
  fi
  echo
}

# Generate orderer genesis block, channel configuration transaction and
# anchor peer update transactions
function generateChannelArtifacts() {
  which configtxgen
  if [ "$?" -ne 0 ]; then
    echo "configtxgen tool not found. exiting"
    exit 1
  fi

  echo "##########################################################"
  echo "#########  Generating Orderer Genesis block ##############"
  echo "##########################################################"
  # Note: For some unknown reason (at least for now) the block file can't be
  # named orderer.genesis.block or the orderer will fail to launch!
    configtxgen -profile SampleMultiNodeEtcdRaft -channelID $SYS_CHANNEL -outputBlock ./genesis.block
  if [ "$?" -ne 0 ]; then
    echo "Failed to generate orderer genesis block..."
    exit 1
  fi
  echo
  echo "#################################################################"
  echo "### Generating channel configuration transaction 'channel.tx' ###"
  echo "#################################################################"
  configtxgen -profile FourSitesChannel -outputCreateChannelTx ./bloqchannel.tx -channelID $CHANNEL_NAME

  if [ "$?" -ne 0 ]; then
    echo "Failed to generate channel configuration transaction..."
    exit 1
  fi

  echo
  echo "#################################################################"
  echo "#######    Generating anchor peer update for Site1MSP   ##########"
  echo "#################################################################"
  configtxgen -profile FourSitesChannel -outputAnchorPeersUpdate \
  ./Site1MSPanchors.tx -channelID $CHANNEL_NAME -asOrg Site1MSP
  if [ "$?" -ne 0 ]; then
    echo "Failed to generate anchor peer update for Org1MSP..."
    exit 1
  fi

  echo
  echo "#################################################################"
  echo "#######    Generating anchor peer update for Site2MSP   ##########"
  echo "#################################################################"
  configtxgen -profile FourSitesChannel -outputAnchorPeersUpdate \
  ./Site2MSPanchors.tx -channelID $CHANNEL_NAME -asOrg Site2MSP
  if [ "$?" -ne 0 ]; then
    echo "Failed to generate anchor peer update for Org2MSP..."
    exit 1
  fi
  echo

  echo
  echo "#################################################################"
  echo "#######    Generating anchor peer update for Site3MSP   ##########"
  echo "#################################################################"
  configtxgen -profile FourSitesChannel -outputAnchorPeersUpdate \
  ./Site3MSPanchors.tx -channelID $CHANNEL_NAME -asOrg Site3MSP
  if [ "$?" -ne 0 ]; then
    echo "Failed to generate anchor peer update for Org3MSP..."
    exit 1
  fi
  echo

  echo
  echo "#################################################################"
  echo "#######    Generating anchor peer update for Site4MSP   ##########"
  echo "#################################################################"
  configtxgen -profile FourSitesChannel -outputAnchorPeersUpdate \
  ./Site4MSPanchors.tx -channelID $CHANNEL_NAME -asOrg Site4MSP
  if [ "$?" -ne 0 ]; then
    echo "Failed to generate anchor peer update for Org3MSP..."
    exit 1
  fi
  echo
}

# channel name defaults to "mychannel"
CHANNEL_NAME="bloqchannel"
SYS_CHANNEL="byfn-sys-channel"

generateCerts
generateChannelArtifacts