package ctlconfig

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strconv"

)

type ConfigParam struct {
	ETCD_PATH_CERT   string
	ETCD_PATH_KEY    string
	ETCD_PATH_CACERT string

	ETCD_VALUE_CERT   string
	ETCD_VALUE_KEY    string
	ETCD_VALUE_CACERT string
}

var g_configParam ConfigParam

func GetconfigParam() *ConfigParam {
	return &g_configParam
}

func getValueFromEnvString(canNil bool, envName string, value *string) {
	*value = os.Getenv(envName)
	fmt.Println(envName, " = [", *value, "]")

	if canNil == false {
		if "" == *value {
			fmt.Println(envName, " fail, cant't be nil")
			os.Exit(1)
		}
	}
}

func getValueFromEnvInt(envName string, value *int) bool {
	var err error
	*value, err = strconv.Atoi(os.Getenv(envName))
	if err != nil {
		fmt.Println(envName, " covert to int fail, err = [", err, "]")
		return false
	}

	fmt.Println(envName, " = [", *value, "]")
	return true
}

func readFile(fileName string) (string, error) {
	f, err := ioutil.ReadFile(fileName)
	if err != nil {
		log.Printf("读取文件失败:%#v", err)
		return "", err
	}
	return string(f), nil
}

func getFileValue(fileName string, value *string) {
	var err error
	*value, err = readFile(fileName)
	if nil != err {
		fmt.Println(fmt.Sprintf("readFile fail, %s, %v", fileName, err))
		os.Exit(1)
	}

	// fmt.Println(fmt.Sprintf("read %s value: %s", fileName, *value))
}

func LoadconfigParam() {

	fmt.Println("=======================================")

	getValueFromEnvString(false, "ETCD_PATH_CERT", &g_configParam.ETCD_PATH_CERT)
	getValueFromEnvString(false, "ETCD_PATH_KEY", &g_configParam.ETCD_PATH_KEY)
	getValueFromEnvString(false, "ETCD_PATH_CACERT", &g_configParam.ETCD_PATH_CACERT)

	getFileValue(g_configParam.ETCD_PATH_CERT, &g_configParam.ETCD_VALUE_CERT)
	getFileValue(g_configParam.ETCD_PATH_KEY, &g_configParam.ETCD_VALUE_KEY)
	getFileValue(g_configParam.ETCD_PATH_CACERT, &g_configParam.ETCD_VALUE_CACERT)

	fmt.Println("=======================================")
}
