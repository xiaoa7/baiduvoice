package main

import (
	b "baiduvoice"
	"io/ioutil"

	"fmt"

	"log"

	"net/http"

	"time"
)

//
var bv b.BaiduVoice

//
func init() {
	bv = b.BaiduVoice{
		AppID:     "11078892",
		APIKey:    "29UihGD07uB9GPcASfWR4Xcg",
		SecretKey: "4b61b1464e7c24bc604132adb7df205e",
		Expires:   time.Now(),
	}

}

//
func main() {
	http.Handle("/", http.FileServer(http.Dir("./res")))
	//
	http.HandleFunc("/tts", func(w http.ResponseWriter, r *http.Request) {
		txt := r.FormValue("txt")
		w.Header().Add("Content-Type", "audio/mp3")
		bv.Tts(txt, 0, 5, 5, 7, w)
	})
	//
	http.HandleFunc("/asr", func(w http.ResponseWriter, r *http.Request) {
		bs, _ := ioutil.ReadAll(r.Body)
		r.Body.Close()
		txt := bv.Asr(bs)
		fmt.Fprint(w, txt)
	})
	//
	err := http.ListenAndServeTLS(":443", "./certificate.pem", "./privatekey.pem", nil)
	if err != nil {
		log.Println(err.Error())
	}
}
