#ifndef __ETCD_LINK_H__
#define __ETCD_LINK_H__

struct redisServer;


enum etcdLinkState {
    ETCD_UNINIT          = 0,
    ETCD_WAIT_CONNECT    = 1,
    ETCD_CONNECTED       = 2,
    ETCD_DIS_CONNECT     = 3,
};

typedef struct etcdLinkClient {
    enum etcdLinkState etcdstate;
    struct redisServer* redis_server;
    void* etcd;
    int etcd_create_path;
    int etcd_on_wath;
    int etcd_in_reconnect;
    char* replace_path;
    long long replace_path_del_tm;
} etcdLinkClient;

void etcdLinkCron(struct redisServer* server);

// 1: true, 0:fail
int etcdLinkTryCreateOldMasterPath(struct redisServer* server);

#endif /* __ETCD_LINK_H__ */
