package kubeproxy

import (
	_const "Mini-K8s/cmd/const"
	"Mini-K8s/pkg/etcdstorage"
	"Mini-K8s/pkg/listwatcher"
	"Mini-K8s/pkg/object"
	"Mini-K8s/pkg/shell"
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"time"
)

type DNSUtil struct {
	ls          *listwatcher.ListWatcher
	stopChannel <-chan struct{}
}

func RunDNS(lsConfig *listwatcher.Config) *DNSUtil {
	dnsConfig := &DNSUtil{}
	ls, err := listwatcher.NewListWatcher(lsConfig)
	if err != nil {
		fmt.Println(err)
		return nil
	}
	dnsConfig.ls = ls

	watchFunc := func() {
		for {
			err := dnsConfig.ls.Watch(_const.DNS_CONFIG_PREFIX, dnsConfig.watchDNSChange, dnsConfig.stopChannel)
			if err != nil {
				fmt.Println("[DNS] watch error" + err.Error())
				time.Sleep(10 * time.Second)
			} else {
				return
			}
		}
	}
	go watchFunc()
	return dnsConfig
}

func (dnsUtil *DNSUtil) watchDNSChange(res etcdstorage.WatchRes) {
	if res.ResType == etcdstorage.DELETE {
		return
	} else {
		// assume only create
		dnsConfig := &object.DNSConfig{}
		err := json.Unmarshal(res.ValueBytes, dnsConfig)
		if err != nil {
			fmt.Println("[DNSWatch] Unmarshall fail" + err.Error())
			return
		}

		filename := "/etc/hosts"

		hostsFile, err := os.OpenFile(filename, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0644)
		defer hostsFile.Close()
		if err != nil {
			fmt.Println(err)
			return
		}
		w1 := bufio.NewWriter(hostsFile)
		fmt.Fprintln(w1, "192.168.1.4\t"+dnsConfig.Host)
		fmt.Fprintln(w1, "127.0.0.1\tlocalhost")

		err = w1.Flush()
		if err != nil {
			fmt.Println(err)
		}

		nginxFile, err := os.OpenFile("/etc/nginx/nginx.conf", os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0644)
		defer nginxFile.Close()
		if err != nil {
			fmt.Println(err)
			return
		}

		content := "user www-data;\nworker_processes auto;\n\nevents {\n\tworker_connections 768;\n}\n\n" +
			"http {\n\tsendfile on;\n\tkeepalive_timeout 65;\n\n\tinclude /etc/nginx/mime.types;\n\tdefault_type application/octet-stream;\n\t\n" +
			"\tserver {\n\t\tlisten 80;\n\t\tserver_name 192.168.1.4;\n\n"

		for _, dnsPath := range dnsConfig.DNSPaths {
			content = content + "\t\tlocation " +
				dnsPath.Path +
				" {\n\t\t\tproxy_pass  " +
				dnsPath.ServiceIp +
				";\n\t\t}\n\n"
		}

		content += "\t}\n}\n"

		w2 := bufio.NewWriter(nginxFile)
		fmt.Fprint(w2, content)

		err = w2.Flush()
		if err != nil {
			fmt.Println(err)
		}

		_, err = shell.ExecCmd("systemctl", "restart nginx")
		if err != nil {
			fmt.Println(err)
		}

	}
}
