package dysms

import (
	"encoding/json"
	"errors"
	"strings"

	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/requests"
	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/responses"
	"github.com/aliyun/alibaba-cloud-sdk-go/services/dysmsapi"
	log "github.com/techidea8/restctl/pkg/log"
)

/*
*
request.Method = "POST"

	request.Scheme = "https" // https | http
	request.Domain = "dysmsapi.aliyuncs.com"
	request.Version = "2017-05-25"
	request.ApiName = "SendSms"
	request.QueryParams["RegionId"] = "cn-beijing"
	request.QueryParams["PhoneNumbers"] = "xxxxxx"                         //手机号
	request.QueryParams["SignName"] = "xxxxx"                               //阿里云验证过的项目名 自己设置
	request.QueryParams["TemplateCode"] = "xxx"       //阿里云的短信模板号 自己设置
	request.QueryParams["TemplateParam"] = "{\"code\":" + "" + "}" //短信模板中的验证码内容 自己生成   之前试过直接返
*/
type DysmsConf struct {
	Endpoint        string
	RegionId        string
	SignName        string
	TemplateCode    string
	AccessKeyId     string
	AccessKeySecret string
}

var defaultDysmsService *DysmsService

type OnResultCallback func(SmsTaskResponse)
type DysmsService struct {
	chSend           chan SmsTask
	conf             DysmsConf
	onResultCallback OnResultCallback
}
type SmsTaskResponse struct {
	Resp   *responses.CommonResponse
	TaskId uint
	Error  error
}
type DysmsServiceOption func(*DysmsService)

func WithResultCallBack(callback OnResultCallback) DysmsServiceOption {
	return func(ds *DysmsService) {
		ds.onResultCallback = callback
	}
}

func SetDefaultService() DysmsServiceOption {
	return func(ds *DysmsService) {
		defaultDysmsService = ds
	}
}

type SmsTask struct {
	TaskId       uint
	PhoneNumbers []string
	Params       map[string]string
}

func (t SmsTask) PhoneNumberString() string {
	return strings.Join(t.PhoneNumbers, ",")
}

func (t SmsTask) ParamsString() string {
	bts, _ := json.Marshal(t.Params)
	return string(bts)
}

// 初始化默认的
func NewDysmsService(conf DysmsConf, options ...DysmsServiceOption) *DysmsService {
	result := &DysmsService{
		chSend: make(chan SmsTask, 100),
		conf:   conf,
	}
	for _, v := range options {
		v(result)
	}
	return result
}
func (svc *DysmsService) Start() {
	go svc.run()
}
func (svc *DysmsService) run() {
	for {
		select {
		case task := <-svc.chSend:
			svc.dispatch(task)
		}
	}
}

func Publish(task SmsTask) error {
	if defaultDysmsService != nil {
		defaultDysmsService.Publish(task)
		return nil
	} else {
		return errors.New("please init DysmsService")
	}
}

func (svc *DysmsService) Publish(task SmsTask) {
	svc.chSend <- task
}
func (svc *DysmsService) dispatch(task SmsTask) {
	if 0 == len(task.PhoneNumbers) {
		log.Error("phone number is empty")
		return
	}
	client, err := dysmsapi.NewClientWithAccessKey(svc.conf.RegionId, svc.conf.AccessKeyId, svc.conf.AccessKeySecret)
	if err != nil {
		log.Error(task.PhoneNumberString(), err.Error())
	}
	request := requests.NewCommonRequest()
	request.Method = "POST"
	request.Scheme = "https" // https | http
	request.Domain = svc.conf.Endpoint
	request.Version = "2017-05-25"
	request.ApiName = "SendSms"
	request.QueryParams["RegionId"] = svc.conf.RegionId
	request.QueryParams["PhoneNumbers"] = task.PhoneNumberString() //手机号
	request.QueryParams["SignName"] = svc.conf.SignName            //阿里云验证过的项目名 自己设置
	request.QueryParams["TemplateCode"] = svc.conf.TemplateCode    //阿里云的短信模板号 自己设置
	request.QueryParams["TemplateParam"] = task.ParamsString()     //短信模板中的验证码内容 自己生成   之前试过直接返回，但是失败，加上code成功。
	response, err := client.ProcessCommonRequest(request)
	if err := client.DoAction(request, response); err != nil {
		log.Error(err.Error())
	} else {
		if response.IsSuccess() {
			log.Info(task.PhoneNumberString(), "SUCCESS")
		} else {
			log.Error(task.PhoneNumberString(), response.BaseResponse.String())
		}
	}
	if svc.onResultCallback != nil {
		go svc.onResultCallback(SmsTaskResponse{
			TaskId: task.TaskId,
			Resp:   response,
			Error:  err,
		})
	}
}
