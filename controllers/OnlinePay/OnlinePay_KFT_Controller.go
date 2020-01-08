package OnlinePay

import (
	"encoding/json"
	"fmt"
	"math"
	"strconv"
	"time"
	"qnsoft/web_api/controllers/OnlinePay/Kuaifutong"
	"qnsoft/web_api/controllers/Token"
	"qnsoft/web_api/models/shop"
	"qnsoft/web_api/utils/ArryHelper"
	"qnsoft/web_api/utils/DateHelper"
	"qnsoft/web_api/utils/DbHelper"
	"qnsoft/web_api/utils/ErrorHelper"
	php2go "qnsoft/web_api/utils/Php2go"
	"qnsoft/web_api/utils/StringHelper"
	"qnsoft/web_api/utils/onlinepay"

	_ "github.com/mattn/go-adodb"
)

/**
*信息实体
 */
type OnlinePay_KFT_Controller struct {
	Token.BaseController
}

/*
测试接口1
*/
func (this *OnlinePay_KFT_Controller) XYK_Demo1() {
	//_Model_User := new(model.User)
	//fmt.Println(_Model_User)

	//this.Data["old"] = map[string]string{"name": Old_Query()}
	//this.Data["new"] = map[string]string{"name": New_Query()}
	_kuaifutong := Kuaifutong.KuaiPayHelper{}
	sss := _kuaifutong.Gbp_pay(10)
	//	fmt.Println(sss)
	ErrorHelper.LogInfo(sss)
	this.Data["json"] = sss
	this.ServeJSON()
}

/*
信用卡3.5单笔付款接口
*/
func (this *OnlinePay_KFT_Controller) XYK_35() {
	_Amount, err := this.GetFloat("Amount") //金额（元转分）
	ErrorHelper.CheckErr(err)
	_CustBankNo := this.GetString("CustBankNo")               //客户银行账户行别
	_CustBankAccountNo := this.GetString("CustBankAccountNo") //往客户的哪张卡上付钱
	_CustName := this.GetString("CustName")                   //客户姓名
	_CustID := this.GetString("CustID")                       //客户证件号码
	_RateAmount, err := this.GetFloat("RateAmount")           //手续费（元转分）

	ErrorHelper.CheckErr(err)
	str_order := date.FormatDate(time.Now(), "yyyyMMddHHmmss") + StringHelper.GetRandomNum(6)
	_model_35 := Kuaifutong.Model_35{
		OrderNo:   str_order, //订单编号 用于标识商户发起的一笔交易,在批量交易中,此编号可写在批量请求文件中,用于标识批量请求中的每一笔交易
		TradeName: "签约",      //交易名称 由商户填写,简要概括此次交易的内容.用于在查询交易记录时,提醒用户此次交易具体做了什么事情
		//MerchantBankAccountNo:    "",                                            //商户银行账号 可空 商户用于付款的银行账户，资金到账T+0模式时必填。
		//MerchantBindPhoneNo:      "",                                            //商户开户时绑定的手机号（可空）对于有些银行账户被扣款时，需要提供此绑定手机号才能进行交易；此手机号和短信通知客户的手机号可以一致，也可以不一致
		Amount:                   strconv.FormatFloat(float64(_Amount*100), 'f', 0, 64),     //交易金额 此次交易的具体金额,单位:分,不支持小数点
		CustBankNo:               _CustBankNo,                                               //客户银行账户行别 客户银行账户所属的银行的编号,具体见第5.3.1章节
		CustBankAccountIssuerNo:  "",                                                        //客户开户行网点号 可空 指支付系统里的行号，具体到某个支行（网点）号；
		CustBankAccountNo:        _CustBankAccountNo,                                        //客户银行账户号 本次交易中,往客户的哪张卡上付钱
		CustName:                 _CustName,                                                 //客户姓名 收钱的客户的真实姓名
		CustBankAcctType:         "",                                                        //客户银行账户类型 可空 指客户的银行账户是个人账户还是企业账户
		CustAccountCreditOrDebit: "",                                                        //客户账户借记贷记类型 可空 若是信用卡，则以下两个参数“信用卡有效期”和“信用卡cvv2”； 1借记 2贷记 4 未知
		CustCardValidDate:        "",                                                        //客户信用卡有效期 可空 信用卡的正下方的四位数，前两位是月份，后两位是年份；
		CustCardCvv2:             "",                                                        //客户信用卡的cvv2 可空 信用卡的背面的三位数
		CustID:                   _CustID,                                                   //客户证件号码
		CustPhone:                "",                                                        //客户手机号 如果商户购买的产品中勾选了短信通知功能，则当商户填写了手机号时,快付通会在交易成功后通过短信通知客户
		Messages:                 "",                                                        //发送客户短信内容 可空 如果填写了,则把此参数值的内容发送给客户；如果没填写，则按照快付通的默认模板发送给客户；
		CustEmail:                "",                                                        //客户邮箱地址 可空 如果商户购买的产品中勾选了邮件通知功能，则当商户填写了邮箱地址时,快付通会在交易成功后通过邮件通知客户
		EmailMessages:            "",                                                        //发送客户邮件内容 可空 如果填写了,则把此参数值的内容发送给客户；如果没填写，则按照快付通的默认模板发送给客户；
		Remark:                   "",                                                        //备注 可空 商户可额外填写备注信息,此信息会传给银行,会在银行的账单信息中显示(具体如何显示取决于银行方,快付通不保证银行肯定能显示)
		CustProtocolNo:           "",                                                        //客户协议编号 可空 扣款人在快付通备案的协议号。
		ExtendParams:             "",                                                        //扩展参数 可空 用于商户的特定业务信息传递，只有商户与快付通约定了传递此参数且约定了参数含义，此参数才有效。参数格式：参数名 1^参数值 1|参数名 2^参数值 2|……多条数据用“|”间隔注意: 不能包含特殊字符（如：#、%、&、+）、敏感词汇, 如果必须使用特殊字符,则需要自行做URL Encoding
		RateAmount:               strconv.FormatFloat(float64(_RateAmount*100), 'f', 0, 64), //商户手续费 可空 本次交易需要扣除的手续费。单位:分,不支持小数点 如不填，则手续费默认0元；
	}
	_kuaifutong := Kuaifutong.KuaiPayHelper{}
	sss := _kuaifutong.Gbp_same_id_credit_card_pay(_model_35)
	fmt.Println(sss)
	ErrorHelper.LogInfo(sss)
	this.Data["json"] = sss
	this.ServeJSON()
}

/*
信用卡3.5单笔付款接口
*/
func (this *OnlinePay_KFT_Controller) XYK_35_Small() {
	_Amount, err := this.GetFloat("Amount") //金额（元转分）
	ErrorHelper.CheckErr(err)
	_CustBankNo := this.GetString("CustBankNo")               //客户银行账户行别
	_CustBankAccountNo := this.GetString("CustBankAccountNo") //往客户的哪张卡上付钱
	_CustName := this.GetString("CustName")                   //客户姓名
	_CustID := this.GetString("CustID")                       //客户证件号码
	_RateAmount, err := this.GetFloat("RateAmount")           //手续费（元转分）

	ErrorHelper.CheckErr(err)
	str_order := date.FormatDate(time.Now(), "yyyyMMddHHmmss") + StringHelper.GetRandomNum(6)
	_model_35 := Kuaifutong.Model_35{
		OrderNo:   str_order, //订单编号 用于标识商户发起的一笔交易,在批量交易中,此编号可写在批量请求文件中,用于标识批量请求中的每一笔交易
		TradeName: "签约",      //交易名称 由商户填写,简要概括此次交易的内容.用于在查询交易记录时,提醒用户此次交易具体做了什么事情
		//MerchantBankAccountNo:    "",                                            //商户银行账号 可空 商户用于付款的银行账户，资金到账T+0模式时必填。
		//MerchantBindPhoneNo:      "",                                            //商户开户时绑定的手机号（可空）对于有些银行账户被扣款时，需要提供此绑定手机号才能进行交易；此手机号和短信通知客户的手机号可以一致，也可以不一致
		Amount:                   strconv.FormatFloat(float64(_Amount*100), 'f', 0, 64),     //交易金额 此次交易的具体金额,单位:分,不支持小数点
		CustBankNo:               _CustBankNo,                                               //客户银行账户行别 客户银行账户所属的银行的编号,具体见第5.3.1章节
		CustBankAccountIssuerNo:  "",                                                        //客户开户行网点号 可空 指支付系统里的行号，具体到某个支行（网点）号；
		CustBankAccountNo:        _CustBankAccountNo,                                        //客户银行账户号 本次交易中,往客户的哪张卡上付钱
		CustName:                 _CustName,                                                 //客户姓名 收钱的客户的真实姓名
		CustBankAcctType:         "",                                                        //客户银行账户类型 可空 指客户的银行账户是个人账户还是企业账户
		CustAccountCreditOrDebit: "",                                                        //客户账户借记贷记类型 可空 若是信用卡，则以下两个参数“信用卡有效期”和“信用卡cvv2”； 1借记 2贷记 4 未知
		CustCardValidDate:        "",                                                        //客户信用卡有效期 可空 信用卡的正下方的四位数，前两位是月份，后两位是年份；
		CustCardCvv2:             "",                                                        //客户信用卡的cvv2 可空 信用卡的背面的三位数
		CustID:                   _CustID,                                                   //客户证件号码
		CustPhone:                "",                                                        //客户手机号 如果商户购买的产品中勾选了短信通知功能，则当商户填写了手机号时,快付通会在交易成功后通过短信通知客户
		Messages:                 "",                                                        //发送客户短信内容 可空 如果填写了,则把此参数值的内容发送给客户；如果没填写，则按照快付通的默认模板发送给客户；
		CustEmail:                "",                                                        //客户邮箱地址 可空 如果商户购买的产品中勾选了邮件通知功能，则当商户填写了邮箱地址时,快付通会在交易成功后通过邮件通知客户
		EmailMessages:            "",                                                        //发送客户邮件内容 可空 如果填写了,则把此参数值的内容发送给客户；如果没填写，则按照快付通的默认模板发送给客户；
		Remark:                   "",                                                        //备注 可空 商户可额外填写备注信息,此信息会传给银行,会在银行的账单信息中显示(具体如何显示取决于银行方,快付通不保证银行肯定能显示)
		CustProtocolNo:           "",                                                        //客户协议编号 可空 扣款人在快付通备案的协议号。
		ExtendParams:             "",                                                        //扩展参数 可空 用于商户的特定业务信息传递，只有商户与快付通约定了传递此参数且约定了参数含义，此参数才有效。参数格式：参数名 1^参数值 1|参数名 2^参数值 2|……多条数据用“|”间隔注意: 不能包含特殊字符（如：#、%、&、+）、敏感词汇, 如果必须使用特殊字符,则需要自行做URL Encoding
		RateAmount:               strconv.FormatFloat(float64(_RateAmount*100), 'f', 0, 64), //商户手续费 可空 本次交易需要扣除的手续费。单位:分,不支持小数点 如不填，则手续费默认0元；
	}
	_kuaifutong := Kuaifutong.KuaiPayHelper{}
	sss := _kuaifutong.Gbp_same_id_credit_card_pay_small(_model_35)
	fmt.Println(sss)
	ErrorHelper.LogInfo(sss)
	this.Data["json"] = sss
	this.ServeJSON()
}

/*
信用卡3.6快捷协议代扣协议申请接口(协议和绑卡都来用这个接口，协议成功后将卡绑定在表【lkt_user_bank_card】中)
*/
func (this *OnlinePay_KFT_Controller) XYK_36() {

	_BankType := this.GetString("BankType")                   //银行行别
	_TreatyType := this.GetString("TreatyType")               //协议类型：11：借记卡扣款 12：信用卡扣款
	_Note := this.GetString("Note")                           //说明 参数可空
	_EndDate := this.GetString("EndDate")                     //协议失效日期
	_HolderName := this.GetString("HolderName")               //持卡人真实姓名
	_BankCardType := this.GetString("BankCardType")           //银行卡类型 1、借记卡 2、信用卡
	_BankCardNo := this.GetString("BankCardNo")               //银行卡号
	_MobileNo := this.GetString("MobileNo")                   //预留手机号码
	_CertificateNo := this.GetString("CertificateNo")         //证件号
	_CustCardValidDate := this.GetString("CustCardValidDate") //客户信用卡有效期 如果协议类型为12时不可为空
	_CustCardCvv2 := this.GetString("CustCardCvv2")           //客户信用卡的cvv2 如果协议类型为12时不可为空
	_model_36 := Kuaifutong.Model_36{
		TreatyType:        _TreatyType,
		Note:              _Note,                                   //说明
		StartDate:         date.FormatDate(time.Now(), "yyyyMMdd"), //协议生效日 //根据当前系统自动生成！
		EndDate:           _EndDate,
		HolderName:        _HolderName,
		BankType:          _BankType,
		BankCardType:      _BankCardType,
		BankCardNo:        _BankCardNo,
		MobileNo:          _MobileNo,
		CertificateType:   "0", //证件类型：0表示身份证
		CertificateNo:     _CertificateNo,
		CustCardValidDate: _CustCardValidDate,
		CustCardCvv2:      _CustCardCvv2,
	}
	_kuaifutong := Kuaifutong.KuaiPayHelper{}
	sss := _kuaifutong.Gbp_same_id_treaty_collect_apply(_model_36)
	fmt.Println(sss)
	ErrorHelper.LogInfo(sss)
	this.Data["json"] = sss
	this.ServeJSON()
}

/*
信用卡3.7快捷协议代扣协议确定接口
商户平台通过此接口确认开通代扣协议，进行四要素鉴权获取进行快捷协议代扣的协议号
*/
func (this *OnlinePay_KFT_Controller) XYK_37() {
	_OrderNo := this.GetString("OrderNo")                     //订单编号
	_SmsSeq := this.GetString("SmsSeq")                       //短信流水号
	_AuthCode := this.GetString("AuthCode")                   //手机动态校验码
	_HolderName := this.GetString("HolderName")               //持卡人真实姓名
	_BankCardNo := this.GetString("BankCardNo")               //银行卡号
	_CustCardValidDate := this.GetString("CustCardValidDate") //客户信用卡有效期
	_CustCardCvv2 := this.GetString("CustCardCvv2")           //客户信用卡的cvv2
	_TreatyType := this.GetString("TreatyType")               //协议类型：11：借记卡扣款 12：信用卡扣款
	_model_37 := Kuaifutong.Model_37{
		OrderNo:           _OrderNo,
		SmsSeq:            _SmsSeq,
		AuthCode:          _AuthCode,
		HolderName:        _HolderName,
		BankCardNo:        _BankCardNo,
		CustCardValidDate: _CustCardValidDate,
		CustCardCvv2:      _CustCardCvv2,
	}
	//协议类型
	TreatyType := _TreatyType
	_kuaifutong := Kuaifutong.KuaiPayHelper{}
	sss := _kuaifutong.Gbp_same_id_confirm_treaty_collect_apply(_model_37, TreatyType)
	fmt.Println(sss)
	ErrorHelper.LogInfo(sss)
	this.Data["json"] = sss
	this.ServeJSON()
}

/*
信用卡3.8快捷协议代扣接口
此接口用于商户平台协议代扣。此接口需要用户先签定协议。
*/
func (this *OnlinePay_KFT_Controller) XYK_38() {
	_OrderNo := this.GetString("OrderNo")     //订单编号
	_TreatyNo := this.GetString("TreatyNo")   //协议号
	_TradeTime := this.GetString("TradeTime") //商户方交易时间
	_Amount, err := this.GetFloat("Amount")   //金额（元转分）
	ErrorHelper.CheckErr(err)
	_CustAccountId := this.GetString("CustAccountId")                 //账户ID
	_HolderName := this.GetString("HolderName")                       //持卡人真实姓名
	_BankType := this.GetString("BankType")                           //银行行别
	_BankCardNo := this.GetString("BankCardNo")                       //银行卡号
	_ExtendParams := this.GetString("ExtendParams")                   //扩展字段
	_MerchantBankAccountNo := this.GetString("MerchantBankAccountNo") //商户银行账户
	_RateAmount, err := this.GetFloat("RateAmount")                   //商户手续费(元)
	ErrorHelper.CheckErr(err)
	_CustCardValidDate := this.GetString("CustCardValidDate") //客户信用卡有效期
	_CustCardCvv2 := this.GetString("CustCardCvv2")           //客户信用卡的cvv2
	_NotifyUrl := this.GetString("NotifyUrl")                 //商户后台通知URL
	_CityCode := this.GetString("CityCode", "4910")           //城市编码
	//公网IP地址
	_SourceIP := ""
	_DeviceID := this.GetString("DeviceID") //设备标识
	_merchantNo := Kuaifutong.R1_merchantNo
	if _Amount <= 1000 {
		_RateAmount = _Amount * 0.0032
		_SourceIP = "39.97.111.217"
		_merchantNo = Kuaifutong.R3_merchantNo
	}

	_model_38 := Kuaifutong.Model_38{
		OrderNo:               _OrderNo,
		TreatyNo:              _TreatyNo,
		TradeTime:             _TradeTime,
		Amount:                strconv.Itoa(int(_Amount)),
		CustAccountId:         _CustAccountId,
		HolderName:            _HolderName,
		BankType:              _BankType,
		BankCardNo:            _BankCardNo,
		ExtendParams:          _ExtendParams,
		MerchantBankAccountNo: _MerchantBankAccountNo,
		RateAmount:            strconv.Itoa(int(_RateAmount)),
		CustCardValidDate:     _CustCardValidDate,
		CustCardCvv2:          _CustCardCvv2,
		NotifyUrl:             _NotifyUrl,
		CityCode:              _CityCode,
		SourceIP:              _SourceIP,
		DeviceID:              _DeviceID,
	}
	//协议类型
	TreatyType := "12"
	_kuaifutong := Kuaifutong.KuaiPayHelper{}
	sss := _kuaifutong.Gbp_same_id_credit_card_treaty_collect(_model_38, TreatyType, _merchantNo)
	fmt.Println(sss)
	ErrorHelper.LogInfo(sss)
	this.Data["json"] = sss
	this.ServeJSON()
}

/*
*3.9快捷协议代扣协议查询接口
*此接口用于商户平台通过此接口查询协议信息。
 */
func (this *OnlinePay_KFT_Controller) XYK_39() {
	//订单编号
	_OrderNo := this.GetString("OrderNo")
	//协议号
	_TreatyNo := this.GetString("TreatyNo")

	_model_39 := Kuaifutong.Model_39{
		OrderNo:  _OrderNo,
		TreatyNo: _TreatyNo,
	}
	_kuaifutong := Kuaifutong.KuaiPayHelper{}
	sss := _kuaifutong.Gbp_query_treaty_info(_model_39)
	fmt.Println(sss)
	ErrorHelper.LogInfo(sss)
	this.Data["json"] = sss
	this.ServeJSON()
}

/*
*3.10快捷协议代扣协议解除接口
*此接口用于商户平台通过此接口解除快捷协议收款协议信息。
 */
func (this *OnlinePay_KFT_Controller) XYK_310() {
	//订单编号
	_OrderNo := this.GetString("OrderNo")
	//协议号
	_TreatyNo := this.GetString("TreatyNo")

	_model_310 := Kuaifutong.Model_310{
		OrderNo:  _OrderNo,
		TreatyNo: _TreatyNo,
	}
	_kuaifutong := Kuaifutong.KuaiPayHelper{}
	sss := _kuaifutong.Gbp_cancel_treaty_info(_model_310)
	fmt.Println(sss)
	ErrorHelper.LogInfo(sss)
	this.Data["json"] = sss
	this.ServeJSON()
}

/*
信用卡3.11用户资金查询接口(对应我的余额时custID必传)
*/
func (this *OnlinePay_KFT_Controller) XYK_311() {
	_ReqNo := this.GetString("reqNo")
	_CustID := this.GetString("custID")
	_PageNum := this.GetString("pageNum")
	_model_311 := Kuaifutong.Model_311{
		ReqNo:   _ReqNo,
		CustID:  _CustID,
		PageNum: _PageNum,
	}
	_kuaifutong := Kuaifutong.KuaiPayHelper{}
	sss := _kuaifutong.Gbp_same_id_credit_card_not_pay_balance(_model_311)
	fmt.Println(sss)
	ErrorHelper.LogInfo(sss)
	this.Data["json"] = sss
	this.ServeJSON()
}

/*
信用卡3.11用户资金查询接口(对应我的余额时custID必传)(小额)
*/
func (this *OnlinePay_KFT_Controller) XYK_311_Small() {
	_ReqNo := this.GetString("reqNo")
	_CustID := this.GetString("custID")
	_PageNum := this.GetString("pageNum")
	_model_311 := Kuaifutong.Model_311{
		ReqNo:   _ReqNo,
		CustID:  _CustID,
		PageNum: _PageNum,
	}
	_kuaifutong := Kuaifutong.KuaiPayHelper{}
	sss := _kuaifutong.Gbp_same_id_credit_card_not_pay_balance_small(_model_311)
	fmt.Println(sss)
	ErrorHelper.LogInfo(sss)
	this.Data["json"] = sss
	this.ServeJSON()
}

/*
信用卡3.12交易查询接口
* 3.12交易查询接口
*用于查询指定的一笔或多笔交易的结果,例如购买支付交易状态
*/
func (this *OnlinePay_KFT_Controller) XYK_312() {
	_StartDate := this.GetString("StartDate")
	_EndDate := this.GetString("EndDate")
	_OrderNo := this.GetString("OrderNo")
	_TradeType := this.GetString("TradeType")
	_Status := this.GetString("Status")
	_model_312 := Kuaifutong.Model_312{
		StartDate: _StartDate,
		EndDate:   _EndDate,
		OrderNo:   _OrderNo,
		TradeType: _TradeType,
		Status:    _Status,
	}
	_kuaifutong := Kuaifutong.KuaiPayHelper{}
	sss := _kuaifutong.Gbp_same_id_credit_card_trade_record_query(_model_312)
	fmt.Println(sss)
	ErrorHelper.LogInfo(sss)
	this.Data["json"] = sss
	this.ServeJSON()
}

/*
信用卡3.13账户余额查询接口
*/
func (this *OnlinePay_KFT_Controller) XYK_313() {

	_kuaifutong := Kuaifutong.KuaiPayHelper{}
	sss := _kuaifutong.Query_available_balance()
	fmt.Println(sss)
	ErrorHelper.LogInfo(sss)
	this.Data["json"] = sss
	this.ServeJSON()
}

/*
*3.14交易类对账文件获取接口
此功能用于给商户提供对账数据。
*/
func (this *OnlinePay_KFT_Controller) XYK_314() {
	_CheckDate := this.GetString("CheckDate")
	_model_314 := Kuaifutong.Model_314{
		CheckDate: _CheckDate,
	}
	_kuaifutong := Kuaifutong.KuaiPayHelper{}
	sss := _kuaifutong.Gbp_same_id_credit_card_recon_result_query(_model_314)
	fmt.Println(sss)
	ErrorHelper.LogInfo(sss)
	this.Data["json"] = sss
	this.ServeJSON()
}

/*
*3.15银行卡三要素验证接口
*此接口用于校验指定的银行卡和用户身份信息是否匹配及正确，
 */
func (this *OnlinePay_KFT_Controller) XYK_315() {
	_CustBankNo := this.GetString("CustBankNo")
	_CustName := this.GetString("CustName")
	_CustBankAccountNo := this.GetString("CustBankAccountNo")
	_CustAccountCreditOrDebit := this.GetString("CustAccountCreditOrDebit")
	_CustCertificationType := this.GetString("CustCertificationType")
	_CustID := this.GetString("CustID")
	_model_315 := Kuaifutong.Model_315{
		CustBankNo:               _CustBankNo,
		CustName:                 _CustName,
		CustBankAccountNo:        _CustBankAccountNo,
		CustAccountCreditOrDebit: _CustAccountCreditOrDebit,
		CustCertificationType:    _CustCertificationType,
		CustID:                   _CustID,
	}
	_kuaifutong := Kuaifutong.KuaiPayHelper{}
	orderStarWith := ""
	sss := _kuaifutong.Gbp_threeMessage_verification(_model_315, orderStarWith)
	fmt.Println(sss)
	ErrorHelper.LogInfo(sss)
	this.Data["json"] = sss
	this.ServeJSON()
}

/*
3.16验证类对账文件获取接口
此功能用于给商户提供对账数据。
*/
func (this *OnlinePay_KFT_Controller) XYK_316() {
	_CheckDate := this.GetString("CheckDate")
	_model_316 := Kuaifutong.Model_316{
		CheckDate: _CheckDate,
	}
	_kuaifutong := Kuaifutong.KuaiPayHelper{}
	sss := _kuaifutong.Gbp_same_id_credit_card_verify_result_query(_model_316)
	fmt.Println(sss)
	ErrorHelper.LogInfo(sss)
	this.Data["json"] = sss
	this.ServeJSON()
}

/*
3.17结算单笔付款接口
单笔付款在功能和接口参数上与单笔收款基本一致,只是付款方变成了商户自己,收款方变成了商户指定的客户
*/
func (this *OnlinePay_KFT_Controller) XYK_317() {
	_Money, err := this.GetFloat("Money")
	ErrorHelper.CheckErr(err)
	_kuaifutong := Kuaifutong.KuaiPayHelper{}
	sss := _kuaifutong.Gbp_pay(_Money)
	//	fmt.Println(sss)
	ErrorHelper.LogInfo(sss)
	this.Data["json"] = sss
	this.ServeJSON()
}

/*
3.18单笔付款协议申请接口
此接口用于平台对商户通过网络申请单笔收款的协议，方便商户提前备案协议。
此接口返回的协议号，用于单笔收款接口代扣时验证客户信息。
*/
func (this *OnlinePay_KFT_Controller) XYK_318() {
	_kuaifutong := Kuaifutong.KuaiPayHelper{}
	sss := _kuaifutong.Gbp_send_treaty_record_to_kft()
	//	fmt.Println(sss)
	ErrorHelper.LogInfo(sss)
	this.Data["json"] = sss
	this.ServeJSON()
}

/*
3.19单笔付款协议查询接口
此接口用于平台对商户通过网络线上查询单笔收款协议状态。
*/
func (this *OnlinePay_KFT_Controller) XYK_319() {
	orderNumber := this.GetString("orderNumber")
	_kuaifutong := Kuaifutong.KuaiPayHelper{}
	sss := _kuaifutong.Gbp_query_treaty_record_info(orderNumber)
	//	fmt.Println(sss)
	ErrorHelper.LogInfo(sss)
	this.Data["json"] = sss
	this.ServeJSON()
}

/*
3.18单笔付款协议申请接口
此接口用于平台对商户通过网络申请单笔收款的协议，方便商户提前备案协议。
此接口返回的协议号，用于单笔收款接口代扣时验证客户信息。
*/
func (this *OnlinePay_KFT_Controller) XYK_322() {
	_kuaifutong := Kuaifutong.KuaiPayHelper{}
	sss := _kuaifutong.Query_available_balance_A()
	//	fmt.Println(sss)
	ErrorHelper.LogInfo(sss)
	this.Data["json"] = sss
	this.ServeJSON()
}

/*
公众号银行列表
*/
func (this *OnlinePay_KFT_Controller) Bind_Bank_List() {
	_model := new(shop.ChannelBankKft)
	rows, err := DbHelper.MySqlDb().SQL("select * from lkt_channel_bank_kft ").Rows(_model)
	ErrorHelper.CheckErr(err)
	defer rows.Close()
	_list := make([]*shop.ChannelBankKft, 0)
	for rows.Next() {
		_ = rows.Scan(_model)
		_model_new := new(shop.ChannelBankKft)
		_model_new.Id = _model.Id
		_model_new.Bankcode = _model.Bankcode
		_model_new.Bankname = _model.Bankname
		_model_new.D0freerate = _model.D0freerate
		_model_new.D0fixrate = _model.D0fixrate
		_model_new.T1freerate = _model.T1freerate
		_model_new.T1fixrate = _model.T1fixrate
		_model_new.Channelid = _model.Channelid
		_model_new.D0myrate = _model.D0myrate
		_list = append(_list, _model_new)
	}
	this.Data["json"] = _list
	this.ServeJSON()
}

/*
公众号绑卡操作
*/
func (this *OnlinePay_KFT_Controller) Bind_Card_Add() {
	var _json_model interface{}
	_StoreId := 8
	_UserId := this.GetString("UserId")                 //用户id
	_Cardholder := this.GetString("Cardholder")         //用户真实姓名
	_IdCard := this.GetString("IdCard")                 //用户身份证号
	_BankCode := this.GetString("BankCode")             //银行行别编号
	_BankName := this.GetString("BankName")             //银行名称
	_BankCardNumber := this.GetString("BankCardNumber") //银行卡卡号
	_Mobile := this.GetString("Mobile")                 //手机号
	_CardType, _ := this.GetInt("CardType")             //卡类型：1储蓄卡 2信用卡
	_Expiretime := this.GetString("Expiretime")         //信用卡有效期
	_Cvv2 := this.GetString("Cvv2")                     //信用卡Cvv2
	//_Iscashonly := this.GetString("Iscashonly")         //现金标志
	//开始三要素验签名
	_model_315 := Kuaifutong.Model_315{
		CustBankNo:               _BankCode,       //银行行别编号
		CustName:                 _Cardholder,     //持卡人
		CustBankAccountNo:        _BankCardNumber, //卡号
		CustAccountCreditOrDebit: "1",             //客户的银行账户是借记类型还是贷记类型 0存折 1借记 2贷记
		CustCertificationType:    "0",
		CustID:                   _IdCard,
	}
	_kuaifutong := Kuaifutong.KuaiPayHelper{}
	orderStarWith := ""
	_return_15 := _kuaifutong.Gbp_threeMessage_verification(_model_315, orderStarWith)
	if _return_15.Status == "1" {
		_model := &shop.UserBankCard{UserId: _UserId, BankCardNumber: _BankCardNumber}
		results, _ := DbHelper.MySqlDb().Where("mobile<>?", "").Get(_model)
		if results { //如果当前卡已经绑定，直接修改
			_update_model := shop.UserBankCard{
				UserId:     _UserId,
				StoreId:    _StoreId,
				Cardholder: _Cardholder,
				IdCard:     _IdCard,
				BankCode:   _BankCode,
				BankName:   _BankName,
				//Branch :      _Branch,
				BankCardNumber: _BankCardNumber,
				Mobile:         _Mobile,
				AddDate:        time.Now(),
				//	IsDefault :      _IsDefault,
				//	MchId    :      _MchId,
				CardType: _CardType,
				//Token       :      _Token,
				Expiretime: _Expiretime,
				Cvv2:       _Cvv2,
				//Treatyid     :      _Treatyid,
				//Treatytype    :      _Treatytype,
				//Treatyenddate :      _Treatyenddate,
				//Iscashonly: _Iscashonly,
				//Delflag: 0,
				//Orderid       :      _Orderid,
				State: 1,
			}

			_, err := DbHelper.MySqlDb().Id(_model.Id).Update(&_update_model)
			ErrorHelper.CheckErr(err)
			_json_model = map[string]interface{}{"code": 1, "msg": "success", "info": "绑卡成功！"}

			///实名认证开始
			if _CardType == 1 { //如果是储蓄卡直接提交实名认证
				_model := &shop.UserAuth{UserId: _UserId}
				results, _ := DbHelper.MySqlDb().Get(_model)
				if results { //如果用户已实名，无需再操作
					if _model.Ysbflag != 1 { //执行修改
						_update_model := shop.UserAuth{
							UserId:        _UserId,
							Truename:      _Cardholder,
							Idcard:        _IdCard,
							Personimg:     "",
							Personimgback: "",
							Authflag:      1,
							Ysbflag:       1,
							Addtime:       time.Now(),
						}
						_, err := DbHelper.MySqlDb().Id(_model.Id).Update(&_update_model)
						ErrorHelper.CheckErr(err)
						_json_model = map[string]interface{}{"code": 1, "msg": "success", "info": "认证已提交，等待审核！"}
					}
				} else { //执行添加
					_new_model := shop.UserAuth{
						UserId:        _UserId,
						Truename:      _Cardholder,
						Idcard:        _IdCard,
						Personimg:     "",
						Personimgback: "",
						Authflag:      1,
						Ysbflag:       1,
						Addtime:       time.Now(),
					}
					_, err := DbHelper.MySqlDb().Insert(_new_model)
					ErrorHelper.CheckErr(err)
					_json_model = map[string]interface{}{"code": 1, "msg": "success", "info": "认证已审核！"}
				}
			}
			///实名认证结束

		} else { //执行添加
			_new_model := shop.UserBankCard{
				UserId:     _UserId,
				StoreId:    _StoreId,
				Cardholder: _Cardholder,
				IdCard:     _IdCard,
				BankCode:   _BankCode,
				BankName:   _BankName,
				//Branch :      _Branch,
				BankCardNumber: _BankCardNumber,
				Mobile:         _Mobile,
				AddDate:        time.Now(),
				//	IsDefault :      _IsDefault,
				//	MchId    :      _MchId,
				CardType: _CardType,
				//Token       :      _Token,
				Expiretime: _Expiretime,
				Cvv2:       _Cvv2,

				//Treatyid     :      _Treatyid,
				//Treatytype    :      _Treatytype,
				//Treatyenddate :      _Treatyenddate,
				//Iscashonly: _Iscashonly,
				Delflag: 0,
				//Orderid       :      _Orderid,
				State: 1,
			}
			_count, err := DbHelper.MySqlDb().Insert(&_new_model)
			ErrorHelper.CheckErr(err)
			if _count > 0 {
				_json_model = map[string]interface{}{"code": 1, "msg": "success", "info": "绑卡成功！"}
			} else {
				_json_model = map[string]interface{}{"code": 1, "msg": "success", "info": "新卡添加失败！"}
			}
			/*
				if _count > 0 {
					_model_315 := Kuaifutong.Model_315{
						CustBankNo:               _BankCode,       //银行行别编号
						CustName:                 _Cardholder,     //持卡人
						CustBankAccountNo:        _BankCardNumber, //卡号
						CustAccountCreditOrDebit: "1",             //客户的银行账户是借记类型还是贷记类型 0存折 1借记 2贷记
						CustCertificationType:    "0",
						CustID:                   _IdCard,
					}
					_kuaifutong := Kuaifutong.KuaiPayHelper{}
					orderStarWith := ""
					_return_15 := _kuaifutong.Gbp_threeMessage_verification(_model_315, orderStarWith)
					if _return_15.Status == "1" {
						_json_model = map[string]interface{}{"code": 1, "msg": "success", "info": "绑卡成功！"}
					} else {
						_json_model = map[string]interface{}{"code": 0, "msg": "fail", "info": "户名、卡号、证件号三要素认证失败！"}
					}
				} else {
					_json_model = map[string]interface{}{"code": 0, "msg": "success", "info": "绑卡失败！！"}
				}
			*/
		}
	} else {
		_json_model = map[string]interface{}{"code": 0, "msg": "fail", "info": "绑卡失败！户名、卡号、证件号三要素认证失败！", "info_other": _return_15}
	}
	this.Data["json"] = _json_model
	this.ServeJSON()
}

/*
公众号卡列表
*/
func (this *OnlinePay_KFT_Controller) Bind_Card_List() {
	_UserId := this.GetString("UserId") //用户id
	_model := new(shop.UserBankCard)
	rows, err := DbHelper.MySqlDb().SQL("select * from lkt_user_bank_card where State=1 and user_id='" + _UserId + "'").Rows(_model)
	ErrorHelper.CheckErr(err)
	defer rows.Close()
	_list := make([]*shop.UserBankCard, 0)
	for rows.Next() {
		_ = rows.Scan(_model)
		_model_new := new(shop.UserBankCard)
		_model_new.Id = _model.Id
		_model_new.StoreId = _model.StoreId
		_model_new.UserId = _model.UserId
		_model_new.Cardholder = _model.Cardholder
		_model_new.IdCard = _model.IdCard
		_model_new.BankCode = _model.BankCode
		_model_new.BankName = _model.BankName
		_model_new.Branch = _model.Branch
		_model_new.BankCardNumber = _model.BankCardNumber
		_model_new.Mobile = _model.Mobile
		_model_new.AddDate = _model.AddDate
		_model_new.IsDefault = _model.IsDefault
		_model_new.MchId = _model.MchId
		_model_new.CardType = _model.CardType
		_model_new.Token = _model.Token
		_model_new.Expiretime = _model.Expiretime
		_model_new.Treatyid = _model.Treatyid
		_model_new.TreatyidSmall = _model.TreatyidSmall
		_model_new.Cvv2 = _model.Cvv2
		_model_new.Iscashonly = _model.Iscashonly
		_model_new.Delflag = _model.Delflag           //签约标识
		_model_new.DelflagSmall = _model.DelflagSmall //小额签约标识
		_model_new.Orderid = _model.Orderid
		_model_new.State = _model.State
		_list = append(_list, _model_new)
	}
	this.Data["json"] = _list
	this.ServeJSON()
}

/*
公众号解绑卡操作
*/
func (this *OnlinePay_KFT_Controller) Bind_Card_Relieve() {
	var _json_model interface{}
	_UserId := this.GetString("UserId")                 //用户id
	_BankCardNumber := this.GetString("BankCardNumber") //银行卡卡号
	_model := &shop.UserBankCard{UserId: _UserId, BankCardNumber: _BankCardNumber}
	results, _ := DbHelper.MySqlDb().Where("mobile<>?", "").Get(_model)
	if results { //如果当前卡已经绑定，直接修改
		_State := 0
		if _model.State == 0 {
			_State = 1
		} else {
			_State = 0
		}
		_update_model := shop.UserBankCard{
			//UserId:         _UserId,
			//	BankCardNumber: _BankCardNumber,
			State: _State,
		}
		_count, err := DbHelper.MySqlDb().Id(_model.Id).Cols("state").Update(&_update_model)
		ErrorHelper.CheckErr(err)
		if _count > 0 {
			_json_model = map[string]interface{}{"code": 1, "msg": "success", "info": "解绑卡成功！"}
		} else {
			_json_model = map[string]interface{}{"code": 0, "msg": "success", "info": "解绑失败！"}
		}
	} else {
		_json_model = map[string]interface{}{"code": 0, "msg": "fail", "info": "解绑卡号或用户不存在！"}
	}
	this.Data["json"] = _json_model
	this.ServeJSON()
}

/*
公众号获取用户单卡信息
*/
func (this *OnlinePay_KFT_Controller) Bind_Card_GetOne() {
	var _json_model interface{}
	_UserId := this.GetString("UserId")         //用户id
	_BankCardNo := this.GetString("BankCardNo") //用户卡号
	_model := &shop.UserBankCard{UserId: _UserId, BankCardNumber: _BankCardNo}
	results, _ := DbHelper.MySqlDb().Where("mobile<>?", "").Get(_model)
	if results {
		_json_model = map[string]interface{}{"code": 1, "msg": "success", "data": _model, "info": "获取绑定卡信息成功！"}
	} else {
		_json_model = map[string]interface{}{"code": 0, "msg": "fail", "data": shop.UserBankCard{}, "info": "获取绑定卡信息失败！"}
	}
	this.Data["json"] = _json_model
	this.ServeJSON()
}

/*
*快捷代扣
 */
func (this *OnlinePay_KFT_Controller) Quick_Pay() {
	var _json_model interface{}
	_UserId := this.GetString("UserId")                               //用户id
	_BankCardNo := this.GetString("BankCardNo")                       //用户出金卡号
	_MerchantBankAccountNo := this.GetString("MerchantBankAccountNo") //用户入金卡号
	_Amount, err := this.GetFloat("Amount")                           //代扣金额 手动输入 从提交获取
	ErrorHelper.CheckErr(err)
	_merchantNo := Kuaifutong.R1_merchantNo
	_CityCode := this.GetString("CityCode", "4910") //城市编码 可空
	_SourceIP := this.GetString("SourceIP")         //公网IP地址 可空 小于等于1000该项必传
	_DeviceID := this.GetString("DeviceID")         //设备标识 可空
	ErrorHelper.CheckErr(err)
	_NotifyUrl := Kuaifutong.QuickPay_Pay_notifyUrl //商户后台通知URL
	//根据userid和出金卡号取代扣签约信息

	_model := &shop.UserBankCard{UserId: _UserId, BankCardNumber: _BankCardNo}
	results, _ := DbHelper.MySqlDb().Get(_model)
	//代扣协议id
	_Treatyid := ""
	if results {
		_Treatyid = _model.Treatyid
		_feilvZ_t0, err := strconv.ParseFloat(Get_Bank_T0_FL(_model.UserId), 64)
		ErrorHelper.CheckErr(err)
		_RateAmount := _Amount * _feilvZ_t0 //快付通费率 快付通手续费 总额*费率
		/*小额临时注释 	if _Amount <= 1000 { //计算快付通小额费率 */
		if _Amount <= 1 { //计算快付通小额费率
			_Treatyid = _model.TreatyidSmall
			_RateAmount = _Amount * 0.0032
			_merchantNo = Kuaifutong.R3_merchantNo
			_SourceIP = Kuaifutong.Server_IP
		}
		_model_38 := Kuaifutong.Model_38{
			OrderNo:   "DK" + date.FormatDate(time.Now(), "yyMMddHHmmss") + strconv.Itoa(php2go.Rand(10, 99)), //订单编号 自动生成
			TreatyNo:  _Treatyid,
			TradeTime: date.FormatDate(time.Now(), "yyyyMMddHHmmss"),         //生成当前时间
			Amount:    strconv.FormatFloat(float64(_Amount*100), 'f', 0, 64), //代扣金额
			//CustAccountId: _CustAccountId,                              //账户ID 协议类型 11：借记卡扣款 12：信用卡扣款 13：余额扣款 余额+借记卡扣款15： 余额+信用卡扣款协议代扣申请接口时，如果协议类型为13、14、15时不可为空
			HolderName: _model.Cardholder, //持卡人真实姓名
			BankType:   _model.BankCode,   //银行行别
			BankCardNo: _BankCardNo,       //银行卡号 卡1出金卡
			//	ExtendParams:          _ExtendParams,                    //扩展字段 当商户为二级商户是此字段必填
			MerchantBankAccountNo: _MerchantBankAccountNo,                                    //商户银行账户 卡2入金卡 从传参来 商户用于收款的银行账户 资金到账T+0模式时必填。
			RateAmount:            strconv.FormatFloat(float64(_RateAmount*100), 'f', 0, 64), //商户手续费
			CustCardValidDate:     _model.Expiretime,                                         //客户信用卡有效期
			CustCardCvv2:          _model.Cvv2,                                               //客户信用卡的cvv2
			NotifyUrl:             _NotifyUrl,                                                //商户后台通知URL 写死的常量直接调取
			CityCode:              _CityCode,                                                 //城市编码 可空
			SourceIP:              _SourceIP,                                                 //公网IP地址 可空 小于等于1000该项必传
			DeviceID:              _DeviceID,                                                 //设备标识 可空
		}
		_kuaifutong := Kuaifutong.KuaiPayHelper{}
		_rerurn_38 := _kuaifutong.Gbp_same_id_credit_card_treaty_collect(_model_38, _model.Treatytype, _merchantNo)
		_str_rt_38 := fmt.Sprintf("%+v", _rerurn_38)
		//开始将代扣信息写入到信用卡订单表[lkt_user_card_order]
		_new_model_order := shop.UserCardOrder{
			UserId:                _UserId,                //用户id
			Ordernoa:              _rerurn_38.OrderNo,     //订单编号
			Amount:                int(_Amount * 100),     //订单金额
			Rateamounta:           int(_RateAmount * 100), //代扣手续费
			Bankcardno:            _BankCardNo,            //出金卡号
			Merchantbankaccountno: _MerchantBankAccountNo, //入金卡号
			ReturnA:               _str_rt_38,             //代扣返回信息
			AddTime:               time.Now(),             //订单创建时间
		}
		_count, err := DbHelper.MySqlDb().Insert(_new_model_order)
		ErrorHelper.CheckErr(err)
		if _count > 0 && err == nil {
			//开始将代扣执行结果以json方式返回
			_json_model = map[string]interface{}{"code": _rerurn_38.Status, "msg": "success", "data": _rerurn_38, "info": "代扣交易提交成功！"}
		} else {
			_json_model = map[string]interface{}{"code": 0, "msg": "fail", "info": "代扣订单写入数据库失败！"}
		}

	} else {
		_json_model = map[string]interface{}{"code": 0, "msg": "fail", "info": "该卡还没有签署代扣协议，不能操作！"}
	}
	this.Data["json"] = _json_model
	this.ServeJSON()
}

/*
*代扣status>1开始执行快捷代付
 */
func (this *OnlinePay_KFT_Controller) Quick_Pay_Ok() {
	var _json_model interface{}
	_UserId := this.GetString("UserId")         //用户id
	_OrderNoA := this.GetString("OrderNoA")     //用户id
	_BankCardNo := this.GetString("BankCardNo") //用户出金卡号
	_Bank_code := Get_Bank_CodeA(_BankCardNo)
	_MerchantBankAccountNo := this.GetString("MerchantBankAccountNo") //用户入金卡号
	//根据用户id和银行卡号取用户和卡的相关信息
	_model := &shop.UserBankCard{UserId: _UserId, BankCardNumber: _MerchantBankAccountNo} //这里是入金卡 臻方便费率应该从出金卡扣除
	results, _ := DbHelper.MySqlDb().Get(_model)
	if results {
		//臻方便代付费率
		_feilvZ_t0, err := strconv.ParseFloat(Get_ZFB_T0_FL(_Bank_code), 64) //出金卡费率
		ErrorHelper.CheckErr(err)
		//开始根据userid和代扣订单编号获取订单总金额
		_model_order := &shop.UserCardOrder{
			UserId:   _UserId,   //用户id
			Ordernoa: _OrderNoA, //订单编号
		}
		_ok, err := DbHelper.MySqlDb().Get(_model_order) //获取代扣订单
		ErrorHelper.CheckErr(err)
		if _ok { //如果代扣订单存在 继续往下执行代付
			_Treatyid := ""
			_AmountY := _model_order.Amount //账单总金额
			ErrorHelper.CheckErr(err)
			_Kft_sxf := _model_order.Rateamounta //快付通代扣手续费
			ErrorHelper.CheckErr(err)
			_RateAmount := int(float64(_AmountY) * _feilvZ_t0) //臻方便代付固定费率 总额*费率
			//开始利用3.11接口查询余额
			_model_311 := Kuaifutong.Model_311{
				//	ReqNo:   _ReqNo,
				CustID:  _model.IdCard,
				PageNum: "1",
			}
			_kuaifutong := Kuaifutong.KuaiPayHelper{}
			_rerurn_11 := Kuaifutong.Return_311{}
			/*小额临时注释 if _AmountY <= 100000 && len(_model.TreatyidSmall) > 5 { //计算快付通小额费率 */
			if _AmountY <= 10 && len(_model.TreatyidSmall) > 5 { //计算快付通小额费率
				_Treatyid = _model.TreatyidSmall
				_Kft_sxf = int(math.Floor(float64(_AmountY)*0.0032 + 0.5))
				_RateAmount = int(math.Floor(float64(_AmountY)*0.0025 + 0.5))
				_model_order.Rateamountb = _RateAmount + 100                                       //总手续费
				_rerurn_11 = _kuaifutong.Gbp_same_id_credit_card_not_pay_balance_small(_model_311) //余额查询
			} else {
				_Treatyid = _model.Treatyid
				_rerurn_11 = _kuaifutong.Gbp_same_id_credit_card_not_pay_balance(_model_311) //余额查询
				bank_code := Get_Bank_Model(_model_order.Bankcardno).BankCode                //获取取银行行别代码
				fellv, err := strconv.ParseFloat(Get_ZFB_T0_FL(bank_code), 64)
				ErrorHelper.CheckErr(err)
				_model_order.Rateamountb = int(math.Floor(float64(_model_order.Amount-_model_order.Rateamounta)*fellv+0.5)) + 100
				ErrorHelper.LogInfo("智能代还走大额出金后查询")
			}
			if _rerurn_11.Status == "1" {
				//_Zh_Amount := _rerurn_11.Details[0].BalanceAmount   //获取当前账户在快付通全部余额
				_Amount := _AmountY - _Kft_sxf //本单应总额度减去快付通扣除手续费=提现额度含臻方便代付手续费
				_Last_Amount, err := strconv.Atoi(_rerurn_11.Details[0].BalanceAmount)
				ErrorHelper.CheckErr(err)
				if _Last_Amount < _AmountY-_Kft_sxf {
					_Amount = _Last_Amount
				}
				_RateAmountA := strconv.Itoa(_model_order.Rateamountb) //臻方便费率 //用臻方便自己费率算代付手续费（从数据库获取臻方便固定费率+2元即200分）
				str_order := "DH" + date.FormatDate(time.Now(), "yyyyMMddHHmmss") + StringHelper.GetRandomNum(6)
				ErrorHelper.LogInfo("代付金额：", _Amount)
				ErrorHelper.LogInfo("代付手续费：", _RateAmountA)
				_model_35 := Kuaifutong.Model_35{
					OrderNo:   str_order, //订单编号 用于标识商户发起的一笔交易,在批量交易中,此编号可写在批量请求文件中,用于标识批量请求中的每一笔交易
					TradeName: "代付",      //交易名称 由商户填写,简要概括此次交易的内容.用于在查询交易记录时,提醒用户此次交易具体做了什么事情
					//MerchantBankAccountNo:    "",                                            //商户银行账号 可空 商户用于付款的银行账户，资金到账T+0模式时必填。
					//MerchantBindPhoneNo:      "",                                            //商户开户时绑定的手机号（可空）对于有些银行账户被扣款时，需要提供此绑定手机号才能进行交易；此手机号和短信通知客户的手机号可以一致，也可以不一致
					Amount:                   strconv.Itoa(_Amount),  //交易金额 此次交易的具体金额,单位:分,不支持小数点
					CustBankNo:               _model.BankCode,        //客户银行账户行别 客户银行账户所属的银行的编号,具体见第5.3.1章节
					CustBankAccountIssuerNo:  "",                     //客户开户行网点号 可空 指支付系统里的行号，具体到某个支行（网点）号；
					CustBankAccountNo:        _MerchantBankAccountNo, //客户银行账户号 本次交易中,往客户的哪张卡上付钱
					CustName:                 _model.Cardholder,      //客户姓名 收钱的客户的真实姓名
					CustBankAcctType:         "",                     //客户银行账户类型 可空 指客户的银行账户是个人账户还是企业账户
					CustAccountCreditOrDebit: "",                     //客户账户借记贷记类型 可空 若是信用卡，则以下两个参数“信用卡有效期”和“信用卡cvv2”； 1借记 2贷记 4 未知
					CustCardValidDate:        "",                     //客户信用卡有效期 可空 信用卡的正下方的四位数，前两位是月份，后两位是年份；
					CustCardCvv2:             "",                     //客户信用卡的cvv2 可空 信用卡的背面的三位数
					CustID:                   _model.IdCard,          //客户证件号码
					CustPhone:                _model.Mobile,          //客户手机号 如果商户购买的产品中勾选了短信通知功能，则当商户填写了手机号时,快付通会在交易成功后通过短信通知客户
					Messages:                 "",                     //发送客户短信内容 可空 如果填写了,则把此参数值的内容发送给客户；如果没填写，则按照快付通的默认模板发送给客户；
					CustEmail:                "",                     //客户邮箱地址 可空 如果商户购买的产品中勾选了邮件通知功能，则当商户填写了邮箱地址时,快付通会在交易成功后通过邮件通知客户
					EmailMessages:            "",                     //发送客户邮件内容 可空 如果填写了,则把此参数值的内容发送给客户；如果没填写，则按照快付通的默认模板发送给客户；
					Remark:                   "",                     //备注 可空 商户可额外填写备注信息,此信息会传给银行,会在银行的账单信息中显示(具体如何显示取决于银行方,快付通不保证银行肯定能显示)
					CustProtocolNo:           _Treatyid,              //客户协议编号 可空 扣款人在快付通备案的协议号。
					ExtendParams:             "",                     //扩展参数 可空 用于商户的特定业务信息传递，只有商户与快付通约定了传递此参数且约定了参数含义，此参数才有效。参数格式：参数名 1^参数值 1|参数名 2^参数值 2|……多条数据用“|”间隔注意: 不能包含特殊字符（如：#、%、&、+）、敏感词汇, 如果必须使用特殊字符,则需要自行做URL Encoding
					RateAmount:               _RateAmountA,           //商户手续费 可空 本次交易需要扣除的手续费。单位:分,不支持小数点 如不填，则手续费默认0元；
				}
				//开始利用35接口执行代付
				_kuaifutong := Kuaifutong.KuaiPayHelper{}
				_return_35 := Kuaifutong.Return_35{}
				/*小额临时注释 if _AmountY <= 100000 && len(_model.TreatyidSmall) > 5 { //走小额代付 */
				if _AmountY <= 10 && len(_model.TreatyidSmall) > 5 { //走小额代付
					_model_35.CustProtocolNo = _model.TreatyidSmall
					_return_35 = _kuaifutong.Gbp_same_id_credit_card_pay_small(_model_35)
				} else { //走大额代付
					_return_35 = _kuaifutong.Gbp_same_id_credit_card_pay(_model_35)
				}
				if _return_35.Status == "1" {
					_str_rt_35 := fmt.Sprintf("%+v", _return_35)
					//开始更新数据库信用卡订单表[lkt_user_card_order]，补充代付部分订单信息
					_RateAmountAA, _ := strconv.Atoi(_RateAmountA)
					_update_model_order := shop.UserCardOrder{
						UserId:      _UserId,            //用户id
						Ordernob:    _return_35.OrderNo, //订单编号
						Rateamountb: _RateAmountAA,      //代扣手续费
						ReturnB:     _str_rt_35,         //代扣返回信息
						//	AddTime:     time.Now(),         //订单完成时间
					}
					_count, err := DbHelper.MySqlDb().Where("OrderNoA=?", _OrderNoA).Update(_update_model_order)
					ErrorHelper.CheckErr(err)
					//fmt.Println(_count)
					if _count > 0 && err == nil {
						//开始将代扣执行结果以json方式返回
						//_json_model = map[string]interface{}{"code": _rerurn_38.Status, "msg": "success", "data": _rerurn_38, "info": "代扣交易提交成功！"}
						//fmt.Println(_return_35)
						switch _return_35.Status {
						case "0":
							_json_model = map[string]interface{}{"code": 0, "msg": "in_process", "data": _return_35, "info": "代付交易正在处理！"}
						case "1":
							//执行成功开始分润！
							FunRun(_UserId, strconv.Itoa(_model_order.Id), strconv.Itoa(_model_order.Amount/100), _model_order.Bankcardno+"-智收")
							_json_model = map[string]interface{}{"code": 1, "msg": "success", "data": _return_35, "info": "代付交易成功！"}
						case "2":
							_json_model = map[string]interface{}{"code": 2, "msg": "success", "data": _return_35, "info": "代付交易失败！"}
						}
					} else {
						_json_model = map[string]interface{}{"code": 0, "msg": "fail", "info": "代付订单写入数据库失败！"}
					}
				} else { //如果代付执行不成功，10秒后利用协程再执行
					ch := make(chan int, 1)
					go func() {
						ErrorHelper.LogInfo("妥协到10秒后提现")
						time.Sleep(time.Second * 20)
						//开始利用协程执行代扣
						//_Zh_Amount := _rerurn_11.Details[0].BalanceAmount   //获取当前账户在快付通全部余额
						_Amount := _AmountY - _Kft_sxf                  //本单应总额度减去快付通扣除手续费=提现额度含臻方便代付手续费
						_RateAmountA := strconv.Itoa(_RateAmount + 100) //臻方便费率 //用臻方便自己费率算代付手续费（从数据库获取臻方便固定费率+2元即200分）
						_Last_Amount, err := strconv.Atoi(_rerurn_11.Details[0].BalanceAmount)
						ErrorHelper.CheckErr(err)
						if _Last_Amount < _AmountY-_Kft_sxf {
							_Amount = _Last_Amount
						}
						str_order := "DH" + date.FormatDate(time.Now(), "yyyyMMddHHmmss") + StringHelper.GetRandomNum(6)
						_model_35 := Kuaifutong.Model_35{
							OrderNo:   str_order, //订单编号 用于标识商户发起的一笔交易,在批量交易中,此编号可写在批量请求文件中,用于标识批量请求中的每一笔交易
							TradeName: "代付",      //交易名称 由商户填写,简要概括此次交易的内容.用于在查询交易记录时,提醒用户此次交易具体做了什么事情
							//MerchantBankAccountNo:    "",                                            //商户银行账号 可空 商户用于付款的银行账户，资金到账T+0模式时必填。
							//MerchantBindPhoneNo:      "",                                            //商户开户时绑定的手机号（可空）对于有些银行账户被扣款时，需要提供此绑定手机号才能进行交易；此手机号和短信通知客户的手机号可以一致，也可以不一致
							Amount:                   strconv.Itoa(_Amount),  //交易金额 此次交易的具体金额,单位:分,不支持小数点
							CustBankNo:               _model.BankCode,        //客户银行账户行别 客户银行账户所属的银行的编号,具体见第5.3.1章节
							CustBankAccountIssuerNo:  "",                     //客户开户行网点号 可空 指支付系统里的行号，具体到某个支行（网点）号；
							CustBankAccountNo:        _MerchantBankAccountNo, //客户银行账户号 本次交易中,往客户的哪张卡上付钱
							CustName:                 _model.Cardholder,      //客户姓名 收钱的客户的真实姓名
							CustBankAcctType:         "",                     //客户银行账户类型 可空 指客户的银行账户是个人账户还是企业账户
							CustAccountCreditOrDebit: "",                     //客户账户借记贷记类型 可空 若是信用卡，则以下两个参数“信用卡有效期”和“信用卡cvv2”； 1借记 2贷记 4 未知
							CustCardValidDate:        "",                     //客户信用卡有效期 可空 信用卡的正下方的四位数，前两位是月份，后两位是年份；
							CustCardCvv2:             "",                     //客户信用卡的cvv2 可空 信用卡的背面的三位数
							CustID:                   _model.IdCard,          //客户证件号码
							CustPhone:                _model.Mobile,          //客户手机号 如果商户购买的产品中勾选了短信通知功能，则当商户填写了手机号时,快付通会在交易成功后通过短信通知客户
							Messages:                 "",                     //发送客户短信内容 可空 如果填写了,则把此参数值的内容发送给客户；如果没填写，则按照快付通的默认模板发送给客户；
							CustEmail:                "",                     //客户邮箱地址 可空 如果商户购买的产品中勾选了邮件通知功能，则当商户填写了邮箱地址时,快付通会在交易成功后通过邮件通知客户
							EmailMessages:            "",                     //发送客户邮件内容 可空 如果填写了,则把此参数值的内容发送给客户；如果没填写，则按照快付通的默认模板发送给客户；
							Remark:                   "",                     //备注 可空 商户可额外填写备注信息,此信息会传给银行,会在银行的账单信息中显示(具体如何显示取决于银行方,快付通不保证银行肯定能显示)
							CustProtocolNo:           _Treatyid,              //客户协议编号 可空 扣款人在快付通备案的协议号。
							ExtendParams:             "",                     //扩展参数 可空 用于商户的特定业务信息传递，只有商户与快付通约定了传递此参数且约定了参数含义，此参数才有效。参数格式：参数名 1^参数值 1|参数名 2^参数值 2|……多条数据用“|”间隔注意: 不能包含特殊字符（如：#、%、&、+）、敏感词汇, 如果必须使用特殊字符,则需要自行做URL Encoding
							RateAmount:               _RateAmountA,           //商户手续费 可空 本次交易需要扣除的手续费。单位:分,不支持小数点 如不填，则手续费默认0元；
						}
						//开始利用35接口执行代付
						_kuaifutong := Kuaifutong.KuaiPayHelper{}
						_return_35 := Kuaifutong.Return_35{}
						/*小额临时注释 if _AmountY <= 100000 && len(_model.TreatyidSmall) > 5 { //走小额代付 */
						if _AmountY <= 10 && len(_model.TreatyidSmall) > 5 { //走小额代付
							_model_35.CustProtocolNo = _model.TreatyidSmall
							_return_35 = _kuaifutong.Gbp_same_id_credit_card_pay_small(_model_35)
						} else { //走大额代付
							_return_35 = _kuaifutong.Gbp_same_id_credit_card_pay(_model_35)
						}
						_str_rt_35 := fmt.Sprintf("%+v", _return_35)
						//开始更新数据库信用卡订单表[lkt_user_card_order]，补充代付部分订单信息
						_RateAmountAA, _ := strconv.Atoi(_RateAmountA)
						_update_model_order := shop.UserCardOrder{
							UserId:      _UserId,            //用户id
							Ordernob:    _return_35.OrderNo, //订单编号
							Rateamountb: _RateAmountAA,      //代扣手续费
							ReturnB:     _str_rt_35,         //代扣返回信息
							//	AddTime:     time.Now(),         //订单完成时间
						}
						_count, err := DbHelper.MySqlDb().Where("OrderNoA=?", _OrderNoA).Update(_update_model_order)
						ErrorHelper.CheckErr(err)
						//fmt.Println(_count)
						if _count > 0 && err == nil {
							//开始将代扣执行结果以json方式返回
							//_json_model = map[string]interface{}{"code": _rerurn_38.Status, "msg": "success", "data": _rerurn_38, "info": "代扣交易提交成功！"}
							//fmt.Println(_return_35)
							switch _return_35.Status {
							case "0":
								_json_model = map[string]interface{}{"code": 0, "msg": "in_process", "data": _return_35, "info": "代付交易正在处理！"}
							case "1":
								//执行成功开始分润！
								FunRun(_UserId, strconv.Itoa(_model_order.Id), strconv.Itoa(_model_order.Amount/100), _model_order.Bankcardno+"-智收")
								_json_model = map[string]interface{}{"code": 1, "msg": "success", "data": _return_35, "info": "代付交易成功！"}
							case "2":
								_json_model = map[string]interface{}{"code": 2, "msg": "success", "data": _return_35, "info": "代付交易失败！"}
							}
						} else {
							_json_model = map[string]interface{}{"code": 0, "msg": "fail", "info": "代付订单写入数据库失败！"}
						}
						ch <- 1
					}()
					//协程完成
				}
			}
		} else {
			_json_model = map[string]interface{}{"code": 0, "msg": "fail", "info": "代扣订单不存在！"}
		}
	} else {
		_json_model = map[string]interface{}{"code": 0, "msg": "fail", "info": "身份证或入金卡信息有误！"}
	}
	this.Data["json"] = _json_model
	this.ServeJSON()
}

/*
*根据代付单号获取整单流水信息
 */
func (this *OnlinePay_KFT_Controller) Card_Order_GetOne() {
	var _json_model interface{}
	_OrderNoB := this.GetString("OrderNoB") //代付订单id
	fmt.Println(_OrderNoB)
	_model := &shop.UserCardOrder{Ordernob: _OrderNoB}
	results, err := DbHelper.MySqlDb().Get(_model)
	ErrorHelper.CheckErr(err)
	if results {
		_from_bank_name := Get_Bank_Name(_model.Bankcardno)
		_to_bank_name := Get_Bank_Name(_model.Merchantbankaccountno)
		_json_model = map[string]interface{}{"code": 1, "msg": "success", "data": _model, "from_bank_name": _from_bank_name, "to_bank_name": _to_bank_name, "info": "代付执行完毕，现返回订单信息！"}
	}
	this.Data["json"] = _json_model
	this.ServeJSON()
}

/*
*根据用户设置金额自动设置成智能自动代换任务列表(在智能还款里，出入金是同一张信用卡)
 */
func (this *OnlinePay_KFT_Controller) Get_Create_ZnDhList() {
	var _json_model interface{}
	_UserId := this.GetString("UserId")         //用户id
	_BankCardNo := this.GetString("BankCardNo") //用户出金卡号
	_BeginTime := this.GetString("BeginTime")   //开始时间
	_EndTime := this.GetString("EndTime")       //结束时间
	_Amount, err := this.GetFloat("Amount")     //代扣金额 手动输入 从提交获取
	_Number, err := this.GetInt("Number")       //每天执行笔数
	_DiffTime := this.GetString("DiffTime")     //执行时间天列表
	fmt.Println(_DiffTime)
	//转换开始日期
	_BeginTimeA, err := date.ParseLocal(_BeginTime)
	ErrorHelper.CheckErr(err)
	//转换结束日期
	_EndTimeA, err := date.ParseAny(_EndTime)
	ErrorHelper.CheckErr(err)
	fmt.Println(_EndTimeA)
	_nian := strconv.Itoa(_BeginTimeA.Year())      //获取年
	_yue := strconv.Itoa(int(_BeginTimeA.Month())) //获取月
	//_start_ri := _BeginTimeA.Day()                 //获取开始日期
	//_end_ri := _EndTimeA.Day()                     //获取结束日期
	//开始各项计算费用
	//自动分配天列表
	_listdate, _list_ri, _zongbishu := onlinepay.PackageDay(_Number, _DiffTime)
	//	_zongbishu := (_end_ri - _start_ri + 1) * _Number //还款总笔数
	//根据用户id和银行卡号取用户和卡的相关信息
	_model := &shop.UserBankCard{UserId: _UserId, BankCardNumber: _BankCardNo} //这里是入金卡 臻方便费率应该从出金卡扣除
	results, err := DbHelper.MySqlDb().Get(_model)
	ErrorHelper.CheckErr(err)
	if results {
		//快付通代扣费率
		_feilv_kft, err := strconv.ParseFloat(Get_Bank_T0_FL(_model.BankCode), 64)
		ErrorHelper.CheckErr(err)
		_RateAmount_kft := _Amount * _feilv_kft //快付通费率 快付通手续费 总额*费率(这里都是实际数据未换算成分)
		//臻方便代付费率
		_feilv_zfb, err := strconv.ParseFloat(Get_ZFB_T0_FL(_model.BankCode), 64) //出金卡费率
		ErrorHelper.CheckErr(err)
		_RateAmount_zfb := _Amount * _feilv_zfb //臻方便手续费 总额*费率(这里都是实际数据未换算成分)

		_ZongFeiYong := _RateAmount_kft + _RateAmount_zfb //总费用=快付通手续费+臻方便手续费(这里都是实际数据未换算成分)
		_ZongEdu := 0.00
		_Yuliujin := 0.00   //预留金
		_Yuliujin_A := 0.00 //预留金
		//费用计算结束
		//开始生成父订单并存入数据库，在这里直接算好代扣手续费和代付手续费
		_new_model_order := &shop.UserCardOrder{
			UserId:                _UserId,                     //用户id
			Amount:                int(_Amount * float64(100)), //订单总金额
			Bankcardno:            _BankCardNo,                 //出金卡号
			Merchantbankaccountno: _BankCardNo,                 //入金卡号
			AddTime:               time.Now(),                  //订单创建时间
			Isjob:                 1,
		}
		_count, err := DbHelper.MySqlDb().Insert(_new_model_order)
		ErrorHelper.CheckErr(err)
		_parent_order_id := 0
		if _count > 0 && err == nil {
			_parent_order_id = _new_model_order.Id
		}
		//存储创建成功的任务订单进数据库，并返回
		_list_all := make([]*shop.UserCardJobOrder, 0)
		if _parent_order_id > 0 {
			//根据单数计算提现费用并加入总费用
			_ZongFeiYong = (_ZongFeiYong + float64(_zongbishu*2)) * 100 //秒到每次+2 总费用转换完之后变分
			//还款设置总额度=设置还款总数应还款总额+费用
			_ZongEdu = _Amount*float64(100) + _ZongFeiYong + float64(1000) //怕不够还，所以多加10元，方便银行结算

			//获取订单列表中最大一笔金额
			//	_on_max_Amount, _ := arry.Max_IntArry_One(_list_edu) //获取最大一笔额度 这种算法暂时不用
			//先按平均数算预留金
			_on_max_Amount := float64(_Amount * float64(100) / float64(_zongbishu))
			_canshu := 5000
			//开始计算预留金
			if _Amount*float64(100) <= 10000*100 {
				_Yuliujin = float64(_on_max_Amount) + _ZongFeiYong + 10000 //预留金=分批还款中单批最大金额+总手续费+100
				_Yuliujin_A = _ZongFeiYong + 1000                          //预留金=分批还款中单批最大金额+总手续费+100
				_canshu = 8000
			} else if _Amount*float64(100) > 10000*100 && _Amount*float64(100) <= 20000*100 {
				_Yuliujin = float64(_on_max_Amount) + _ZongFeiYong + 20000 //预留金=分批还款中单批最大金额+总手续费+200
				_Yuliujin_A = _ZongFeiYong + 2000                          //预留金=分批还款中单批最大金额+总手续费+200
				_canshu = 15000
			} else if _Amount*float64(100) > 20000*100 && _Amount*float64(100) <= 30000*100 {
				_Yuliujin = float64(_on_max_Amount) + _ZongFeiYong + 30000 //预留金=分批还款中单批最大金额+总手续费+300
				_Yuliujin_A = _ZongFeiYong + 3000                          //预留金=分批还款中单批最大金额+总手续费+300
				_canshu = 20000
			} else if _Amount*float64(100) > 30000*100 && _Amount*float64(100) <= 40000*100 {
				_Yuliujin = float64(_on_max_Amount) + _ZongFeiYong + 40000 //预留金=分批还款中单批最大金额+总手续费+400
				_Yuliujin_A = _ZongFeiYong + 4000                          //预留金=分批还款中单批最大金额+总手续费+400
				_canshu = 30000
			} else {
				_Yuliujin = float64(_on_max_Amount) + _ZongFeiYong + 50000 //预留金=分批还款中单批最大金额+总手续费+500
				_Yuliujin_A = _ZongFeiYong + 5000                          //预留金=分批还款中单批最大金额+总手续费+500
				_canshu = 35000
			}
			//自动分配额度列表
			_list_edu := onlinepay.DFPackage(_zongbishu, int(_ZongEdu/100), _canshu/100, int((_Yuliujin-_ZongFeiYong)/100)) //总额度不平均分配与总笔数
			//获取最大一笔额度
			_on_max_Amount_A, _ := arry.Max_IntArry_One(_list_edu)    //获取最大一笔额度
			_Yuliujin_A = _Yuliujin_A + float64(_on_max_Amount_A*100) //回算预留金
			//自动分配小时列表
			_list_xiaoshi := onlinepay.PackageHour(_zongbishu, _Number, 8, 20)
			//自动分配分钟列表
			_list_fen := onlinepay.PackageTime(_zongbishu, 1, 50)
			//自动分配秒列表
			_list_miao := onlinepay.PackageTime(_zongbishu, 2, 50)
			//定时任务计划列表拼单开始
			_list_order := []map[string]string{}
			//遍历各个设置 生成自动订单
			for i, _ := range _list_edu {
				_order := map[string]string{}
				_order["amount"] = strconv.Itoa(_list_edu[i] * 100)
				_order["nian"] = _nian
				_order["yue"] = StringHelper.Str_Left(_yue, "0", 2)
				_order["ri"] = StringHelper.Str_Left(strconv.Itoa(_list_ri[i]), "0", 2)
				_order["xiaoshi"] = StringHelper.Str_Left(strconv.Itoa(_list_xiaoshi[i]), "0", 2)
				_order["fen"] = StringHelper.Str_Left(strconv.Itoa(_list_fen[i]), "0", 2)
				_order["miao"] = StringHelper.Str_Left(strconv.Itoa(_list_miao[i]), "0", 2)
				_list_order = append(_list_order, _order)
			}
			for i, _model := range _list_order { //遍历历史订单，生成正式订单！
				//单笔执行时间
				_ImplementTime := _listdate[i] + " " + _model["xiaoshi"] + ":" + _model["fen"] + ":" + _model["miao"]
				//开始生成子任务订单并存入数据库
				Amount_A, _ := strconv.Atoi(_model["amount"]) //单笔总金额
				//快付通手续费
				_kft_sxf := math.Floor(float64(Amount_A)*_feilv_kft + 0.5)
				//臻方便手续费
				_zfb_sxf := math.Floor(float64(Amount_A)*_feilv_zfb + 100 + 0.5)
				//代扣协议id
				_Treatyid := ""
				/*小额临时注释 if Amount_A <= 100000 && len(Get_Bank_Model(_BankCardNo).TreatyidSmall) > 5 { */
				if Amount_A <= 10 && len(Get_Bank_Model(_BankCardNo).TreatyidSmall) > 5 {
					_Treatyid = Get_Bank_Model(_BankCardNo).TreatyidSmall
					//小额重新计算快付通手续费
					_kft_sxf = math.Floor(float64(Amount_A)*0.0032 + 0.5)
					//小额重新计算臻方便手续费
					_zfb_sxf = math.Floor(float64(Amount_A)*0.0025 + 100 + 0.5)
				} else { //大额协议号
					_Treatyid = Get_Bank_Model(_BankCardNo).Treatyid
				}
				_new_model_job_order := &shop.UserCardJobOrder{
					ParentOrderId:         _parent_order_id,                                                             //父订单
					UserId:                _UserId,                                                                      //用户id
					Amount:                Amount_A,                                                                     //单笔任务订单金额
					Rateamounta:           int(_kft_sxf),                                                                //单笔快付通代扣费 //快付通代扣手续费 //要四舍五入
					Rateamountb:           int(_zfb_sxf),                                                                //单笔臻方便代付费 //臻方便代还手续费 //要四舍五入
					Bankcardno:            _BankCardNo,                                                                  //出金卡号
					Merchantbankaccountno: _BankCardNo,                                                                  //入金卡号
					Treatyid:              _Treatyid,                                                                    //出金卡协议号
					AddTime:               time.Now(),                                                                   //单笔订单任务计划订单创建时间
					TaskName:              _BankCardNo + "-智能代还",                                                        //单笔订单任务计划任务名称
					CronSpec:              "0 " + _model["fen"] + " " + _model["xiaoshi"] + " " + _model["ri"] + " * ?", //单笔订单任务计划任务时间表达式 例如：每月15日上午10:15触发
					ImplementTime:         _ImplementTime,                                                               //单笔订单任务计划执行时间
					Status:                1,                                                                            //单笔订单任务计划任务状态：0停用 1启用
					IsFinish:              0,                                                                            //订单完成状态
				}
				//_count, err := DbHelper.MySqlDb().Insert(_new_model_job_order)
				//ErrorHelper.CheckErr(err)
				if _count > 0 {
					_list_all = append(_list_all, _new_model_job_order)
				}
			}
		}
		if len(_list_all) > 0 {
			_json_model = map[string]interface{}{"code": 1, "msg": "success", "order_parent_id": _parent_order_id, "order": map[string]interface{}{"all_amount": float32(_Amount), "all_sum": _zongbishu, "yuliujin": float32(_Yuliujin_A / 100), "zongfeiyong": float32(_ZongFeiYong / 100)}, "data": _list_all, "info": "计划订单创建成功！"}
		} else {
			_json_model = map[string]interface{}{"code": 0, "msg": "fail", "info": "计划订单创建失败！"}
		}
	} else {
		_json_model = map[string]interface{}{"code": 0, "msg": "fail", "info": "用户id和所操作卡不匹配！"}
	}
	this.Data["json"] = _json_model
	this.ServeJSON()
}

/*
*订单确认
 */
func (this *OnlinePay_KFT_Controller) Get_Create_ZnDhList_Confirm() {
	var _json_model interface{}
	_Ok := this.GetString("Ok") //是否确认
	_UserId := this.GetString("UserId")
	_BankCardNo := this.GetString("BankCardNo") //银行卡卡号
	_List_Order := this.GetString("List_Order") //获取要确认的订单
	fmt.Println(_List_Order)

	var _list_job_order []shop.UserCardJobOrder
	errA := json.Unmarshal([]byte(_List_Order), &_list_job_order)
	ErrorHelper.CheckErr(errA)
	_list_job_order_A := make([]shop.UserCardJobOrder, 0)
	for _, item := range _list_job_order {
		if len(item.Bankcardno) > 5 {
			_list_job_order_A = append(_list_job_order_A, item)
		}
	}
	//xxx := append(_list_job_order_A, shop.UserCardJobOrder{})
	//fmt.Println(xxx)
	if _Ok == "yes" {
		_model_job_order := shop.UserCardJobOrder{UserId: _UserId, Bankcardno: _BankCardNo}
		_ok, err := DbHelper.MySqlDb().Where(" is_finish=0 ").Get(&_model_job_order)
		ErrorHelper.CheckErr(err)
		if _ok {
			_json_model = map[string]interface{}{"code": 0, "msg": "fail", "err_info": err, "info": "该卡已有预约，请勿重复创建！"}
		} else {
			_count, err := DbHelper.MySqlDb().Insert(&_list_job_order_A)
			ErrorHelper.CheckErr(err)
			if _count > 0 {
				_json_model = map[string]interface{}{"code": 1, "msg": "success", "info": "计划订单确认创建完成！"}
			} else {
				_json_model = map[string]interface{}{"code": 0, "msg": "fail", "err_info": err, "info": "计划订单确认创建失败！"}
			}
		}
	} else {
		_json_model = map[string]interface{}{"code": 0, "msg": "fail", "info": "计划订单确认创建失败！"}
	}

	this.Data["json"] = _json_model
	this.ServeJSON()
}

/*
*信用卡协议签约
 */
func (this *OnlinePay_KFT_Controller) Card_Treaty() {
	var _json_model interface{}
	_UserId := this.GetString("UserId")                       //用户id
	_TreatyType := this.GetString("TreatyType")               //协议类型：11：借记卡扣款 12：信用卡扣款
	_Note := ""                                               // this.GetString("Note") //说明 参数可空
	_EndDate := this.GetString("EndDate")                     //协议失效日期
	_HolderName := this.GetString("HolderName")               //持卡人真实姓名
	_BankType := this.GetString("BankType")                   //银行行别代码
	_BankCardType := this.GetString("BankCardType")           //银行卡类型 1、借记卡 2、信用卡
	_BankCardNo := this.GetString("BankCardNo")               //银行卡号
	_MobileNo := this.GetString("MobileNo")                   //预留手机号码
	_CertificateNo := this.GetString("CertificateNo")         //证件号
	_CustCardValidDate := this.GetString("CustCardValidDate") //客户信用卡有效期
	_CustCardCvv2 := this.GetString("CustCardCvv2")           //客户信用卡的cvv2
	_model_36 := Kuaifutong.Model_36{
		TreatyType:        _TreatyType,
		Note:              _Note,                                   //说明
		StartDate:         date.FormatDate(time.Now(), "yyyyMMdd"), //协议生效日 //根据当前系统自动生成！
		EndDate:           _EndDate,
		HolderName:        _HolderName,
		BankType:          _BankType,
		BankCardType:      _BankCardType,
		BankCardNo:        _BankCardNo,
		MobileNo:          _MobileNo,
		CertificateType:   "0", //证件类型：0表示身份证
		CertificateNo:     _CertificateNo,
		CustCardValidDate: _CustCardValidDate,
		CustCardCvv2:      _CustCardCvv2,
	}
	//开始快捷协议代扣协议申请
	_kuaifutong := Kuaifutong.KuaiPayHelper{}
	_rerurn_36 := _kuaifutong.Gbp_same_id_treaty_collect_apply(_model_36)
	//fmt.Println(_rerurn_36)
	if _rerurn_36.SmsSeq != "" && _rerurn_36.OrderNo != "" { //如果短信流水号不为空，同时订单编号也不为空说明代扣协议申请成功，接着开始确认协议
		_json_model = map[string]interface{}{"code": 1, "msg": "success", "data": _rerurn_36, "info": "代扣协议申请成功！",
			"next_data": map[string]string{
				"UserId":        _UserId,
				"Note":          _Note,          //协议说明
				"Enddate":       _EndDate,       //协议失效日期
				"Bankcardtype":  _BankCardType,  //银行卡类型 1、借记卡 2、信用卡
				"Mobileno":      _MobileNo,      //预留手机号码
				"Certificateno": _CertificateNo, //身份证号
				/*----捎带传参完毕，以下是确认必传参数-----*/
				"SmsSeq":            _rerurn_36.SmsSeq,
				"AuthCode":          "-1",
				"HolderName":        _HolderName,
				"BankCardNo":        _BankCardNo,
				"CustCardValidDate": _CustCardValidDate,
				"CustCardCvv2":      _CustCardCvv2,
				"TreatyType":        _TreatyType,
				"OrderNo":           _rerurn_36.OrderNo,
			},
		}

	} else {
		_json_model = map[string]interface{}{"code": 0, "msg": "fail", "data": _rerurn_36, "info": "代扣协议申请失败！"}
	}
	this.Data["json"] = _json_model
	this.ServeJSON()
}

/*
*信用卡协议签约确认
 */
func (this *OnlinePay_KFT_Controller) Card_Treaty_Confirm() {
	var _json_model interface{}
	_UserId := this.GetString("UserId")               //用户id
	_EndDate := this.GetString("EndDate")             //协议失效日期
	_BankCardType := this.GetString("BankCardType")   //银行卡类型 1、借记卡 2、信用卡
	_MobileNo := this.GetString("MobileNo")           //预留手机号码
	_CertificateNo := this.GetString("CertificateNo") //证件号
	/*----捎带传参完毕，以下是确认必传参数-----*/
	_OrderNo := this.GetString("OrderNo")                     //订单编号
	_SmsSeq := this.GetString("SmsSeq")                       //短信流水号
	_AuthCode := this.GetString("AuthCode")                   //手机动态校验码
	_HolderName := this.GetString("HolderName")               //持卡人真实姓名
	_BankCardNo := this.GetString("BankCardNo")               //银行卡号
	_CustCardValidDate := this.GetString("CustCardValidDate") //客户信用卡有效期
	_CustCardCvv2 := this.GetString("CustCardCvv2")           //客户信用卡的cvv2
	_TreatyType := this.GetString("TreatyType")               //协议类型：11：借记卡扣款 12：信用卡扣款
	_model_37 := Kuaifutong.Model_37{
		OrderNo:           _OrderNo,
		SmsSeq:            _SmsSeq,
		AuthCode:          _AuthCode,
		HolderName:        _HolderName,
		BankCardNo:        _BankCardNo,
		CustCardValidDate: _CustCardValidDate,
		CustCardCvv2:      _CustCardCvv2,
	}
	_kuaifutong := Kuaifutong.KuaiPayHelper{}
	_rerurn_37 := _kuaifutong.Gbp_same_id_confirm_treaty_collect_apply(_model_37, _TreatyType)
	//fmt.Println(_rerurn_37)
	if _rerurn_37.OrderNo != "" && _rerurn_37.TreatyId != "" { //如果协议号不为空，同时订单编号也不为空说明代扣协议确认成功,将代扣协议信息存入数据库表[lkt_user_card_treaty]
		_CardType, err := strconv.Atoi(_BankCardType)
		ErrorHelper.CheckErr(err)
		_model := &shop.UserBankCard{UserId: _UserId, BankCardNumber: _BankCardNo}
		results, _ := DbHelper.MySqlDb().Get(_model)
		if results {
			_update_model := shop.UserBankCard{
				StoreId:        8,
				UserId:         _UserId,                        //用户id
				Treatyid:       _rerurn_37.TreatyId,            //协议号
				Treatytype:     _TreatyType,                    //协议类型：11：借记卡扣款 12：信用卡扣款 13：余额扣款 14：余额+借记卡扣款 15： 余额+信用卡扣款
				Treatyenddate:  _EndDate,                       //协议失效日期
				Cardholder:     _HolderName,                    //持卡人真实姓名
				BankCode:       Get_Bank_Code(_model.BankName), //银行行别代码
				CardType:       _CardType,                      //银行卡类型 1、借记卡 2、信用卡
				BankCardNumber: _BankCardNo,                    //银行卡号
				Mobile:         _MobileNo,                      //预留手机号
				IdCard:         _CertificateNo,                 //身份证号
				Expiretime:     _CustCardValidDate,             //客户信用卡有效期
				Cvv2:           _CustCardCvv2,                  //cvv2
				Delflag:        1,                              //是否签约标志
			}
			_, err := DbHelper.MySqlDb().Id(_model.Id).Update(_update_model)
			ErrorHelper.CheckErr(err)
		} else {
			_new_model := shop.UserBankCard{
				StoreId:       8,
				UserId:        _UserId,             //用户id
				Treatyid:      _rerurn_37.TreatyId, //协议号
				Treatytype:    _TreatyType,         //协议类型：11：借记卡扣款 12：信用卡扣款 13：余额扣款 14：余额+借记卡扣款 15： 余额+信用卡扣款
				Treatyenddate: _EndDate,            //协议失效日期
				Cardholder:    _HolderName,         //持卡人真实姓名
				//	BankCode:       Get_Bank_Code(_model.BankName), //银行行别代码
				CardType:       _CardType,          //银行卡类型 1、借记卡 2、信用卡
				BankCardNumber: _BankCardNo,        //银行卡号
				Mobile:         _MobileNo,          //预留手机号
				IdCard:         _CertificateNo,     //身份证号
				Expiretime:     _CustCardValidDate, //客户信用卡有效期
				Cvv2:           _CustCardCvv2,      //cvv2
				Delflag:        1,                  //是否签约标志
			}
			_, err := DbHelper.MySqlDb().Insert(_new_model)
			ErrorHelper.CheckErr(err)

		}
		_json_model = map[string]interface{}{"code": 1, "msg": "success", "data": _rerurn_37, "info": "代扣协议申请确认成功！"}
	} else {
		_json_model = map[string]interface{}{"code": 1, "msg": "success", "data": _rerurn_37, "info": "代扣协议申请确认失败！"}
	}
	this.Data["json"] = _json_model
	this.ServeJSON()
}

/*
*信用卡协议签约(小额)
 */
func (this *OnlinePay_KFT_Controller) Card_Treaty_Small() {
	var _json_model interface{}
	_UserId := this.GetString("UserId")                       //用户id
	_TreatyType := this.GetString("TreatyType")               //协议类型：11：借记卡扣款 12：信用卡扣款
	_Note := ""                                               // this.GetString("Note") //说明 参数可空
	_EndDate := this.GetString("EndDate")                     //协议失效日期
	_HolderName := this.GetString("HolderName")               //持卡人真实姓名
	_BankType := this.GetString("BankType")                   //银行行别代码
	_BankCardType := this.GetString("BankCardType")           //银行卡类型 1、借记卡 2、信用卡
	_BankCardNo := this.GetString("BankCardNo")               //银行卡号
	_MobileNo := this.GetString("MobileNo")                   //预留手机号码
	_CertificateNo := this.GetString("CertificateNo")         //证件号
	_CustCardValidDate := this.GetString("CustCardValidDate") //客户信用卡有效期
	_CustCardCvv2 := this.GetString("CustCardCvv2")           //客户信用卡的cvv2
	_model_36 := Kuaifutong.Model_36{
		TreatyType:        _TreatyType,
		Note:              _Note,                                   //说明
		StartDate:         date.FormatDate(time.Now(), "yyyyMMdd"), //协议生效日 //根据当前系统自动生成！
		EndDate:           _EndDate,
		HolderName:        _HolderName,
		BankType:          _BankType,
		BankCardType:      _BankCardType,
		BankCardNo:        _BankCardNo,
		MobileNo:          _MobileNo,
		CertificateType:   "0", //证件类型：0表示身份证
		CertificateNo:     _CertificateNo,
		CustCardValidDate: _CustCardValidDate,
		CustCardCvv2:      _CustCardCvv2,
	}
	//开始快捷协议代扣协议申请
	_kuaifutong := Kuaifutong.KuaiPayHelper{}
	_rerurn_36 := _kuaifutong.Gbp_same_id_treaty_collect_apply_small(_model_36)
	//fmt.Println(_rerurn_36)
	if _rerurn_36.SmsSeq != "" && _rerurn_36.OrderNo != "" { //如果短信流水号不为空，同时订单编号也不为空说明代扣协议申请成功，接着开始确认协议
		_json_model = map[string]interface{}{"code": 1, "msg": "success", "data": _rerurn_36, "info": "代扣协议申请成功！",
			"next_data": map[string]string{
				"UserId":        _UserId,
				"Note":          _Note,          //协议说明
				"Enddate":       _EndDate,       //协议失效日期
				"Bankcardtype":  _BankCardType,  //银行卡类型 1、借记卡 2、信用卡
				"Mobileno":      _MobileNo,      //预留手机号码
				"Certificateno": _CertificateNo, //身份证号
				/*----捎带传参完毕，以下是确认必传参数-----*/
				"SmsSeq":            _rerurn_36.SmsSeq,
				"AuthCode":          "-1",
				"HolderName":        _HolderName,
				"BankCardNo":        _BankCardNo,
				"CustCardValidDate": _CustCardValidDate,
				"CustCardCvv2":      _CustCardCvv2,
				"TreatyType":        _TreatyType,
				"OrderNo":           _rerurn_36.OrderNo,
			},
		}

	} else {
		_json_model = map[string]interface{}{"code": 0, "msg": "fail", "data": _rerurn_36, "info": "代扣协议申请失败！"}
	}
	this.Data["json"] = _json_model
	this.ServeJSON()
}

/*
*信用卡协议签约确认(小额)
 */
func (this *OnlinePay_KFT_Controller) Card_Treaty_Confirm_Small() {
	var _json_model interface{}
	_UserId := this.GetString("UserId")               //用户id
	_EndDate := this.GetString("EndDate")             //协议失效日期
	_BankCardType := this.GetString("BankCardType")   //银行卡类型 1、借记卡 2、信用卡
	_MobileNo := this.GetString("MobileNo")           //预留手机号码
	_CertificateNo := this.GetString("CertificateNo") //证件号
	/*----捎带传参完毕，以下是确认必传参数-----*/
	_OrderNo := this.GetString("OrderNo")                     //订单编号
	_SmsSeq := this.GetString("SmsSeq")                       //短信流水号
	_AuthCode := this.GetString("AuthCode")                   //手机动态校验码
	_HolderName := this.GetString("HolderName")               //持卡人真实姓名
	_BankCardNo := this.GetString("BankCardNo")               //银行卡号
	_CustCardValidDate := this.GetString("CustCardValidDate") //客户信用卡有效期
	_CustCardCvv2 := this.GetString("CustCardCvv2")           //客户信用卡的cvv2
	_TreatyType := this.GetString("TreatyType")               //协议类型：11：借记卡扣款 12：信用卡扣款
	_model_37 := Kuaifutong.Model_37{
		OrderNo:           _OrderNo,
		SmsSeq:            _SmsSeq,
		AuthCode:          _AuthCode,
		HolderName:        _HolderName,
		BankCardNo:        _BankCardNo,
		CustCardValidDate: _CustCardValidDate,
		CustCardCvv2:      _CustCardCvv2,
	}
	_kuaifutong := Kuaifutong.KuaiPayHelper{}
	_rerurn_37 := _kuaifutong.Gbp_same_id_confirm_treaty_collect_apply_small(_model_37, _TreatyType)
	//fmt.Println(_rerurn_37)
	if _rerurn_37.OrderNo != "" && _rerurn_37.TreatyId != "" { //如果协议号不为空，同时订单编号也不为空说明代扣协议确认成功,将代扣协议信息存入数据库表[lkt_user_card_treaty]
		_CardType, err := strconv.Atoi(_BankCardType)
		ErrorHelper.CheckErr(err)
		_model := &shop.UserBankCard{UserId: _UserId, BankCardNumber: _BankCardNo}
		results, _ := DbHelper.MySqlDb().Get(_model)
		if results {
			_update_model := shop.UserBankCard{
				StoreId:        8,
				UserId:         _UserId,                        //用户id
				TreatyidSmall:  _rerurn_37.TreatyId,            //协议号
				Treatytype:     _TreatyType,                    //协议类型：11：借记卡扣款 12：信用卡扣款 13：余额扣款 14：余额+借记卡扣款 15： 余额+信用卡扣款
				Treatyenddate:  _EndDate,                       //协议失效日期
				Cardholder:     _HolderName,                    //持卡人真实姓名
				BankCode:       Get_Bank_Code(_model.BankName), //银行行别代码
				CardType:       _CardType,                      //银行卡类型 1、借记卡 2、信用卡
				BankCardNumber: _BankCardNo,                    //银行卡号
				Mobile:         _MobileNo,                      //预留手机号
				IdCard:         _CertificateNo,                 //身份证号
				Expiretime:     _CustCardValidDate,             //客户信用卡有效期
				Cvv2:           _CustCardCvv2,                  //cvv2
				DelflagSmall:   1,                              //是否签约标志
			}
			_, err := DbHelper.MySqlDb().Id(_model.Id).Update(_update_model)
			ErrorHelper.CheckErr(err)
		} else {
			_new_model := shop.UserBankCard{
				StoreId:       8,
				UserId:        _UserId,             //用户id
				TreatyidSmall: _rerurn_37.TreatyId, //小额协议号
				Treatytype:    _TreatyType,         //协议类型：11：借记卡扣款 12：信用卡扣款 13：余额扣款 14：余额+借记卡扣款 15： 余额+信用卡扣款
				Treatyenddate: _EndDate,            //协议失效日期
				Cardholder:    _HolderName,         //持卡人真实姓名
				//	BankCode:       Get_Bank_Code(_model.BankName), //银行行别代码
				CardType:       _CardType,          //银行卡类型 1、借记卡 2、信用卡
				BankCardNumber: _BankCardNo,        //银行卡号
				Mobile:         _MobileNo,          //预留手机号
				IdCard:         _CertificateNo,     //身份证号
				Expiretime:     _CustCardValidDate, //客户信用卡有效期
				Cvv2:           _CustCardCvv2,      //cvv2
				DelflagSmall:   1,                  //小额是否签约标志
			}
			_, err := DbHelper.MySqlDb().Insert(_new_model)
			ErrorHelper.CheckErr(err)

		}
		_json_model = map[string]interface{}{"code": 1, "msg": "success", "data": _rerurn_37, "info": "代扣协议申请确认成功！"}
	} else {
		_json_model = map[string]interface{}{"code": 1, "msg": "success", "data": _rerurn_37, "info": "代扣协议申请确认失败！"}
	}
	this.Data["json"] = _json_model
	this.ServeJSON()
}

/*
*判断用户当前出金卡是否签订代扣协议
 */
func (this *OnlinePay_KFT_Controller) Card_Treaty_IsOk() {
	var _json_model interface{}
	_UserId := this.GetString("UserId")         //用户id
	_BankCardNo := this.GetString("BankCardNo") //用户出金卡号
	//根据userid和出金卡号取代扣签约信息
	_model := &shop.UserBankCard{UserId: _UserId, BankCardNumber: _BankCardNo}
	results, _ := DbHelper.MySqlDb().Get(_model)
	if results {
		if len(_model.Treatyid) > 2 {
			_json_model = map[string]interface{}{"code": 1, "msg": "success", "data": _model, "info": "已签订代扣协议！"}
		} else {
			_json_model = map[string]interface{}{"code": 2, "msg": "fail", "data": _model, "info": "还未签订代扣协议！"}
		}
	} else {
		_json_model = map[string]interface{}{"code": 0, "msg": "fail", "data": _model}
	}
	this.Data["json"] = _json_model
	this.ServeJSON()
}

/*
*判断用户当前出金卡是否签小额订代扣协议
 */
func (this *OnlinePay_KFT_Controller) Card_Treaty_IsOk_Samll() {
	var _json_model interface{}
	_UserId := this.GetString("UserId")         //用户id
	_BankCardNo := this.GetString("BankCardNo") //用户出金卡号
	//根据userid和出金卡号取代扣签约信息
	_model := &shop.UserBankCard{UserId: _UserId, BankCardNumber: _BankCardNo}
	results, _ := DbHelper.MySqlDb().Get(_model)
	if results {
		if len(_model.TreatyidSmall) > 2 && _model.DelflagSmall > 0 {
			_json_model = map[string]interface{}{"code": 1, "msg": "success", "data": _model, "info": "已签订小额代扣协议！"}
		} else {
			_json_model = map[string]interface{}{"code": 2, "msg": "fail", "data": _model, "info": "还未签订小额代扣协议！"}
		}
	} else {
		_json_model = map[string]interface{}{"code": 0, "msg": "fail", "data": _model}
	}
	this.Data["json"] = _json_model
	this.ServeJSON()
}

/*
*判断用户是否实名认证
 */
func (this *OnlinePay_KFT_Controller) User_Is_Auth() {
	var _json_model interface{}
	_UserId := this.GetString("UserId") //用户id
	//根据userid取实名认证信息
	_model := &shop.UserAuth{UserId: _UserId, Ysbflag: 1}
	results, _ := DbHelper.MySqlDb().Get(_model)
	if !results {
		_json_model = map[string]interface{}{"code": 0, "msg": "fail", "data": _model}
	} else {
		_json_model = map[string]interface{}{"code": 1, "msg": "success", "data": _model}
	}
	this.Data["json"] = _json_model
	this.ServeJSON()
}

/*
*用户实名认证
 */
func (this *OnlinePay_KFT_Controller) User_Auth_Add() {
	var _json_model interface{}
	_UserId := this.GetString("UserId")               //用户id
	_Truename := this.GetString("Truename")           //真实姓名
	_Idcard := this.GetString("Idcard")               //身份证号
	_Personimg := this.GetString("Personimg")         //身份证正面照片
	_Personimgback := this.GetString("Personimgback") //身份证背面照片
	_model := &shop.UserAuth{UserId: _UserId}
	results, _ := DbHelper.MySqlDb().Get(_model)
	if results { //如果用户已实名，无需再操作
		if _model.Ysbflag == 1 {
			_json_model = map[string]interface{}{"code": 0, "msg": "fail", "info": "该用户已实名认证过无需再认证！"}
		} else { //执行修改
			_update_model := shop.UserAuth{
				UserId:        _UserId,
				Truename:      _Truename,
				Idcard:        _Idcard,
				Personimg:     _Personimg,
				Personimgback: _Personimgback,
				Addtime:       time.Now(),
			}
			_, err := DbHelper.MySqlDb().Id(_model.Id).Update(&_update_model)
			ErrorHelper.CheckErr(err)
			_json_model = map[string]interface{}{"code": 1, "msg": "success", "info": "认证已提交，等待审核！"}
		}
	} else { //执行添加
		_new_model := shop.UserAuth{
			UserId:        _UserId,
			Truename:      _Truename,
			Idcard:        _Idcard,
			Personimg:     _Personimg,
			Personimgback: _Personimgback,
			Authflag:      0,
			Ysbflag:       0,
			Addtime:       time.Now(),
		}
		_, err := DbHelper.MySqlDb().Insert(_new_model)
		ErrorHelper.CheckErr(err)
		_json_model = map[string]interface{}{"code": 1, "msg": "success", "info": "认证已提交，等待审核！"}
	}
	this.Data["json"] = _json_model
	this.ServeJSON()
}

/*
*用户实名认证后台审核
 */
func (this *OnlinePay_KFT_Controller) User_Auth_Allow() {
	var _json_model interface{}
	_UserId := this.GetString("UserId") //用户id

	_model := &shop.UserAuth{UserId: _UserId}
	results, _ := DbHelper.MySqlDb().Get(_model)
	if results { //如果用户已实名，无需再操作
		_update_model := shop.UserAuth{
			UserId:   _UserId,
			Authflag: 1,
			Ysbflag:  1,
		}
		_, err := DbHelper.MySqlDb().Id(_model.Id).Update(&_update_model)
		ErrorHelper.CheckErr(err)
		_json_model = map[string]interface{}{"code": 1, "msg": "success", "info": "审核成功！"}
	} else { //执行添加
		_json_model = map[string]interface{}{"code": 0, "msg": "fail", "info": "请选择要审核的用户！"}
	}
	this.Data["json"] = _json_model
	this.ServeJSON()
}

/*------------------------------回传处理开始--------------------------------*/
func (this *OnlinePay_KFT_Controller) Notify_kft() {
	_Language := this.GetString("Language")
	_CallerIp := this.GetString("CallerIp")
	_SignatureAlgorithm := this.GetString("SignatureAlgorithm")
	_SignatureInfo := this.GetString("SignatureInfo")
	_OrderNo := this.GetString("OrderNo")
	_Status := this.GetString("Status")
	_FailureDetails := this.GetString("FailureDetails")
	_ErrorCode := this.GetString("ErrorCode")
	_Amount := this.GetString("Amount")
	_MerchantId := this.GetString("MerchantId")
	_retutn_model := Kuaifutong.Return_38_A{
		//语言
		Language: _Language,
		//调用端IP
		CallerIp: _CallerIp,
		//参数签名算法
		SignatureAlgorithm: _SignatureAlgorithm,
		//签名值
		SignatureInfo: _SignatureInfo,
		//订单编号
		OrderNo: _OrderNo,
		//交易状态
		Status: _Status,
		//失败详情
		FailureDetails: _FailureDetails,
		//错误码
		ErrorCode: _ErrorCode,
		//交易金额
		Amount: _Amount,
		//商户编号
		MerchantId: _MerchantId,
	}
	ErrorHelper.LogInfo("这里开始显示回传内容：")
	ErrorHelper.LogInfo(_retutn_model)
	//接下来要将交易结果与交易订单比对并修改订单相关参数
	this.Data["json"] = _retutn_model
	this.ServeJSON()
}
func (this *OnlinePay_KFT_Controller) quickPayReturn() {
	this.ServeJSON()
}

/*------------------------------回传处理结束--------------------------------*/

/*-------其它公用函数--------*/

/*
*根据银行代码获取当前银行t+0费率
 */
func Get_Bank_T0_FL(_UserId string) string {
	_rt := "0.0058"
	// _model_new := &shop.ChannelBankKft{Bankcode: _bank_code}
	// has, err := DbHelper.MySqlDb().Get(_model_new)
	// ErrorHelper.CheckErr(err)
	// if has {
	// 	if _model_new.D0freerate != "" {
	// 		_rt = _model_new.D0freerate
	// 	}
	// }
	_model_new := &shop.CardAmount{UserId: _UserId}
	has, err := DbHelper.MySqlDb().Get(_model_new)
	ErrorHelper.CheckErr(err)
	if has {
		if _model_new.UserEdfl > 0 {
			_rt = "0.00" + strconv.Itoa(_model_new.UserEdfl)
		}
	}
	return _rt
}

/*
*根据银行代码获取当前臻方便t+0费率
 */
func Get_ZFB_T0_FL(_bank_code string) string {
	//_rt := "0.0010"
	// _model_new := &shop.ChannelBankKft{Bankcode: _bank_code}
	// has, err := DbHelper.MySqlDb().Get(_model_new)
	// ErrorHelper.CheckErr(err)
	// if has {
	// 	if _model_new.D0myrate != "" {
	// 		_rt = _model_new.D0myrate
	// 	}
	// }
	return "0" //_rt
}

/*
*根据银行代码获取当前银行别代码
 */
func Get_Bank_Code(_bank_name string) string {
	_rt := ""
	_model_new := &shop.ChannelBankKft{Bankname: _bank_name}
	has, err := DbHelper.MySqlDb().Get(_model_new)
	ErrorHelper.CheckErr(err)
	if has {
		_rt = _model_new.Bankcode
	} else {
		_rt = ""
	}
	return _rt
}

/*
*根据银行卡号查询银行名称
 */
func Get_Bank_Name(_Bank_No string) string {
	_rt := ""
	_model_new := &shop.UserBankCard{BankCardNumber: _Bank_No}
	has, err := DbHelper.MySqlDb().Get(_model_new)
	ErrorHelper.CheckErr(err)
	if has {
		_rt = _model_new.BankName
	}
	return _rt
}

/*
*根据银行卡号获取银行行别代码
 */
func Get_Bank_CodeA(_Bank_No string) string {
	_rt := ""
	_model_new := &shop.UserBankCard{BankCardNumber: _Bank_No}
	has, err := DbHelper.MySqlDb().Get(_model_new)
	ErrorHelper.CheckErr(err)
	if has {
		_rt = _model_new.BankCode
	}
	return _rt
}

/*
*根据银行卡号获取获取绑卡信息
 */
func Get_Bank_Model(_Bank_No string) *shop.UserBankCard {
	_rt := &shop.UserBankCard{}
	_model_new := &shop.UserBankCard{BankCardNumber: _Bank_No}
	has, err := DbHelper.MySqlDb().Get(_model_new)
	ErrorHelper.CheckErr(err)
	if has {
		_rt = _model_new
	}
	return _rt
}

/*
*分润处理
 */
func FunRun(user_id, order_no, amount, mark string) {
	type _rt struct {
		code int
		msg  string
	}
	// _HeaderData := map[string]string{"Content-Type": "application/x-www-form-urlencoded"}
	// _BodyData := map[string]interface{}{"user_id": user_id, "order_no": order_no, "type": "2", "amount": amount, "mark": mark}
	// _http_url := "https://shop.xhdncppf.com/index.php?module=app&action=index&store_id=8&app=calc_profit"
	// _req := WebHelper.HttpPost(_http_url, _HeaderData, _BodyData)
	// err := json.Unmarshal([]byte(_req), &_rt{})
	// ErrorHelper.CheckErr(err)
}
