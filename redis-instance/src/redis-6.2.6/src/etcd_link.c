#include "etcd_link.h"
#include "server.h"
#include "cluster.h"
#include <etcd/client_c.h>

static etcdLinkClient* etcdLinkCreateClient(struct redisServer* server);
static void etcdLinkInitServer(etcdLinkClient* etcdc);
static void etcdLinkUpdate(etcdLinkClient* etcdc);
static void etcdLinkUpdateMaster(etcdLinkClient* etcdc);
static void etcdLinkUpdateSlave(etcdLinkClient* etcdc);

etcdLinkClient* etcdLinkCreateClient(struct redisServer* server)
{
     etcdLinkClient* etcdc = zmalloc(sizeof(*etcdc));

    etcdc->etcd = NULL;
    etcdc->etcdstate = ETCD_UNINIT;
    etcdc->redis_server = server;
    etcdc->etcd_create_path = 0;
    etcdc->etcd_on_wath = 0;
    etcdc->etcd_in_reconnect = 0;
    etcdc->replace_path = NULL;
    etcdc->replace_path_del_tm = 0;

    return etcdc;
}

void etcdLinkInitServer(etcdLinkClient* etcdc)
{
    if (etcdc->etcd) {
        etcdc->etcdstate = ETCD_CONNECTED;
        return;
    }

    // TODO: 测试模式不用https
    etcdc->etcd = etcd_Client_WithUrl(etcdc->redis_server->etcd_addr);
    if (NULL == etcdc->etcd) {
        serverLog(LL_WARNING, "etcd, etcd_Client_WithUrl fail, addr = [%s]", etcdc->redis_server->etcd_addr);
        return;
    }

    etcdc->etcdstate = ETCD_CONNECTED;
    serverLog(LL_NOTICE, "etcd, etcd_Client_WithUrl connect over, addr = [%s]", etcdc->redis_server->etcd_addr);
}


int etcdLinkTryCreateOldMasterPath(struct redisServer* server)
{
    return 0;
}

void etcdLinkCron(struct redisServer* server)
{
    if (!server->etcd_addr) {
        return;
    }

    if (!server->etcdClient) {
        server->etcdClient = etcdLinkCreateClient(server);
        return;
    }

    if (!server->etcdClient) {
        return;
    }

     etcdLinkClient* etcdc = server->etcdClient;
    if (ETCD_WAIT_CONNECT == etcdc->etcdstate) {
        return;
    }

    if (ETCD_UNINIT == etcdc->etcdstate) {
        etcdLinkInitServer(etcdc);
        return;
    }

    if (ETCD_DIS_CONNECT == etcdc->etcdstate) {
        etcdLinkInitServer(etcdc);
        return;
    }

    if (ETCD_CONNECTED != etcdc->etcdstate) {
        return;
    }

    etcdLinkUpdate(etcdc);
}

void etcdLinkUpdate(etcdLinkClient* etcdc)
{
    struct redisServer* server = etcdc->redis_server;
    int was_master = server->masterhost == NULL;

    if (was_master) {
        return etcdLinkUpdateMaster(etcdc);
    } else {
        return etcdLinkUpdateSlave(etcdc);
    }
}

void etcdLinkUpdateMaster(etcdLinkClient* etcdc)
{
    // struct redisServer* server = etcdc->redis_server;

    char name_path[CLUSTER_NAMELEN + 1] = {0};
    snprintf(name_path, CLUSTER_NAMELEN + 1, "%s", "/sys-redis-ha-default-redis-test");

    void* etcd = etcdc->etcd;
    int ret = etcd_add(etcd, name_path, "value", 1);
    if (0 != ret) {
        serverLog(LL_WARNING, "etcd, etcd_add fail, name_path = [%s]", name_path);
        return;
    }
}

void etcdLinkUpdateSlave(etcdLinkClient* etcdc)
{

}
