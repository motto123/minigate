package server

import (
	"com.minigame.component/base"
	"com.minigame.component/codec/message"
	"com.minigame.component/log"
	"com.minigame.proto/auth"
	"com.minigame.proto/common"
	"com.minigame.proto/gate"
	"context"
	"database/sql"
	"github.com/pkg/errors"
	"google.golang.org/protobuf/proto"
)

func loginHandler(s *base.BusinessServer, sid uint64, msg *message.Message, ri *gate.ReceiveInfo) (reqMsg, respMsg proto.Message) {
	var err error
	req := new(auth.LoginReq)
	resp := &auth.LoginResp{CodeInfo: &common.ResponseCode{Code: common.Code_InternalError}}

	defer func() {
		s.DefaultErrProcess(resp.CodeInfo, err)
		msg.Type = message.Response
		reqMsg = req
		respMsg = resp
		ri.ReceiverSids = append(ri.ReceiverSids, sid)
	}()

	if len(msg.Data) == 0 {
		resp.CodeInfo.Code = common.Code_InvalidParams
		log.Debugf(tag, "invalid params")
		return
	}

	err = proto.Unmarshal(msg.Data, req)
	if err != nil {
		err = errors.WithStack(err)
		log.Errorf(tag, "proto.Unmarshal failed, err: %+v", err)
		return
	}

	if req.Account == "" || req.Password == "" {
		resp.CodeInfo.Code = common.Code_InvalidParams
		log.Debugf(tag, "invalid params")
		return
	}

	var u auth.User
	row := s.MysqlDb.QueryRow(`select id, nickname from user where account = ? and password = ?;`, req.Account, req.Password)
	err = row.Scan(&u.Id, &u.Nickname)
	if err != nil && err != sql.ErrNoRows {
		log.Errorf(tag, "row.Scan failed, err: %+v", err)
		return
	}
	if u.Id == 0 {
		resp.CodeInfo.Code = common.Code_IncorrectAccountOrPassword
		return
	}

	generalResp, err := Srv.gateSrvCli.BindUserDataForSession(context.Background(), &gate.BindUserDataForSessionReq{
		Sid:  sid,
		User: &u,
	})
	if err != nil {
		log.Errorf(tag, "gateSrvCli.BindUserDataForSession failed, err: %+v", err)
		return
	}
	if generalResp.Code != 0 {
		log.Errorf(tag, "gateSrvCli.BindUserDataForSession failed, internalCode: %s", generalResp.Code)
		return
	}

	resp.CodeInfo.Code = common.Code_Ok
	resp.User = &u
	return
}
