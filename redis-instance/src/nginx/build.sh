#!/bin/bash

HERE=`pwd`

cd nginx-1.21.1

find . -type f -exec sed -i 's/\r//' {} \;
chmod +x ./configure

./configure --prefix=/usr/bin --with-poll_module --with-file-aio --with-stream --without-pcre \
    --without-http_charset_module               \
    --without-http_gzip_module                  \
    --without-http_ssi_module                   \
    --without-http_userid_module                \
    --without-http_access_module                \
    --without-http_auth_basic_module            \
    --without-http_mirror_module                \
    --without-http_autoindex_module             \
    --without-http_geo_module                   \
    --without-http_map_module                   \
    --without-http_split_clients_module         \
    --without-http_referer_module               \
    --without-http_rewrite_module               \
    --without-http_proxy_module                 \
    --without-http_fastcgi_module               \
    --without-http_uwsgi_module                 \
    --without-http_scgi_module                  \
    --without-http_grpc_module                  \
    --without-http_memcached_module             \
    --without-http_limit_conn_module            \
    --without-http_limit_req_module             \
    --without-http_empty_gif_module             \
    --without-http_browser_module               \
    --without-http_upstream_hash_module         \
    --without-http_upstream_ip_hash_module      \
    --without-http_upstream_least_conn_module   \
    --without-http_upstream_random_module       \
    --without-http_upstream_keepalive_module    \
    --without-http_upstream_zone_module         \
    --without-http                              \
    --without-http-cache    

make -j "$(nproc)"

cd $HERE

echo `date` cp nginx-1.21.1/objs/nginx /usr/bin/ -rvf
cp nginx-1.21.1/objs/nginx /usr/bin/ -rvf
