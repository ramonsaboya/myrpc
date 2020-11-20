package server

import (
	"github.com/ramonsaboya/myrpc/commons"
	"github.com/ramonsaboya/myrpc/miop"
)

type CalculatorInvoker struct {
	proxy commons.ClientProxy
}

func NewCalculatorInvoker(proxy *commons.ClientProxy) *CalculatorInvoker {
	return &CalculatorInvoker{
		proxy: *proxy,
	}
}

func (c *CalculatorInvoker) Invoke() error {
	srh, err := NewSRH(c.proxy.Protocol, c.proxy.Host, c.proxy.Port)
	if err != nil {
		return err
	}
	marshaller := commons.Marshaller{}
	calculator := Calculator{}
	res := miop.Packet{}
	var reply interface{}

	for {
		rcvMsgBytes, err := srh.Receive()
		if err != nil {
			return err
		}

		req, err := marshaller.Unmarshall(rcvMsgBytes)
		if err != nil {
			return err
		}
		operation := req.Bd.ReqHeader.Operation

		switch operation {
		case "EquationRoots":
			_a := int(req.Bd.ReqBody.Body[0].(float64))
			_b := int(req.Bd.ReqBody.Body[1].(float64))
			_c := int(req.Bd.ReqBody.Body[2].(float64))
			reply = calculator.EquationRoots(_a, _b, _c)
		}

		repHeader := miop.ReplyHeader{RequestId: req.Bd.ReqHeader.RequestId, Status: 200}
		repBody := miop.ReplyBody{OperationResult: reply}
		header := miop.Header{MessageType: commons.MIOPREQUEST}
		body := miop.Body{RepHeader: repHeader, RepBody: repBody}
		res = miop.Packet{Hdr: header, Bd: body}

		msgToClientBytes, err := marshaller.Marshall(res)
		if err != nil {
			return err
		}

		err = srh.Send(msgToClientBytes)
		if err != nil {
			return err
		}
	}
}
