package notifications

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/brpaz/github-stars-notion-sync/internal/log"
)

const (
	getTokenUrl = "https://qyapi.weixin.qq.com/cgi-bin/gettoken"
	sendMsgUrl  = "https://qyapi.weixin.qq.com/cgi-bin/message/send?access_token="
)

type WechatNotifier struct {
	corpid     string
	corpsecret string
	toUser     string
	agentid    string
}

type GetTokenRsp struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int    `json:"expires_in"`
	ErrCode     int    `json:"errcode"`
	ErrMsg      string `json:"errmsg"`
}

type SendMspRsp struct {
	ErrCode int    `json:"errcode"`
	ErrMsg  string `json:"errmsg"`
}

func NewWechatNotifier(params string) *WechatNotifier {
	if len(params) <= 0 {
		log.Info(nil, "Required argument --notification-wechat-params(cli) or WATCHTOWER_NOTIFICATION_WECHAT_PARAMS(env) is empty.")
		return nil
	}

	allParas := strings.Split(params, ",")
	if len(allParas) < 4 {
		log.Error(nil, "Required argument --notification-wechat(cli) or NOTIFICATION_WECHAT(env) is invalid.")
		return nil
	}

	return &WechatNotifier{
		corpid:     allParas[0],
		corpsecret: allParas[1],
		toUser:     allParas[2],
		agentid:    allParas[3],
	}
}

func (w *WechatNotifier) SendMsg(msg string) {
	if w == nil {
		return
	}

	data := map[string]interface{}{
		"touser":  w.toUser,
		"msgtype": "text",
		"agentid": w.agentid,
		"text": map[string]string{
			"content": msg,
		},
		"safe": "0",
	}

	requestBody, err := json.Marshal(&data)
	if err != nil {
		log.Error(nil, fmt.Sprintf("SendMsg marshal send msg body err: %v", err))
		return
	}

	responseBody, err := httpPost(sendMsgUrl+w.getAccessToken(), requestBody)
	if err != nil {
		log.Error(nil, fmt.Sprintf("SendMsg httpPost err: %v", err))
		return
	}

	var rsp SendMspRsp
	if err = json.Unmarshal(responseBody, &rsp); err != nil {
		log.Error(nil, fmt.Sprintf("SendMsg Unmarshal err: %v", err))
		return
	}

	if strings.Contains(rsp.ErrMsg, "ok") {
		log.Info(nil, "SendMsg success")
		return
	}

	log.Error(nil, fmt.Sprintf("SendMsg err: %v", rsp.ErrMsg))
}

func (w *WechatNotifier) getAccessToken() string {
	data := map[string]interface{}{
		"corpid":     w.corpid,
		"corpsecret": w.corpsecret,
	}

	requestBody, err := json.Marshal(&data)
	if err != nil {
		log.Error(nil, fmt.Sprintf("marshal get access token body err: %v", err))
		return ""
	}

	responseBody, err := httpPost(getTokenUrl, requestBody)
	if err != nil {
		log.Error(nil, fmt.Sprintf("getAccessToken httpPost err: %v", err))
		return ""
	}

	var rsp GetTokenRsp
	if err = json.Unmarshal(responseBody, &rsp); err != nil {
		log.Error(nil, fmt.Sprintf("Unmarshal err: %v", err))
		return ""
	}

	return rsp.AccessToken
}

func httpPost(url string, data []byte) ([]byte, error) {
	client := &http.Client{}
	request, err := http.NewRequest("POST", url, bytes.NewBuffer(data))
	if err != nil {
		return nil, err
	}

	request.Header.Set("Content-Type", "application/json")

	response, err := client.Do(request)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	responseBody := new(bytes.Buffer)
	_, err = responseBody.ReadFrom(response.Body)
	if err != nil {
		return nil, err
	}

	return responseBody.Bytes(), nil
}
