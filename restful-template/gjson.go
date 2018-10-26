package main

import (
	"gitee.com/johng/gf/g/encoding/gjson"
	"fmt"
)




func main() {


	b := `{"id":"ch_qLS0SS4afbb98y5K0GqXzHGS","object":"charge","created":1522308204,"livemode":true,"paid":false,"refunded":false,"reversed":false,"app":"app_y9qLmL00yjbPaDOu","channel":"wx_pub_qr","order_no":"10000046","client_ip":"127.0.0.1","amount":1,"amount_settle":1,"currency":"cny","subject":"Your Subject","body":"Your Body","extra":{"product_id":"1001"},"time_paid":0,"time_expire":1522315404,"time_settle":0,"transaction_no":"","refunds":{"object":"list","has_more":false,"url":"/v1/charges/ch_qLS0SS4afbb98y5K0GqXzHGS/refunds","data":[]},"amount_refunded":0,"failure_code":"","failure_msg":"","metadata":{},"credential":{},"description":""}`


	//a:=`{"id":"ch_qLS0SS4afbb98y5K0GqXzHGS","object":"charge","created":1522308204,"livemode":true,"paid":false,"refunded":false,"reversed":false,"app":"app_y9qLmL00yjbPaDOu","channel":"wx_pub_qr","order_no":"10000046","client_ip":"127.0.0.1","amount":1,"amount_settle":1,"currency":"cny","subject":"Your Subject","body":"Your Body","extra":{"product_id":"1001"},"time_paid":0,"time_expire":1522315404,"time_settle":0,"transaction_no":"","refunds":{"object":"list","has_more":false,"url":"/v1/charges/ch_qLS0SS4afbb98y5K0GqXzHGS/refunds","data":[]},"amount_refunded":0,"failure_code":"","failure_msg":"","metadata":{},"credential":{},"description":""}`
	a :=`{"id": "evt_401180104103535002411103", "created": 1515033335, "livemode": true,
"type": "order.succeeded", "data": {"object": {"id": "2011801040000059652"}}}`

	g1,_:=gjson.DecodeToJson([]byte(b))
	fmt.Println(g1.GetString("id"))
	g2,err :=gjson.DecodeToJson([]byte(a))
	if err != nil{
		fmt.Println(err)
	}else {
		fmt.Println(g2.GetString("data.object.id"))
	}

}
