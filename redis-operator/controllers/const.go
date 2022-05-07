package controllers

const (
	STATUS_ON_START            = 0  // 启动初始状态
	STATUS_IN_CHECK_DELETE     = 1  // 检测是否要被删除
	STATUS_IN_PROCESS_DELETE   = 2  // 执行删除
	STATUS_IN_CHECK_STS        = 3  // 检测sts
	STATUS_IN_CREATE_STS       = 4  // 创建sts
	STATUS_IN_CHECK_PODS       = 5  // 检测pods
	STATUS_IN_CHECK_SERVICE    = 6  // 检测service
	STATUS_IN_CREATE_SERVICE   = 7  // 创建service
	STATUS_IN_CHECK_REDIS_HA   = 8  // 检测redis的ha
	STATUS_IN_CREATE_REDIS_HA  = 9  // 组件redis的主从关系
	STATUS_IN_FIX_REDIS_SERVER = 10 // 修复redis-server进程异常
	STATUS_IN_FORCE_REDO_POD   = 11 // pod使用了localpv出现异常无法自动调度, 强行调度pod
	STATUS_IN_CHECK_NORMAL     = 12 // 正常的心跳检测
	STATUS_IN_SET_ETCD_CRT     = 13 // 设置etcd的crt
)

const (
	TIME_INTERVAL_NORMAL = 10 // 正常间隔
	TIME_INTERVAL_EVENT  = 20 // 发生事件后的事件间隔
)
