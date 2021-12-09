package main

import (
	"bytes"
	"fmt"
	"github.com/thoas/go-funk"
	"io/ioutil"
	"path"
	"regexp"
	"strings"
)

const (
	pocDir = `C:\Users\test\Dev\xray\pocs`
)

/* 坑点

poc-yaml-discuz-wooyun-2010-080723 cookie 会丢掉末尾的 ;
poc-yaml-confluence-cve-2019-3396-lfi host 改不了
couchdb-cve-2017-12635.yml // content-length 不应该被设置
*/

/*
todo:
74cms-sqli.yml 包错误
dedecms-guestbook-sqli.yml // r2 packet
ecshop-collection-list-sqli.yml // poc error substr(1, 32)
joomla-cve-2018-7314-sql.yml // r1 packet
jupyter-notebook-unauthorized-access.yml //  !response.body.bcontains(b"Password:")
msvod-sqli.yml // r1 packet
phpcms-cve-2018-19127.yml // r1 packet
phpok-sqli.yml // r1 packet
sangfor-edr-arbitrary-admin-login.yml // xray 错误, run poc poc-yaml-sangfor-edr-arbitrary-admin-login err: execute r0() err: Get "http://127.0.0.1:7788/ui/login.php?user=admin": 302 response missing Location header
weblogic-cve-2020-14750.yml, xray 302 问题
solr-cve-2019-0193.yml 可能是 solr-velocity-template-rce.yml 变量冲突
weblogic-ssrf.yml // xray 发的 cookie key="value"，原本没有 “”
confluence-cve-2019-3396-lfi.yml // 不经过 burp 可以
glassfish-cve-2017-1000028-lfi.yml // invalid utf-8
*/

func main() {
	data, err := ioutil.ReadFile(`C:\Users\vbkoa\Downloads\xray\tt4.html`)
	if err != nil {
		panic(err)
	}
	scanned := regexp.MustCompile(`poc-yaml-.*?"`).FindAllString(string(data), -1)
	for i, name := range scanned {
		scanned[i] = name[9:len(name)-1] + ".yml"
	}

	infos, err := ioutil.ReadDir(pocDir)
	if err != nil {
		panic(err)
	}
	var expected []string
	for _, info := range infos {
		if !strings.HasSuffix(info.Name(), "yml") {
			continue
		}
		data, err := ioutil.ReadFile(path.Join(pocDir, info.Name()))
		if err != nil {
			panic(err)
		}
		if bytes.Contains(data, []byte("newReverse")) {
			continue
		}
		expected = append(expected, info.Name())
	}
	fmt.Println(scanned)
	fmt.Println(expected)
	black := strings.Split(blackList, "\n")
	for i, b := range black {
		black[i] = strings.TrimSpace(b)
	}
	for _, name := range expected {
		if !funk.ContainsString(scanned, name) && !funk.Contains(black, name) {
			fmt.Println(name)
		}
	}
}

var blackList = `
amtt-hiboss-server-ping-rce.yml
aspcms-backend-leak.yml
bash-cve-2014-6271.yml
drupal-cve-2019-6340.yml
elasticsearch-unauth.yml
gateone-cve-2020-35736.yml
gitlist-rce-cve-2018-1000533.yml
glassfish-cve-2017-1000028-lfi.yml
harbor-cve-2019-16097.yml
hikvision-info-leak.yml
hikvision-intercom-service-default-password.yml
iis-put-getshell.yml
joomla-cnvd-2019-34135-rce.yml
kafka-manager-unauth.yml
kong-cve-2020-11710-unauth.yml
laravel-debug-info-leak.yml
netentsec-ngfw-rce.yml
novnc-url-redirection-cve-2021-3654.yml
phpstudy-nginx-wrong-resolve.yml
ruijie-eweb-rce-cnvd-2021-09650.yml
spark-webui-unauth.yml
tensorboard-unauth.yml
terramaster-tos-rce-cve-2020-28188.yml
thinkcmf-write-shell.yml
thinkphp-v6-file-write.yml
tomcat-cve-2017-12615-rce.yml
webmin-cve-2019-15107-rce.yml
yonyou-nc-arbitrary-file-upload.yml
rabbitmq-default-password.yml
activemq-default-password.yml
74cms-sqli.yml
dedecms-guestbook-sqli.yml
ecshop-collection-list-sqli.yml
joomla-cve-2018-7314-sql.yml
jupyter-notebook-unauthorized-access.yml
msvod-sqli.yml
phpcms-cve-2018-19127.yml
phpok-sqli.yml
sangfor-edr-arbitrary-admin-login.yml
weblogic-cve-2020-14750.yml
solr-cve-2019-0193.yml 
weblogic-ssrf.yml
confluence-cve-2019-3396-lfi.yml
glassfish-cve-2017-1000028-lfi.yml 
`
