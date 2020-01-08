package controllers

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"
	"qnsoft/web_api/models/shop"
	"qnsoft/web_api/utils/DateHelper"
	"qnsoft/web_api/utils/DbHelper"
	"qnsoft/web_api/utils/ErrorHelper"
	"qnsoft/web_api/utils/JobHelper"
	php2go "qnsoft/web_api/utils/Php2go"
	"qnsoft/web_api/utils/WebHelper"
)

func Taobao_joblist() {
	//开始采两小时前集所有订单
	_fasks := make([]*JobHelper.Task_model, 0)
	_model := JobHelper.Task_model{Id: 1001, Name: "定时采集淘宝渠道所有订单", Spec: "0 0/2 1-23 * * ?"} //每天早上1点到23点 每2分钟执行一次
	_fasks = append(_fasks, &_model)
	JobHelper.InitJobs(_fasks, job_qudao_all_Yesterday) //开始采集昨天全部订单
	//开始采集昨天结算订单
	_fasks_A := make([]*JobHelper.Task_model, 0)
	_model_A := JobHelper.Task_model{Id: 1002, Name: "定时采集淘宝渠道结算订单", Spec: "0 0/5 1-23 * * ?"} //每天早上1点到23点 每5分钟执行一次
	_fasks_A = append(_fasks_A, &_model_A)
	JobHelper.InitJobs(_fasks_A, job_qudao_ok_Yesterday) //采集昨天结算订单
	//实时采集最新订单
	_fasks_C := make([]*JobHelper.Task_model, 0)
	_model_C := JobHelper.Task_model{Id: 1003, Name: "实时采集淘宝渠道最新订单", Spec: "0 0/10 1-23 * * ?"} //每天早上1点到23点 每10分钟执行一次
	_fasks_C = append(_fasks_C, &_model_C)
	JobHelper.InitJobs(_fasks_C, job_qudao_now) //实时采集最新订单
}

/*
开始采集两小时前全部订单(每天全天执行，采集前两小时前的所有订单，顺便将漏采的补录到数据库)
*/
func job_qudao_all_Yesterday() {
	fmt.Println("开始采集所有渠道订单")
	_str_today_time := php2go.URLEncode(date.FormatDate(time.Now(), "yyyy-MM-dd HH:mm:ss"))
	var _start_time time.Time
	_model_new := &shop.TaobaokeJobtime{Id: 7}
	results, err := DbHelper.MySqlDb().Get(_model_new)
	ErrorHelper.CheckErr(err)
	if results {
		_start_time = _model_new.AllYesterdayStartTime
	}
	_str_start_time := php2go.URLEncode(date.FormatDate(_start_time, "yyyy-MM-dd HH:mm:ss"))
	_str_end_time := php2go.URLEncode(date.FormatDate(_start_time.Add(time.Minute*+20), "yyyy-MM-dd HH:mm:ss"))
	if _str_end_time < _str_today_time { //执行小于今天0点的订单
		//老版接口
		//_http_url := "http://api.vephp.com/order?vekey=V00002504Y26508322&span=1200&order_scene=2&tk_status=1&start_time=" + _str_start_time
		//新版接口
		_http_url := "http://api.vephp.com/orderdetails?vekey=V00002504Y26508322&start_time=" + _str_start_time + "&end_time=" + _str_end_time + "&order_scene=2&tk_status="
		//fmt.Println(_http_url)
		_req := Self_Get(_http_url, nil)
		//	fmt.Println(_req) Order_Model
		_order_model := Order_Model{}
		json.Unmarshal([]byte(_req), &_order_model)
		json_list := _order_model.Data.Results.PublisherOrderDto
		// json.Unmarshal(json_data4, &json_list)
		if len(json_list) > 0 {
			for _, v := range json_list {
				//fmt.Println(v)
				_CreateTime, _ := date.ParseAny(v.TkPaidTime)
				_ClickTime, _ := date.ParseAny(v.ClickTime)
				_RelationID := strconv.FormatInt(v.RelationID, 10)
				_Buyuser, _UserId := Get_userid_userpid(_RelationID) //根据RelationID取购买人user_id和父id
				is_ok, _ := Is_Order_Exist(v.TradeID)
				if is_ok { //修改订单状态
					ErrorHelper.LogInfo("获取的要分润的id及父id[?-?]", _Buyuser, _UserId)
					_model_update := &shop.TaobaokeRecord{UserId: _UserId, Buyuser: _Buyuser, Status: v.TkStatus, ClickTime: _ClickTime, OStatus: 0}
					_, err := DbHelper.MySqlDb().Where(" sno=? ", v.TradeID).Cols("user_id", "buyuser", "status", "click_time", "o_status").Update(_model_update)
					ErrorHelper.CheckErr(err)
					ErrorHelper.LogInfo("修改所有采集淘宝客新订单【" + v.TradeID + "】成功！")
				} else { //新增订单
					_model_insert := &shop.TaobaokeRecord{StoreId: 8, UserId: _UserId, Buyuser: _Buyuser, Sno: v.TradeID, GoodsTitle: v.ItemTitle, GoodsImg: "http:" + v.ItemImg, GoodsPrice: v.ItemPrice, Num: v.ItemNum, Amount: v.AlipayTotalPrice, CommRate: v.TotalCommissionRate, CommAmount: v.PubSharePreFee, Status: v.TkStatus, CreateTime: _CreateTime, ClickTime: _ClickTime, OStatus: 0}
					_, err := DbHelper.MySqlDb().Insert(_model_insert)
					ErrorHelper.CheckErr(err)
					ErrorHelper.LogInfo("新增所有采集淘宝客新订单【" + v.TradeID + "】成功！")
				}

			}
		}
		_model_update := &shop.TaobaokeJobtime{Id: 7, AllYesterdayStartTime: _start_time.Add(time.Minute * +15)}
		_, err1 := DbHelper.MySqlDb().Id(_model_update.Id).Cols("all_yesterday_start_time").Update(_model_update)
		ErrorHelper.CheckErr(err1)
	} else {
		ks_time := time.Date(time.Now().Year(), time.Now().Month(), time.Now().Day()-1, 0, 0, 0, 0, time.Now().Location())
		_model_update := &shop.TaobaokeJobtime{Id: 7, AllYesterdayStartTime: ks_time}
		_, err1 := DbHelper.MySqlDb().Id(_model_update.Id).Cols("all_yesterday_start_time").Update(_model_update)
		ErrorHelper.CheckErr(err1)
	}
}

/*
开始采集两小时前所有结算订单(每天全天执行，采集之两小时前的所有结算订单，顺便将漏采的补录到数据库)
*/
func job_qudao_ok_Yesterday() {
	fmt.Println("开始采集所有渠道结算订单")
	//_today := time.Now().AddDate(0, 0, -1)
	_str_today_time := php2go.URLEncode(date.FormatDate(time.Now(), "yyyy-MM-dd HH:mm:ss"))
	var _start_time time.Time
	_model_new := &shop.TaobaokeJobtime{Id: 7}
	results, err := DbHelper.MySqlDb().Get(_model_new)
	ErrorHelper.CheckErr(err)
	if results {
		_start_time = _model_new.OkYesterdayStartTime
	}
	_str_start_time := php2go.URLEncode(date.FormatDate(_start_time, "yyyy-MM-dd HH:mm:ss"))
	_str_end_time := php2go.URLEncode(date.FormatDate(_start_time.Add(time.Minute*+20), "yyyy-MM-dd HH:mm:ss"))
	if _str_end_time < _str_today_time { //执行小于今天0点的订单
		//老版接口
		//_http_url := "http://api.vephp.com/order?vekey=V00002504Y26508322&span=1200&order_scene=2&tk_status=1&start_time=" + _str_start_time
		//新版接口
		_http_url := "http://api.vephp.com/orderdetails?vekey=V00002504Y26508322&query_type=3&start_time=" + _str_start_time + "&end_time=" + _str_end_time + "&order_scene=2&tk_status=3" //tk_status淘客订单状态：12-付款，13-关闭，14-确认收货，3-结算成功;不传，表示所有状态。
		//fmt.Println(_http_url)
		_req := Self_Get(_http_url, nil)
		//	fmt.Println(_req) Order_Model
		_order_model := Order_Model{}
		json.Unmarshal([]byte(_req), &_order_model)
		json_list := _order_model.Data.Results.PublisherOrderDto
		// json.Unmarshal(json_data4, &json_list)
		if len(json_list) > 0 {
			for _, v := range json_list {
				fmt.Println(v)
				_CreateTime, _ := date.ParseAny(v.TkPaidTime)
				_ClickTime, _ := date.ParseAny(v.ClickTime)
				_UserId := strconv.FormatInt(v.RelationID, 10)
				_RelationID := strconv.FormatInt(v.RelationID, 10)
				_Buyuser, _UserId := Get_userid_userpid(_RelationID) //根据RelationID取购买人user_id和父id
				is_ok, ostatus := Is_Order_Exist(v.TradeID)
				if is_ok { //修改订单为结算状态,并开始执行分润
					if ostatus == 0 {
						ErrorHelper.LogInfo("获取的要分润的id及父id[?-?]", _Buyuser, _UserId)
						_model_update := &shop.TaobaokeRecord{UserId: _UserId, Buyuser: _Buyuser, Status: v.TkStatus, ClickTime: _ClickTime, OStatus: 1}
						_, err := DbHelper.MySqlDb().Where(" sno=? ", v.TradeID).Cols("user_id", "buyuser", "status", "click_time", "o_status").Update(_model_update)
						ErrorHelper.CheckErr(err)
						ErrorHelper.LogInfo("修改结算采集淘宝客新订单【" + v.TradeID + "】成功！接下来开始执行分润：")
						_http_fr_url := "https://shop.xhdncppf.com/index.php?module=app&action=index&app=calc_ThreeApi_Profit&t_type=2&user_id=" + _Buyuser + "&order_no=" + v.TradeID + "&amount_rate=" + v.TotalCommissionRate + "&amount=" + v.PubSharePreFee + "&mark=" + v.ItemTitle
						_req_fr := Self_Get(_http_fr_url, nil)
						ErrorHelper.LogInfo("【"+v.TradeID+"】分润结果为：", _req_fr)
					}

				} else { //新增订单
					_model_insert := &shop.TaobaokeRecord{StoreId: 8, UserId: _UserId, Buyuser: _Buyuser, Sno: v.TradeID, GoodsTitle: v.ItemTitle, GoodsImg: "http:" + v.ItemImg, GoodsPrice: v.ItemPrice, Num: v.ItemNum, Amount: v.AlipayTotalPrice, CommRate: v.TotalCommissionRate, CommAmount: v.PubSharePreFee, Status: v.TkStatus, CreateTime: _CreateTime, ClickTime: _ClickTime, OStatus: 1}
					_, err := DbHelper.MySqlDb().Insert(_model_insert)
					ErrorHelper.CheckErr(err)
					ErrorHelper.LogInfo("新增结算采集淘宝客新订单【" + v.TradeID + "】成功！")
				}
			}
		}
		_model_update := &shop.TaobaokeJobtime{Id: 7, OkYesterdayStartTime: _start_time.Add(time.Minute * +15)}
		_, err1 := DbHelper.MySqlDb().Id(_model_update.Id).Cols("ok_yesterday_start_time").Update(_model_update)
		ErrorHelper.CheckErr(err1)
	} else {
		ks_time := time.Date(time.Now().Year(), time.Now().Month(), time.Now().Day()-1, 0, 0, 0, 0, time.Now().Location())
		_model_update := &shop.TaobaokeJobtime{Id: 7, OkYesterdayStartTime: ks_time}
		_, err1 := DbHelper.MySqlDb().Id(_model_update.Id).Cols("ok_yesterday_start_time").Update(_model_update)
		ErrorHelper.CheckErr(err1)
	}
}

/*
实时采集最新的渠道订单(每10分钟采集一次)
*/
func job_qudao_now() {
	fmt.Println("开始采集所有渠道订单")
	_str_start_time := php2go.URLEncode(date.FormatDate(time.Now().Add(time.Minute*-11), "yyyy-MM-dd HH:mm:ss"))
	_str_end_time := php2go.URLEncode(date.FormatDate(time.Now(), "yyyy-MM-dd HH:mm:ss"))
	//if _str_start_time < _str_end_time {
	//新版接口
	_http_url := "http://api.vephp.com/orderdetails?vekey=V00002504Y26508322&start_time=" + _str_start_time + "&end_time=" + _str_end_time + "&order_scene=2&tk_status=" //tk_status淘客订单状态：12-付款，13-关闭，14-确认收货，3-结算成功;不传，表示所有状态。默认为全部订单
	_req := Self_Get(_http_url, nil)
	_order_model := Order_Model{}
	json.Unmarshal([]byte(_req), &_order_model)
	json_list := _order_model.Data.Results.PublisherOrderDto
	if len(json_list) > 0 {
		for _, v := range json_list {
			fmt.Println(v)
			_CreateTime, _ := date.ParseAny(v.TkPaidTime)
			_ClickTime, _ := date.ParseAny(v.ClickTime)
			_RelationID := strconv.FormatInt(v.RelationID, 10)
			_Buyuser, _UserId := Get_userid_userpid(_RelationID) //根据RelationID取购买人user_id和父id
			is_ok, _ := Is_Order_Exist(v.TradeID)
			if !is_ok { //订单不存在立即采集
				_model_insert := &shop.TaobaokeRecord{StoreId: 8, UserId: _UserId, Buyuser: _Buyuser, Sno: v.TradeID, GoodsTitle: v.ItemTitle, GoodsImg: "http:" + v.ItemImg, GoodsPrice: v.ItemPrice, Num: v.ItemNum, Amount: v.AlipayTotalPrice, CommRate: v.TotalCommissionRate, CommAmount: v.PubSharePreFee, Status: v.TkStatus, CreateTime: _CreateTime, ClickTime: _ClickTime, OStatus: v.TkStatus}
				_, err := DbHelper.MySqlDb().Insert(_model_insert)
				ErrorHelper.CheckErr(err)
				ErrorHelper.LogInfo("实时采集淘宝客订单【" + v.TradeID + "】成功！")
			}
		}
	}
}

/*
根据订单编号查询订单是否存在
*/
func Is_Order_Exist(_Sno string) (bool, int) {
	//_ok := false
	_model_order := &shop.TaobaokeRecord{Sno: _Sno}
	_ok, err := DbHelper.MySqlDb().Get(_model_order)
	ErrorHelper.CheckErr(err)
	return _ok, _model_order.OStatus
}

/*
根据渠道编号获取用户user_id及父id
*/
func Get_userid_userpid(_RelationId string) (string, string) {
	_rt1, _rt2 := "", ""
	_model_user := &shop.User{RelationId: _RelationId}
	_ok, err := DbHelper.MySqlDb().Get(_model_user)
	ErrorHelper.CheckErr(err)
	if _ok {
		_rt1 = _model_user.UserId
		_rt2 = _model_user.Referee
	}
	return _rt1, _rt2
}

/*
 get提交
 _map 提交参数
*/
func Self_Get(_http_url string, _map map[string]interface{}) string {
	_HeaderData := map[string]string{"Content-Type": "application/json"}
	_req := WebHelper.HttpGet(_http_url, _HeaderData, _map)
	return _req
}

/*
订单实体
*/
type Order_Model struct {
	Error string     `json:"error"`
	Msg   string     `json:"msg"`
	Data  Data_Model `json:"data"`
}

type Data_Model struct {
	HasNext       bool          `json:"has_next"`
	HasPre        bool          `json:"has_pre"`
	PageNo        int           `json:"page_no"`
	PageSize      int           `json:"page_size"`
	PositionIndex string        `json:"position_index"`
	Results       Results_Model `json:"results"`
}
type Results_Model struct {
	PublisherOrderDto []PublisherOrderDto_Model `json:"publisher_order_dto"`
}

type PublisherOrderDto_Model struct {
	AdzoneID         int64  `json:"adzone_id"`
	AdzoneName       string `json:"adzone_name"`
	AlimamaRate      string `json:"alimama_rate"`
	AlimamaShareFee  string `json:"alimama_share_fee"`
	AlipayTotalPrice string `json:"alipay_total_price"`
	ClickTime        string `json:"click_time"`
	FlowSource       string `json:"flow_source"`
	IncomeRate       string `json:"income_rate"`
	ItemCategoryName string `json:"item_category_name"`
	ItemID           int64  `json:"item_id"`
	ItemImg          string `json:"item_img"`
	ItemLink         string `json:"item_link"`
	ItemNum          int    `json:"item_num"`
	ItemPrice        string `json:"item_price"`
	ItemTitle        string `json:"item_title"`
	OrderType        string `json:"order_type"`
	PubID            int    `json:"pub_id"`
	PubShareFee      string `json:"pub_share_fee"`
	PubSharePreFee   string `json:"pub_share_pre_fee"`
	PubShareRate     string `json:"pub_share_rate"`
	RefundTag        int    `json:"refund_tag"`
	RelationID       int64  `json:"relation_id"`
	SellerNick       string `json:"seller_nick"`
	SellerShopTitle  string `json:"seller_shop_title"`
	SiteID           int    `json:"site_id"`
	SiteName         string `json:"site_name"`
	SubsidyFee       string `json:"subsidy_fee"`
	SubsidyRate      string `json:"subsidy_rate"`
	SubsidyType      string `json:"subsidy_type"`
	//订单付款的时间，该时间同步淘宝，可能会略晚于买家在淘宝的订单创建时间
	TbPaidTime                         string `json:"tb_paid_time"`
	TerminalType                       string `json:"terminal_type"`
	TkCommissionFeeForMediaPlatform    string `json:"tk_commission_fee_for_media_platform"`
	TkCommissionPreFeeForMediaPlatform string `json:"tk_commission_pre_fee_for_media_platform"`
	TkCommissionRateForMediaPlatform   string `json:"tk_commission_rate_for_media_platform"`
	//订单创建的时间，该时间同步淘宝，可能会略晚于买家在淘宝的订单创建时间
	TkCreateTime string `json:"tk_create_time"`
	TkOrderRole  int    `json:"tk_order_role"`
	//订单在淘宝拍下付款的时间
	TkPaidTime          string `json:"tk_paid_time"`
	TkStatus            int    `json:"tk_status"`
	TkTotalRate         string `json:"tk_total_rate"`
	TotalCommissionFee  string `json:"total_commission_fee"`
	TotalCommissionRate string `json:"total_commission_rate"`
	TradeID             string `json:"trade_id"`
	TradeParentID       string `json:"trade_parent_id"`
}
