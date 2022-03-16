package main

import (
	"flag"
	"fmt"
	"github.com/fatih/color"
	jsoniter "github.com/json-iterator/go"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

var Client http.Client
// 菜单参数
var Target string
var DorkFile string
var Keyword string
var Token string
var TokenFile string
var NeedWait bool
var NeedWaitSecond int64
var EachWait int64
// 所有token和dork
var Tokennum = 0
var Tokens []string
var Dorks []string

func query(dork string, token string) {
	// 构造请求
	guri := "https://api.github.com/search/code"
	uri, _ := url.Parse(guri)

	param := url.Values{}
	param.Set("q", dork)
	uri.RawQuery = param.Encode()

	req, _ := http.NewRequest("GET", uri.String(), nil)
	req.Header.Set("accept", "application/vnd.github.v3+json")
	req.Header.Set("Authorization", fmt.Sprintf("token %s", token))
	req.Header.Set("User-Agent", "HelloGitHub")

	resp, err := Client.Do(req)

	// 结果判断
	if err != nil {
		color.Red("error: %v", err)
	} else {
		source, _ := ioutil.ReadAll(resp.Body)
		var tmpSource map[string]jsoniter.Any
		_ = jsoniter.Unmarshal(source, &tmpSource)

		if tmpSource["documentation_url"] != nil { // 错误拦截
			if NeedWait {
				color.Blue("error: %s ; and we need wait %ds", jsoniter.Get(source, "documentation_url").ToString(), NeedWaitSecond)
				time.Sleep(time.Second * time.Duration(NeedWaitSecond))
				token = getToken()
				query(dork, token)
			} else {
				color.Red("error: %s", jsoniter.Get(source, "documentation_url").ToString())
			}
		} else if tmpSource["total_count"] != nil { // 总数
			totalCount := jsoniter.Get(source, "total_count").ToInt()
			totalCountString := color.YellowString(fmt.Sprintf("(%s)", strconv.Itoa(totalCount)))
			uriString := color.GreenString(strings.Replace(uri.String(), "https://api.github.com/search/code", "https://github.com/search", -1))
			fmt.Println(dork, " | ", totalCountString, " | ", uriString)
		} else { // 其他未知错误
			color.Blue("unknown error happened: %s", string(source))
		}
	}

}

func menu(){
	flag.StringVar(&DorkFile, "gd", "", "github dorks file path")
	flag.StringVar(&Keyword, "gk", "", "github search keyword")
	flag.StringVar(&Token, "token", "", "github personal access token")
	flag.StringVar(&TokenFile, "tf", "", "github personal access token file")
	flag.StringVar(&Target, "target", "", "target which search in github")
	flag.BoolVar(&NeedWait, "nw", true, "if get github api rate limited, need wait ?")
	flag.Int64Var(&NeedWaitSecond, "nws", 10, "how many seconds does it wait each time")
	flag.Int64Var(&EachWait, "ew", 0, "how many seconds does each request should wait ?")

	flag.Usage = func() {
		color.Green(`
 ____    __  __  __  __  ____  ___ 
(  _ \  /. |(  \/  )/  )(_  _)/ __)
 )(_) )(_  _))    (  )(   )(  \__ \
(____/   (_)(_/\/\_)(__) (__) (___/

                       v 0.1
`)
		fmt.Fprintf(flag.CommandLine.Output(), "Usage of %s:\n", os.Args[0])
		flag.PrintDefaults()
	}
	flag.Parse()

	// 判断是否输入参数
	if flag.NFlag() == 0 { // 使用的命令行参数个数，这个地方可以用来判断用户是否输入参数（程序正常情况下会使用默认参数）
		flag.Usage()
		os.Exit(0)
	}
	// 判断是否有目标
	if Target == "" {
		color.Red("require target")
		os.Exit(0)
	}
	// 判断是否有关键词
	if DorkFile == "" && Keyword == "" {
		color.Red("require keyword or dorkfile")
		os.Exit(0)
	}
	// 判断是否有token
	if Token == "" && TokenFile == "" {
		color.Red("require token or tokenfile")
		os.Exit(0)
	}
}

/*
解析token和dork
 */
func parseparam(){
	// 解析token
	if Token != "" {
		Tokens = []string {Token}
	} else if TokenFile != "" {
		tfres,err := ioutil.ReadFile(TokenFile)
		if err != nil {
			color.Red("file error: %v", err)
			os.Exit(0)
		} else {
			tfresLine := strings.Split(string(tfres), "\n")
			for {
				if tfresLine[len(tfresLine)-1] == "" {
					tfresLine = tfresLine[:len(tfresLine)-1]
				} else {
					break
				}
			}
			Tokens = tfresLine
		}
	}
	// 解析dork
	if Keyword != "" {
		Dorks = []string {Keyword}
	} else if DorkFile != "" {
		dkres,err := ioutil.ReadFile(DorkFile)
		if err != nil {
			color.Red("file error: %v", err)
			os.Exit(0)
		} else {
			dkresLine := strings.Split(string(dkres), "\n")
			for {
				if dkresLine[len(dkresLine)-1] == "" {
					dkresLine = dkresLine[:len(dkresLine)-1]
				} else {
					break
				}
			}
			Dorks = dkresLine
		}
	}
	color.Blue("[+] got %d tokens and %d dorks\n\n", len(Tokens), len(Dorks))
}

/*
多个token轮询
 */
func getToken() string {
	token := Tokens[Tokennum]
	Tokennum += 1
	if len(Tokens) == Tokennum {
		Tokennum = 0
	}
	return token
}

func main() {
	menu()
	parseparam()
	Client = http.Client{}

	for _,dork := range Dorks {
		token := getToken()
		query(fmt.Sprintf("%s %s", Target, dork), token)
		time.Sleep(time.Second * time.Duration(EachWait))
	}
}
