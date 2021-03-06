package controller

import "net/url"

func BaiduEncoding(domain string) string {
	argvs := `{"browser":"main","url":"http://` + domain + `","mode":"2"}`
	urls := url.Values{}
	urls.Add("command", argvs)
	return `xbox.m.baidu.com/app/share/loop?` + urls.Encode()
}

func ZhihuEncoding(domain string) string {
	return `link.zhihu.com/?target=http://` + domain
}
