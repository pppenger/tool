package main

import "testing"

const TestFinanceMsg = `2021-11-11 20:09:28.793 utils/store.go:71 ERROR AddFinancialMsg|mysql error: msg &{Id:0 SendUid:9366696 SendMsgid:2062229660002333406 SendCreateTime:2021-11-11 20:09:26 +0800 CST BillId:FK1142404025138965530353967281 WithdrawBillId: ReplyBillId:FK1149664027138968142854724601 Status:0 ReplyUid:9793406 ReplyMsgid:0 ReplyCreateTime:2021-11-11 20:09:26 +0800 CST Coins:2 Points:200 IsSysSend:0 MsgType:1 ChatUpGiftId:0}, err invalid connection {"trace_id":"3064f70f562caf6d8414af22","namespace":"localmeet"}`

func TestParseToFinanceMsg(t *testing.T) {
	msg, err := ParseToFinanceMsg(TestFinanceMsg)
	if err != nil {
		t.Fatalf("ParseToFinanceMsg err； %v", err)
	}

	t.Logf("msg -> %+v", msg)
}
