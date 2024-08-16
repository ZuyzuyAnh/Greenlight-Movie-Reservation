package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/zpmep/hmacutil"
	"io"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

type object map[string]interface{}

var (
	app_id = "2554"
	key1   = "sdngKKJmqEMzvh5QQcdD2A9XBSKUNaYn"
	key2   = "trMrHtvjo6myautxDUiAcYsVtaeQ8nhf"
)

type Params struct {
	AppUser       string
	ItemPrice     string
	ReservationId string
}

type ZaloResponse struct {
	Code int    `json:"return_code"`
	Msg  string `json:"return_message"`
	Url  string `json:"order_url"`
}

func CreaterOrder(p Params) (*ZaloResponse, error) {
	embedData, _ := json.Marshal(object{})
	items, _ := json.Marshal([]object{})

	// request data
	params := make(url.Values)
	params.Add("app_id", app_id)
	params.Add("amount", p.ItemPrice)
	params.Add("app_user", p.AppUser)
	params.Add("embed_data", string(embedData))
	params.Add("item", string(items))
	params.Add("description", "Lazada - Payment for the order")
	params.Add("bank_code", "zalopayapp")
	params.Add("callback_url", "https://4d92-2405-4802-1ce4-f9a0-c8b5-9cbf-2550-184a.ngrok-free.app/callback")
	params.Add("app_time", strconv.FormatInt(time.Now().UnixNano()/int64(time.Millisecond), 10)) // miliseconds

	params.Add("app_trans_id", p.ReservationId) // translation missing: vi.docs.shared.sample_code.comments.app_trans_id

	// appid|app_trans_id|appuser|amount|apptime|embeddata|item
	data := fmt.Sprintf("%v|%v|%v|%v|%v|%v|%v", params.Get("app_id"), params.Get("app_trans_id"), params.Get("app_user"), params.Get("amount"), params.Get("app_time"), params.Get("embed_data"), params.Get("item"))
	params.Add("mac", hmacutil.HexStringEncode(hmacutil.SHA256, key1, data))

	// Content-Type: application/x-www-form-urlencoded
	res, err := http.PostForm("https://sb-openapi.zalopay.vn/v2/create", params)

	// parse response
	if err != nil {
		log.Fatal("error parse: ", err)
	}
	defer res.Body.Close()

	body, _ := io.ReadAll(res.Body)

	var result ZaloResponse

	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}

	fmt.Println(result.Code)

	if result.Code != 1 {
		return nil, errors.New(result.Msg)
	}

	return &result, nil
}

func (app *application) zaloPayCallBackHandler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	var cbdata map[string]interface{}

	decoder := json.NewDecoder(r.Body)
	decoder.Decode(&cbdata)

	requestMac := cbdata["mac"].(string)
	dataStr := cbdata["data"].(string)

	var dataMap map[string]interface{}

	_ = json.Unmarshal([]byte(dataStr), &dataMap)

	mac := hmacutil.HexStringEncode(hmacutil.SHA256, key2, dataStr)

	result := make(map[string]interface{})

	if mac != requestMac {
		result["return_code"] = -1
		result["return_message"] = "mac not equal"
	} else {
		result["return_code"] = 1
		result["return_message"] = "success"

		fmt.Println("Sending:", dataMap["app_trans_id"].(string))
		value, _ := strconv.ParseInt(dataMap["app_trans_id"].(string)[7:], 10, 64)

		app.transChannel <- value

		var dataJSON map[string]interface{}
		json.Unmarshal([]byte(dataStr), &dataJSON)
	}

	resultJSON, _ := json.Marshal(result)
	fmt.Fprintf(w, "%s", resultJSON)
}
