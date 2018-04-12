package baiduvoice

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

const (
	OauthURL = "https://aip.baidubce.com/oauth/2.0/token?grant_type=client_credentials&client_id=%s&client_secret=%s"
	TtsURL   = "http://tsn.baidu.com/text2audio?tex=%s&lan=zh&ctp=1&cuid=%d&tok=%s&per=%d&spd=%d&pit=%d&vol=%d"
	AsrURL   = "http://vop.baidu.com/server_api?dev_pid=1536&token=%s&cuid=%d"
)

type BaiduVoice struct {
	accessToken string //
	Scope       string
	Expires     time.Time
	AppID       string
	APIKey      string
	SecretKey   string
}

//维护accesstoken更新
func (bv *BaiduVoice) getAccessToken() string {
	if bv.Expires.Sub(time.Now()) < 10*time.Second { //需要重新获取AccessToken
		bv.makeAccessToken()
	}
	return bv.accessToken
}

func (bv *BaiduVoice) makeAccessToken() {
	resp, err := http.Get(fmt.Sprintf(OauthURL, bv.APIKey, bv.SecretKey))
	if err != nil {
		log.Println(err.Error())
		return
	}
	bs, _ := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	result := struct {
		AccessToken   string `json:"access_token"`
		SessionKey    string `json:"session_key"`
		Scope         string `json:"scope"`
		SessionSecret string `json:"session_secret"`
		ExpiresIn     int    `json:"expires_in"`
	}{}
	err = json.Unmarshal(bs, &result)
	if err != nil {
		log.Println(err.Error())
		return
	}
	bv.Scope = result.Scope
	bv.accessToken = result.AccessToken
	bv.Expires = time.Now().Add(time.Duration(result.ExpiresIn) * time.Second)
}

/**
 * 参数说明
 *   text 要合成的文本
 *   per 发音人选择, 0为普通女声，1为普通男生，3为情感合成-度逍遥，4为情感合成-度丫丫，默认为普通女声
 *   spd 语速，取值0-9，默认为5中语速
 *   pit 音调，取值0-9，默认为5中语调
 *   vol 音量，取值0-9，默认为5中音量
 *   w 输出流，百度返回的是audio/mp3,网页可以直接播放
 */
func (bv *BaiduVoice) Tts(text string, per, spd, pit, vol int, w io.Writer) {
	req := "?" + text
	u, _ := url.Parse(req)
	encodedtext := u.Query().Encode()
	encodedtext = encodedtext[:len(encodedtext)-1]
	cuid := rand.Intn(100000) + 10000
	tts_url := fmt.Sprintf(TtsURL, encodedtext, cuid, bv.getAccessToken(),
		per, spd, pit, vol)
	resp, err := http.Get(tts_url)
	if err != nil {
		log.Println(err.Error())
		return
	}
	bs, _ := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	w.Write(bs)
}

/**
 * 参数说明
 * bs pcm音频文件
 */
func (bv *BaiduVoice) Asr(bs []byte) string {
	cuid := rand.Intn(100000) + 10000
	asr_url := fmt.Sprintf(AsrURL, bv.getAccessToken(), cuid)
	client := http.DefaultClient
	req, err := http.NewRequest("POST", asr_url, bytes.NewReader(bs))
	if err != nil {
		log.Println(err.Error())
		return ""
	}
	req.Header.Add("Content-Type", "audio/pcm;rate=8000")
	req.Header.Add("Content-Length", strconv.Itoa(len(bs)))
	resp, err := client.Do(req)
	if err != nil {
		log.Println(err.Error())
		return ""
	}
	bs, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println(err.Error())
		return ""
	}
	resp.Body.Close()
	//
	result := struct {
		ErrorMsg string   `json:"err_msg"`
		ErrorNo  int      `json:"err_no"`
		Result   []string `json:"result"`
	}{}
	err = json.Unmarshal(bs, &result)
	if err != nil {
		log.Println(err.Error(), string(bs))
		return ""
	}
	if result.ErrorNo == 0 {
		return result.Result[0]
	} else {
		log.Println(result.ErrorMsg)
		return ""
	}
}
