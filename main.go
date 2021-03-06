package main

import (
	"encoding/json"
	"fmt"
	"github.com/go-resty/resty/v2"
	"github.com/mdp/qrterminal/v3"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"
)

const ua = "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/100.0.4896.127 Safari/537.36\""

var (
	config                         = &Config{}
	client                         = resty.New()
	login                          = resty.New()
	cookies                        []*http.Cookie
	orderId                        string
	itemName                       string
	strStartTime                   string
	oauthKey                       string
	qrCodeUrl                      string
	loginStatus                    bool
	startTime, errorTime, waitTime int64
	bp, price                      float64
	rankInfo                       *Rank
)

type GetLoginUrl struct {
	Code   int  `json:"code"`
	Status bool `json:"status"`
	Ts     int  `json:"ts"`
	Data   struct {
		Url      string `json:"url"`
		OauthKey string `json:"oauthKey"`
	} `json:"data"`
}

type GetLoginInfo struct {
	Status bool `json:"status"`
	//Data    string `json:"data"`
	Message string `json:"message"`
}

type Config struct {
	BuyNum      string `json:"buy_num"`
	CouponToken string `json:"coupon_token"`
	Device      string `json:"device"`
	ItemId      string `json:"item_id"`
	TimeBefore  int    `json:"time_before"`
	Cookies     struct {
		SESSDATA        string `json:"SESSDATA"`
		BiliJct         string `json:"bili_jct"`
		DedeUserID      string `json:"DedeUserID"`
		DedeUserIDCkMd5 string `json:"DedeUserID__ckMd5"`
	} `json:"cookies"`
}

type Static struct {
	AppId    int    `json:"appId"`
	Platform int    `json:"platform"`
	Version  string `json:"version"`
	Abtest   string `json:"abtest"`
}

type Details struct {
	Data struct {
		Name       string `json:"name"`
		Properties struct {
			SaleTimeBegin    string `json:"sale_time_begin"`
			SaleBpForeverRaw string `json:"sale_bp_forever_raw"`
		}
		CurrentActivity struct {
			PriceBpForever float64 `json:"price_bp_forever"`
		} `json:"current_activity"`
	} `json:"data"`
}

type Now struct {
	Data struct {
		Now int64 `json:"now"`
	} `json:"data"`
}

type Navs struct {
	Code int `json:"code"`
	Data struct {
		Wallet struct {
			BcoinBalance float64 `json:"bcoin_balance"`
		} `json:"wallet"`
	} `json:"data"`
}

type Asset struct {
	Data struct {
		Id   int `json:"id"`
		Item struct {
			ItemId int `json:"item_id"`
		} `json:"item"`
	} `json:"data"`
}

type Rank struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Ttl     int    `json:"ttl"`
	Data    struct {
		Rank []struct {
			Mid      int    `json:"mid"`
			Nickname string `json:"nickname"`
			Avatar   string `json:"avatar"`
			Number   int    `json:"number"`
		} `json:"rank"`
	} `json:"data"`
}

type Create struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Ttl     int    `json:"ttl"`
	Data    struct {
		OrderId  string `json:"order_id"`
		State    string `json:"state"`
		BpEnough int    `json:"bp_enough"`
	} `json:"data"`
}

type Query struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Ttl     int    `json:"ttl"`
	Data    struct {
		OrderId  string `json:"order_id"`
		Mid      int    `json:"mid"`
		Platform string `json:"platform"`
		ItemId   int    `json:"item_id"`
		PayId    string `json:"pay_id"`
		State    string `json:"state"`
	} `json:"data"`
}

type Wallet struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Ttl     int    `json:"ttl"`
	Data    struct {
		BcoinBalance  float64 `json:"bcoin_balance"`
		CouponBalance int     `json:"coupon_balance"`
	} `json:"data"`
}

type SuitAsset struct {
	Data struct {
		Fan struct {
			IsFan      bool   `json:"is_fan"`
			Token      string `json:"token"`
			Number     int    `json:"number"`
			Color      string `json:"color"`
			Name       string `json:"name"`
			LuckItemId int    `json:"luck_item_id"`
			Date       string `json:"date"`
		} `json:"fan"`
	} `json:"data"`
}

// ????????????
func webLogin() {
	log.Println("??????????????? SESSDATA, ????????????????????????????????????????????????.")
	_, _ = fmt.Scanf("%s", "")
	getLoginUrl()
	genQrCode()
	getLoginInfo()
}

// ?????????????????????token
func getLoginUrl() {
	g := &GetLoginUrl{}
	_, err := login.R().
		SetResult(g).
		Get("/qrcode/getLoginUrl")

	checkErr(err)
	qrCodeUrl = g.Data.Url
	oauthKey = g.Data.OauthKey
	//fmt.Println(r)
	//fmt.Println(qrCodeUrl)
	//fmt.Println(oauthKey)
}

// ?????????????????????
func getLoginInfo() {
	for {
		task := time.NewTimer(3 * time.Second)
		data := map[string]string{
			"oauthKey": oauthKey,
		}

		g := &GetLoginInfo{}
		r, err := login.R().
			SetFormData(data).
			SetResult(g).
			Post("/qrcode/getLoginInfo")

		checkErr(err)
		loginStatus = g.Status
		//fmt.Println(r)
		//fmt.Println(loginStatus)

		if loginStatus == true {
			cookies = r.Cookies()
			config.Cookies.DedeUserID = cookies[0].Value
			config.Cookies.DedeUserIDCkMd5 = cookies[1].Value
			config.Cookies.SESSDATA = cookies[2].Value
			config.Cookies.BiliJct = cookies[3].Value

			result, err := json.MarshalIndent(config, "", " ")
			checkErr(err)
			
			err = ioutil.WriteFile("./config.json", result, 644)
			checkErr(err)

			break
		}
		<-task.C
	}

}

// ???????????????
func genQrCode() {
	qrterminal.Generate(qrCodeUrl, qrterminal.L, os.Stdout)
}

func nav() {
	params := map[string]string{
		"csrf": config.Cookies.BiliJct,
	}
	navs := &Navs{}
	_, err := client.R().
		SetResult(navs).
		SetQueryParams(params).
		Get("/web-interface/nav")
	checkErr(err)
	if navs.Code == -101 {
		log.Fatalln("???????????????????????????cookies.")
	}
	bp = navs.Data.Wallet.BcoinBalance
	log.Printf("????????????, ????????????B????????????: %v.", bp)
}

func popup() {
	params := map[string]string{
		"csrf": config.Cookies.BiliJct,
	}
	_, err := client.R().
		SetQueryParams(params).
		Get("/garb/popup")
	checkErr(err)
	//fmt.Println(r)
}

func detail() {
	params := map[string]string{
		"csrf":    config.Cookies.BiliJct,
		"item_id": config.ItemId,
		"part":    "suit",
	}
	details := &Details{}
	_, err := client.R().
		SetQueryParams(params).
		SetResult(details).
		Get("/garb/v2/mall/suit/detail")
	checkErr(err)
	itemName = details.Data.Name
	strStartTime = details.Data.Properties.SaleTimeBegin
	startTime, err = strconv.ParseInt(strStartTime, 10, 64)
	checkErr(err)
	if details.Data.CurrentActivity.PriceBpForever == 0 {
		p, _ := strconv.ParseFloat(details.Data.Properties.SaleBpForeverRaw, 64)
		price = p / 100
	} else {
		price = details.Data.CurrentActivity.PriceBpForever / 100
	}
	log.Printf("????????????: %v???????????????: %v.", details.Data.Name, startTime)
	if price > bp {
		log.Fatalf("???????????????????????????????????????????????? %.2f B???.", price)
	}
}

func asset() {
	params := map[string]string{
		"csrf":    config.Cookies.BiliJct,
		"item_id": config.ItemId,
		"part":    "suit",
	}
	_, err := client.R().
		SetQueryParams(params).
		Get("/garb/user/asset")
	checkErr(err)
	//fmt.Println(r)
}

func state() {
	params := map[string]string{
		"csrf":    config.Cookies.BiliJct,
		"item_id": config.ItemId,
		"part":    "suit",
	}
	_, err := client.R().
		SetQueryParams(params).
		Get("/garb/user/reserve/state")
	checkErr(err)
}

func rank() {
	params := map[string]string{
		"csrf":    config.Cookies.BiliJct,
		"item_id": config.ItemId,
	}
	ranks := &Rank{}
	_, err := client.R().
		SetQueryParams(params).
		SetResult(ranks).
		Get("/garb/rank/fan/recent")
	checkErr(err)
	rankInfo = ranks
	//fmt.Println(r)
}

func stat() {
	params := map[string]string{
		"csrf":    config.Cookies.BiliJct,
		"item_id": config.ItemId,
	}
	_, err := client.R().
		SetQueryParams(params).
		Get("/garb/order/user/stat")
	checkErr(err)
	//fmt.Println(r)
}

func coupon() {
	params := map[string]string{
		"csrf":    config.Cookies.BiliJct,
		"item_id": config.ItemId,
	}
	_, err := client.R().
		SetQueryParams(params).
		Get("/garb/coupon/usable")
	checkErr(err)
	//fmt.Println(r)
}

func create() {
	for {
		// 500ms ????????????
		task := time.NewTimer(1 * time.Second)
		params := map[string]string{
			"add_month":    "-1",
			"buy_num":      config.BuyNum,
			"coupon_token": "",
			"csrf":         config.Cookies.BiliJct,
			"currency":     "bp",
			"item_id":      config.ItemId,
			"platform":     config.Device,
		}
		creates := &Create{}
		r, err := client.R().
			SetQueryParams(params).
			SetResult(creates).
			Post("/garb/v2/trade/create")
		checkErr(err)
		switch creates.Code {
		case 0: // ??????????????????????????????????????????
			orderId = creates.Data.OrderId
			log.Println(r)
			task.Reset(0)
			break
		case -403:
			log.Fatalln("???????????????.")
		case 69949:
			errorTime += 1
			log.Println(r)
			log.Println("?????????64494.")
			go coupon()
			if errorTime >= 5 {
				log.Fatalln("??????????????????????????????????????????...")
			}
		default:
			errorTime += 1
			log.Println(r)
			go coupon()
			if errorTime >= 5 {
				log.Println(r)
				log.Fatalln("??????????????????????????????????????????...")
			}
		}
		<-task.C
	}
}

func tradeQuery() {
	for {
		task := time.NewTimer(500 * time.Millisecond)
		params := map[string]string{
			"csrf":     config.Cookies.BiliJct,
			"order_id": orderId,
		}
		query := &Query{}
		_, err := client.R().
			SetQueryParams(params).
			SetResult(query).
			Get("/garb/trade/query")
		checkErr(err)
		state := query.Data.State
		if state == "paid" {
			log.Println("???????????????.")
			task.Reset(0 * time.Millisecond)
			break
		} else {
			errorTime += 1
			//log.Println(r)
			if errorTime >= 5 {
				log.Fatalln("??????????????????????????????????????????...")
			}
		}
		<-task.C
	}
}

func wallet() {
	params := map[string]string{
		"platform": "android",
	}
	response := &Wallet{}
	_, err := client.R().
		SetQueryParams(params).
		SetResult(response).
		Get("/garb/user/wallet?platform")
	checkErr(err)
	log.Printf("?????????????????????: %v.", response.Data.BcoinBalance)
}

func suitAsset() {
	params := map[string]string{
		"item_id": config.ItemId,
		"part":    "suit",
		"trial":   "0",
	}
	response := &SuitAsset{}
	_, err := client.R().
		SetQueryParams(params).
		SetResult(response).
		Get("garb/user/suit/asset")
	checkErr(err)
	//fmt.Println(r)
	log.Printf("??????: %v ??????: %v.", itemName, response.Data.Fan.Number)
}

func now() {
	result := &Now{}
	clock := resty.New()
	for {
		r, err := clock.R().
			SetResult(result).
			EnableTrace().
			SetHeader("user-agent", ua).
			Get("http://api.bilibili.com/x/report/click/now")
		checkErr(err)
		if result.Data.Now >= startTime-28 {
			waitTime = r.Request.TraceInfo().TotalTime.Milliseconds()
			break
		}
	}
}

func clientInfo() {
	test := resty.New()
	resp, err := test.R().
		EnableTrace().
		SetHeader("user-agent", ua).
		Get("https://api.bilibili.com/client_info")
	fmt.Println("Response Info:")
	fmt.Println("  Error      :", err)
	fmt.Println("  Status Code:", resp.StatusCode())
	fmt.Println("  Status     :", resp.Status())
	fmt.Println("  Proto      :", resp.Proto())
	fmt.Println("  Time       :", resp.Time())
	fmt.Println("  Received At:", resp.ReceivedAt())
	fmt.Println("  Body       :\n", resp)
	fmt.Println("  Headers    :", resp.Header())
	fmt.Println()

	// Explore trace info
	fmt.Println("Request Trace Info:")
	ti := resp.Request.TraceInfo()
	fmt.Println("  DNSLookup     :", ti.DNSLookup)
	fmt.Println("  ConnTime      :", ti.ConnTime)
	fmt.Println("  TCPConnTime   :", ti.TCPConnTime)
	fmt.Println("  TLSHandshake  :", ti.TLSHandshake)
	fmt.Println("  ServerTime    :", ti.ServerTime)
	fmt.Println("  ResponseTime  :", ti.ResponseTime)
	fmt.Println("  TotalTime     :", ti.TotalTime)
	fmt.Println("  IsConnReused  :", ti.IsConnReused)
	fmt.Println("  IsConnWasIdle :", ti.IsConnWasIdle)
	fmt.Println("  ConnIdleTime  :", ti.ConnIdleTime)
	fmt.Println("  RequestAttempt:", ti.RequestAttempt)
	fmt.Println("  RemoteAddr    :", ti.RemoteAddr.String())
}

func outPutRank() {
	if rankInfo.Data.Rank == nil {
		log.Println("??????????????????????????????????????????")
		return
	}
	log.Println("??????????????????:")
	fmt.Println("")
	for _, x := range rankInfo.Data.Rank {
		fmt.Printf("??????: %v\t?????????: %v\n", x.Number, x.Nickname)
	}
	fmt.Println("")
}

func waitToStart() {
	log.Println("??????????????????...")
	for {
		task := time.NewTimer(1 * time.Millisecond)
		t := time.Now().Unix()
		fmt.Printf("%v\r", t)
		if t >= startTime-30 {
			log.Println("?????????????????????")
			task.Reset(0 * time.Millisecond)
			break
		}
		<-task.C
	}
}

func init() {
	// ?????????log
	log.SetFlags(log.Ldate | log.Lmicroseconds)
	jsonFile, err := ioutil.ReadFile("./config.json")
	checkErr(err)
	err = json.Unmarshal(jsonFile, config)
	checkErr(err)

	// ??????
	if config.Cookies.SESSDATA == "" {
		login.SetHeader("user-agent", ua)
		login.SetBaseURL("https://passport.bilibili.com")
		webLogin()
	}

	cookies = []*http.Cookie{
		{Name: "SESSDATA", Value: config.Cookies.SESSDATA},
		{Name: "bili_jct", Value: config.Cookies.BiliJct},
		{Name: "DedeUserID", Value: config.Cookies.DedeUserID},
		{Name: "DedeUserID__ckMd5", Value: config.Cookies.DedeUserIDCkMd5},
	}

	client.SetHeader("user-agent", ua)
	client.SetBaseURL("https://api.bilibili.com/x")
	client.SetCookies(cookies)
}

func main() {
	// ??????
	//clientInfo()
	//os.Exit(0)

	// ????????????
	nav()

	// ??????
	popup()

	// ??????????????????
	detail()

	// ???????????????!!
	asset()
	state()
	rank()
	stat()
	coupon()

	// ??????????????????
	outPutRank()

	// ????????????
	//startTime = time.Now().Unix() + 10

	// ??????????????????????????????????????????????????????????????????
	waitToStart()

	// ??????b????????????????????????????????????????????????
	now()

	// ?????????
	time.Sleep((27000 - time.Duration(waitTime) - time.Duration(config.TimeBefore)) * time.Millisecond)

	// ???????????????
	start := time.NewTimer(1000 * time.Millisecond)
	detail()
	go asset()
	go state()
	go rank()
	go stat()
	go coupon()
	<-start.C

	// ????????????
	create()

	// ????????????
	tradeQuery()

	// ????????????
	nav()
	wallet()

	// ????????????
	suitAsset()
}

func checkErr(err error) {
	if err != err {
		log.Fatalln(err)
	}
}
