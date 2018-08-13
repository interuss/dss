#!/bin/sh
set -ex

ORIGINALDIR=`pwd`
ZOOBASEDIR=`pwd`/zoo
mkdir -p $ZOOBASEDIR
cd $ZOOBASEDIR

export ZOO_USER=zookeeper
export ZOO_CONF_DIR=$ZOOBASEDIR/zoo_conf
export ZOO_DATA_DIR=$ZOOBASEDIR/zoo_data
export ZOO_DATA_LOG_DIR=$ZOOBASEDIR/zoo_datalog
export ZOO_PORT=2181
export ZOO_TICK_TIME=2000
export ZOO_INIT_LIMIT=5
export ZOO_SYNC_LIMIT=2
export ZOO_MAX_CLIENT_CNXNS=60
export ZOO_SERVERS=0.0.0.0:2888:3888


mkdir -p "$ZOO_DATA_LOG_DIR" "$ZOO_DATA_DIR" "$ZOO_CONF_DIR";

ZOO_DISTRO_NAME=zookeeper-3.4.12
wget -q "https://www.apache.org/dist/zookeeper/$ZOO_DISTRO_NAME/$ZOO_DISTRO_NAME.tar.gz"; \
wget -q "https://www.apache.org/dist/zookeeper/$ZOO_DISTRO_NAME/$ZOO_DISTRO_NAME.tar.gz.asc"; \
tar -xzf "$ZOO_DISTRO_NAME.tar.gz"; \
mv "$ZOO_DISTRO_NAME/conf/"* "$ZOO_CONF_DIR"; \
rm -rf "$GNUPGHOME" "$ZOO_DISTRO_NAME.tar.gz" "$ZOO_DISTRO_NAME.tar.gz.asc"


ZOO_CONFIG="$ZOO_CONF_DIR/zoo.cfg"

echo "clientPort=$ZOO_PORT" >> "$ZOO_CONFIG"
echo "dataDir=$ZOO_DATA_DIR" >> "$ZOO_CONFIG"
echo "dataLogDir=$ZOO_DATA_LOG_DIR" >> "$ZOO_CONFIG"

echo "tickTime=$ZOO_TICK_TIME" >> "$ZOO_CONFIG"
echo "initLimit=$ZOO_INIT_LIMIT" >> "$ZOO_CONFIG"
echo "syncLimit=$ZOO_SYNC_LIMIT" >> "$ZOO_CONFIG"

echo "maxClientCnxns=$ZOO_MAX_CLIENT_CNXNS" >> "$ZOO_CONFIG"

for server in $ZOO_SERVERS; do
    echo "$server" >> "$ZOO_CONFIG"
done

echo "${ZOO_MY_ID:-1}" > "$ZOO_DATA_DIR/myid"


echo "export PATH=\$PATH:$ZOOBASEDIR/$ZOO_DISTRO_NAME/bin" >> ../setup.sh
echo "export ZOOCFGDIR=$ZOO_CONF_DIR" >> ../setup.sh
echo "export ZOO_LOG_DIR=$ZOO_LOG_DIR" >> ../setup.sh


cd $ORIGINALDIR
