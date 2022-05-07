#!/bin/bash

ETCD_HOSTS=$ETCD_ADDR
NGINX_CONF=/usr/bin/nginx-etcd.conf

echo `date` ETCD_HOSTS=$ETCD_HOSTS
echo `date` NGINX_CONF=$NGINX_CONF

ETCD_HOST_ARR=(${ETCD_HOSTS//,/ })

# 先插入
for ETCD_HOST in ${ETCD_HOST_ARR[*]}
do
    echo `date` ETCD_HOST=$ETCD_HOST

    SERVER_HOST="server $ETCD_HOST weight=5;"

    echo `date` SERVER_HOST=$SERVER_HOST

    echo `date` sed -i "/__ETCD_SERVER_REPLACE/a\\\t$ETCD_HOST" $NGINX_CONF
    sed -i "/__ETCD_SERVER_REPLACE/a\\\t$SERVER_HOST" $NGINX_CONF
done

# 再删除
echo `date` sed -i "/__ETCD_SERVER_REPLACE/d"  $NGINX_CONF
sed -i "/__ETCD_SERVER_REPLACE/d"  $NGINX_CONF

echo `date` cat $NGINX_CONF
cat $NGINX_CONF

echo `date` rm -rvf /data/nginx-etcd.pid
rm -rvf /data/nginx-etcd.pid

echo `date` /usr/bin/nginx -c $NGINX_CONF -e /data/nginx-etcd.log
/usr/bin/nginx -c $NGINX_CONF -e /data/nginx-etcd.log

