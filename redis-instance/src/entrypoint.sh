#!/bin/bash

mkdir -p /data
cd /data

touch /data/predixy.log
touch /data/redis.log

mkdir -p ./conf-redis
mkdir -p ./conf-predixy

bash /usr/bin/nginx-start.sh

if [ ! -f "./conf-redis/redis.conf" ]; then
    cp /usr/bin/conf-redis/* ./conf-redis/
fi

if [ ! -f "./conf-predixy/predixy.conf" ]; then
    cp /usr/bin/conf-predixy/* ./conf-predixy/

    LOCALHOST=`cat /etc/hosts | grep svc.cluster.local | awk -F ' ' '{print $2}'`

    echo `date` LOCALHOST=$LOCALHOST

    if [ "" == "LOCALHOST" ]; then
        exit 1
    fi

    MYSELF=`echo $LOCALHOST | awk -F '.' '{print $1}'`
    SERVICE=`echo $LOCALHOST | awk -F '.' '{print $2}'`
    NS=`echo $LOCALHOST | awk -F '.' '{print $3}'`

    echo `date` MYSELF=$MYSELF
    echo `date` SERVICE=$SERVICE
    echo `date` NS=$NS


    MYSELF_ADDR=(${MYSELF//-/ })
    MYSELF_ADDR_LAST=${#MYSELF_ADDR[*]}
    let MYSELF_ADDR_LAST--

    MASTER_ADDR=""
    SLAVE_ADDR=""
    BASE_ADDR=""

    index=0
    for VALUE in ${MYSELF_ADDR[*]}
    do
        if [ "$index" == "$MYSELF_ADDR_LAST" ]; then
            MASTER_ADDR="$BASE_ADDR-0"
            SLAVE_ADDR="$BASE_ADDR-1"
            break
        fi

        if [ "" == "$BASE_ADDR" ]; then
            BASE_ADDR=$VALUE
        else
            BASE_ADDR="$BASE_ADDR-$VALUE"
        fi

        let index++
    done


    MASTER_ADDR="$MASTER_ADDR.$SERVICE.$NS.svc.cluster.local"
    SLAVE_ADDR="$SLAVE_ADDR.$SERVICE.$NS.svc.cluster.local"
    
    echo `date` MASTER_ADDR=$MASTER_ADDR
    echo `date` SLAVE_ADDR=$SLAVE_ADDR

    echo `date` sed -i "s/__REDIS_ADDR_MASTER/$MASTER_ADDR/g" ./conf-predixy/standalone.conf
    sed -i "s/__REDIS_ADDR_MASTER/$MASTER_ADDR/g" ./conf-predixy/standalone.conf
    
    # echo `date` sed -i "s/__REDIS_ADDR_SLAVE/$SLAVE_ADDR/g" ./conf-predixy/standalone.conf
    # sed -i "s/__REDIS_ADDR_SLAVE/$SLAVE_ADDR/g" ./conf-predixy/standalone.conf

    echo `date` cat ./conf-predixy/standalone.conf
    cat ./conf-predixy/standalone.conf
fi

redis-server ./conf-redis/redis.conf

predixy  ./conf-predixy/predixy.conf &

tini tail -- -f /data/redis.log
