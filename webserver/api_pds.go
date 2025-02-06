package webserver

import (
	"context"
	"net/http"
	"socialat/be/atlib"
	"socialat/be/utils"
	"socialat/be/webserver/portal"

	"github.com/bluesky-social/indigo/xrpc"
)

type apiPds struct {
	*WebServer
}

func (a *apiPds) getPdsSession(w http.ResponseWriter, r *http.Request) {
	claims, _ := a.credentialsInfo(r)
	ctx := context.Background()
	authInfo := &xrpc.AuthInfo{
		AccessJwt:  claims.AccessJwt,
		RefreshJwt: claims.RefreshJwt,
		Handle:     claims.Handle,
		Did:        claims.Did,
	}
	err := atlib.HandlerValidSession(ctx, a.conf.PdsServer, authInfo)
	// if get pdf session successfully
	if err != nil {
		// get pds user
		handle := utils.GetHandleFromUsername(a.conf.PdsServer, claims.UserName)
		// get pds user from db
		pdsUser, err := a.service.GetPdsUserByHandle(handle)
		if err != nil {
			log.Errorf("get pds user failed. %v", err)
			utils.Response(w, http.StatusInternalServerError, err, nil)
			return
		}
		// connect to pds server to get new session
		authInfo, err = atlib.ConnectToGetSession(ctx, a.conf.PdsServer, handle, pdsUser.Password)
		if err != nil {
			log.Errorf("create new pds session failed. %v", err)
			utils.Response(w, http.StatusInternalServerError, err, nil)
			return
		}
	}
	utils.ResponseOK(w, authInfo)
}

func (a *apiPds) getPdsTimeline(w http.ResponseWriter, r *http.Request) {
	var timeLineReq portal.GetTimelineRequest
	a.parseJSONAndValidate(r, &timeLineReq)
	claims, _ := a.credentialsInfo(r)
	ctx := context.Background()
	pdsAgent := atlib.NewBasicAgent(ctx, a.conf.PdsServer)
	pdsAgent.SetClientAuth(claims.AccessJwt, claims.RefreshJwt, claims.Handle, claims.Did)

	timeLineOutput, err := pdsAgent.GetTimeline(ctx, "", int64(utils.LimitOfFetchTimeline))

	if err != nil {
		log.Errorf("get timeline failed: %v", err)
		utils.Response(w, http.StatusInternalServerError, err, nil)
		return
	}
	utils.ResponseOK(w, timeLineOutput)
}
