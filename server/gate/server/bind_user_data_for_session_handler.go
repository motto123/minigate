package server

import (
	"com.minigame.component/log"
	"com.minigame.proto/common"
	"com.minigame.proto/gate"
	"context"
)

type GateGrpcSrv struct{}

func (g GateGrpcSrv) BindUserDataForSession(ctx context.Context, req *gate.BindUserDataForSessionReq) (resp *gate.GeneralResp, err error) {
	resp = new(gate.GeneralResp)
	resp.Code = common.InternalCode_InternalError1

	if req.User == nil {
		log.Debugf(tag, "invalid params")
		resp.Code = common.InternalCode_InvalidParams1
		return
	}

	session2 := Srv.sessionM.Load(req.Sid)
	if session2 == nil {
		log.Errorf(tag, "sessionM.Load failed, because session2 %d is nil", req.Sid)
		return
	}

	session2.SetCustomData(req.User.Id, req.User.Nickname)
	Srv.sessionM.StoreWithUid(req.User.Id, session2)

	resp.Code = common.InternalCode_Ok1
	return
}
