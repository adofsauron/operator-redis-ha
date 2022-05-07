#ifndef __ETCD_CLIENT_C_H__
#define __ETCD_CLIENT_C_H__

#ifdef  __cplusplus
#define ETCD_EXPORT_C extern "C"
#else
#define ETCD_EXPORT_C
#endif  //__cplusplus

ETCD_EXPORT_C void* etcd_Client_WithSSL(const char* etcd_url,
                                    const char* ca,
                                    const char* cert,
                                    const char* key);

ETCD_EXPORT_C void* etcd_Client_WithUrl(const char* etcd_url);

ETCD_EXPORT_C int etcd_add(void* etcd, const char* key, const char* value, int ttl);

ETCD_EXPORT_C int etcd_rmdir(void* etcd, const char* dir);

ETCD_EXPORT_C int etcd_set(void* etcd, const char* key, const char* value, int ttl);

#undef ETCD_EXPORT_C

#endif /* __ETCD_CLIENT_C_H__ */
