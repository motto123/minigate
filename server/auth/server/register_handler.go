package server

import (
	"com.minigame.component/base"
	"com.minigame.component/codec/message"
	"com.minigame.component/log"
	"com.minigame.proto/auth"
	"com.minigame.proto/common"
	"com.minigame.proto/gate"
	"database/sql"
	"github.com/pkg/errors"
	"google.golang.org/protobuf/proto"
)

func registerHandler(s *base.BusinessServer, sessionId uint64, msg *message.Message, ri *gate.ReceiveInfo) (reqMsg, respMsg proto.Message) {
	var err error
	req := new(auth.RegisterReq)
	resp := &auth.RegisterResp{CodeInfo: &common.ResponseCode{Code: common.Code_InternalError}}

	defer func() {
		s.DefaultErrProcess(resp.CodeInfo, err)
		msg.Type = message.Response
		reqMsg = req
		respMsg = resp
		ri.ReceiverSids = append(ri.ReceiverSids, sessionId)
	}()

	if len(msg.Data) == 0 {
		resp.CodeInfo.Code = common.Code_InvalidParams
		log.Debugf(tag, "invalid params")
		return
	}

	err = proto.Unmarshal(msg.Data, req)
	err = errors.WithStack(err)
	if err != nil {
		log.Errorf(tag, "proto.Unmarshal failed, err: %+v", err)
		return
	}

	if req.Account == "" || req.Password == "" {
		resp.CodeInfo.Code = common.Code_IncorrectAccountOrPassword
		log.Debugf(tag, "invalid params")
		return
	}
	//TODO: register logic
	var uid int64
	row := s.MysqlDb.QueryRow(`select id from user where account = ? and password = ?;`, req.Account, req.Password)
	err = row.Scan(&uid)
	if err != nil && err != sql.ErrNoRows {
		log.Errorf(tag, "row.Scan failed, err: %+v", err)
		return
	}
	if uid != 0 {
		resp.CodeInfo.Code = common.Code_UserExisting
		return
	}
	result, err := s.MysqlDb.Exec(`insert into user(nickname, account, password) values(?,?,?);`,
		req.Nickname, req.Account, req.Password)
	if err != nil {
		log.Errorf(tag, "MysqlDb.Exec failed, err: %+v", err)
		return
	}
	affected, err := result.RowsAffected()
	if err != nil {
		err = errors.WithStack(err)
		log.Errorf(tag, "result.RowsAffected failed, err: %+v", err)
		return
	}
	if affected == 0 {
		resp.CodeInfo.Code = common.Code_OperationFailed
		err := errors.New("insert failed, because affected is 0")
		log.Errorf(tag, "%+v", err)
		return
	}

	resp.CodeInfo.Code = common.Code_Ok
	id, _ := result.LastInsertId()
	resp.Id = id
	return
}
