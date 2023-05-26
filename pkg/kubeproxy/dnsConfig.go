package kubeproxy

import (
	"Mini-K8s/pkg/etcdstorage"
	"Mini-K8s/pkg/listwatcher"
	"bufio"
	"fmt"
	"os"
	"time"
)

type DNSConfig struct {
	ls          *listwatcher.ListWatcher
	stopChannel <-chan struct{}
}

func RunDNS(lsConfig *listwatcher.Config) *DNSConfig {
	dnsConfig := &DNSConfig{}
	ls, err := listwatcher.NewListWatcher(lsConfig)
	if err != nil {
		fmt.Println(err)
		return nil
	}
	dnsConfig.ls = ls
	watchFunc := func() {
		for {
			err := dnsConfig.ls.Watch("/dnsConfig", dnsConfig.watchDNSChange, dnsConfig.stopChannel)
			if err != nil {
				fmt.Println("[dnsConfig] watch error" + err.Error())
				time.Sleep(10 * time.Second)
			} else {
				return
			}
		}
	}
	go watchFunc()
	return dnsConfig
}

func (dnsConfig *DNSConfig) watchDNSChange(res etcdstorage.WatchRes) {
	if res.ResType == etcdstorage.DELETE {
		return
	} else {
		// assume only create
		f, err := os.OpenFile("/home/lcz/Core", os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0644)
		defer f.Close()
		if err != nil {
			fmt.Println(err)
			return
		}

		w := bufio.NewWriter(f)

		fmt.Fprint(w, ".:53 {\n\tbind 127.0.0.1\n\thosts {\n\t\t")
		fmt.Fprint(w, "127.0.0.1 example.lcz.com")
		fmt.Fprint(w, "\n\t\tfallthrough\n\t}\n\tforward . /etc/resolv.conf\n}")

		err = w.Flush()
		if err != nil {
			fmt.Println(err)
		}
		return
	}
}

func TestDns() {
	f, err := os.OpenFile("/home/lcz/Core", os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0644)
	defer f.Close()
	if err != nil {
		fmt.Println(err)
		return
	}

	w := bufio.NewWriter(f)

	fmt.Fprint(w, ".:53 {\n\tbind 127.0.0.1\n\thosts {\n")
	fmt.Fprintln(w, "\t\t127.0.0.1 example.lcz.com")
	fmt.Fprintln(w, "\t\t127.0.0.1 example.lcz")
	fmt.Fprint(w, "\t\tfallthrough\n\t}\n\tforward . /etc/resolv.conf\n}")

	err = w.Flush()
	if err != nil {
		fmt.Println(err)
	}
	return
}
