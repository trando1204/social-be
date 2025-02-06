package webserver

import (
	"context"
	"fmt"
	"net/http"
	"socialat/be/atlib"
	"socialat/be/authpb"
	"socialat/be/storage"
	"socialat/be/utils"
	"socialat/be/webserver/portal"
	"time"

	"github.com/bluesky-social/indigo/xrpc"
)

type apiAuth struct {
	*WebServer
}

type authClaims struct {
	Id          uint64
	UserRole    utils.UserRole
	Expire      int64
	UserName    string
	DisplayName string
	AccessJwt   string
	RefreshJwt  string
	Handle      string
	Did         string
}

func (c authClaims) Valid() error {
	timestamp := time.Now().Unix()
	if timestamp >= c.Expire {
		return fmt.Errorf("the credential is expired")
	}
	return nil
}

func (a *apiAuth) CancelPasskeyRegister(w http.ResponseWriter, r *http.Request) {
	res, err := a.service.CancelRegisterHandler(r.Context(), &authpb.CancelRegisterRequest{
		SessionKey: r.FormValue("sessionKey"),
	})
	if err != nil {
		utils.Response(w, http.StatusInternalServerError, err, nil)
		return
	}
	if res.Error {
		utils.Response(w, http.StatusInternalServerError, fmt.Errorf(res.Msg), nil)
		return
	}
	utils.ResponseOK(w, res.Data)
}

func (a *apiAuth) StartPasskeyRegister(w http.ResponseWriter, r *http.Request) {
	username := r.FormValue("username")
	resData, err := a.service.BeginRegistrationHandler(r.Context(), &authpb.WithUsernameRequest{
		Username: username,
	})
	if err != nil {
		utils.Response(w, http.StatusInternalServerError, err, nil)
		return
	}

	if resData.Error {
		utils.Response(w, http.StatusInternalServerError, fmt.Errorf(resData.Msg), nil)
		return
	}
	utils.ResponseOK(w, resData.Data)
}

func (a *apiAuth) UpdatePasskeyStart(w http.ResponseWriter, r *http.Request) {
	//Get login sesssion token string
	loginToken := r.Header.Get("Authorization")
	if utils.IsEmpty(loginToken) {
		utils.Response(w, http.StatusInternalServerError, fmt.Errorf("get login token failed"), nil)
		return
	}
	res, err := a.service.BeginUpdatePasskeyHandler(r.Context(), &authpb.CommonRequest{
		AuthToken: loginToken,
	})
	if err != nil {
		utils.Response(w, http.StatusInternalServerError, err, nil)
		return
	}
	if res.Error {
		utils.Response(w, http.StatusInternalServerError, fmt.Errorf(res.Msg), nil)
		return
	}
	utils.ResponseOK(w, res.Data)
}

func (a *apiAuth) login(w http.ResponseWriter, r *http.Request) {
	var f portal.LoginForm
	err := a.parseJSON(r, &f)
	if err != nil {
		utils.Response(w, http.StatusBadRequest, err, nil)
		return
	}
	resData, err := a.service.LoginByPassword(r.Context(), &authpb.WithPasswordRequest{
		Username: f.UserName,
		Password: f.Password,
	})
	if err != nil {
		utils.Response(w, http.StatusInternalServerError, err, nil)
		return
	}
	var data map[string]any
	err = utils.JsonStringToObject(resData.Data, &data)
	if err != nil {
		utils.Response(w, http.StatusInternalServerError, err, nil)
		return
	}
	var authClaim storage.AuthClaims
	userClaims, userExist := data["user"]
	token, tokenExist := data["token"]
	if !userExist || !tokenExist {
		utils.Response(w, http.StatusInternalServerError, fmt.Errorf("Get login token failed"), nil)
		return
	}
	tokenString := token.(string)
	err = utils.CatchObject(userClaims, &authClaim)
	if err != nil {
		utils.Response(w, http.StatusInternalServerError, err, nil)
		return
	}
	handle := utils.GetHandleFromUsername(a.conf.PdsServer, authClaim.Username)
	// get pds user from db
	pdsUser, err := a.service.GetPdsUserByHandle(handle)
	if err != nil {
		log.Errorf("pds user not exist. %v", err)
		utils.Response(w, http.StatusInternalServerError, err, nil)
		return
	}
	// connect to pds server
	ctx := context.Background()
	agent := atlib.NewAgent(ctx, a.conf.PdsServer, handle, pdsUser.Password)
	jwtOut, err := agent.Connect(ctx)
	if err != nil {
		log.Errorf("connect to pds server failed. %v", err)
		utils.Response(w, http.StatusInternalServerError, err, nil)
		return
	}
	utils.ResponseOK(w, Map{
		"token":     tokenString,
		"loginType": int(storage.AuthLocalUsernamePassword),
		"userInfo":  authClaim,
		"pdsJwt":    jwtOut,
	})
}

func (a *apiAuth) register(w http.ResponseWriter, r *http.Request) {
	var f portal.RegisterForm
	err := a.parseJSONAndValidate(r, &f)
	if err != nil {
		utils.Response(w, http.StatusBadRequest, err, nil)
		return
	}
	resData, err := a.service.RegisterByPassword(r.Context(), &authpb.WithPasswordRequest{
		Username: f.UserName,
		Password: f.Password,
	})
	if err != nil {
		utils.Response(w, http.StatusInternalServerError, err, nil)
		return
	}
	var data map[string]any
	err = utils.JsonStringToObject(resData.Data, &data)
	if err != nil {
		utils.Response(w, http.StatusInternalServerError, err, nil)
		return
	}
	var authClaim storage.AuthClaims
	userClaims, userExist := data["user"]
	token, tokenExist := data["token"]
	if !userExist || !tokenExist {
		utils.Response(w, http.StatusInternalServerError, fmt.Errorf("Get login token failed"), nil)
		return
	}
	tokenString := token.(string)
	err = utils.CatchObject(userClaims, &authClaim)
	if err != nil {
		utils.Response(w, http.StatusInternalServerError, err, nil)
		return
	}
	// create bluesky pds account
	pdsJwt, err := a.CreateBlueskyPdsAccount(&authClaim, f.Email)
	if err != nil {
		utils.Response(w, http.StatusInternalServerError, err, nil)
		return
	}
	//handler login imediately
	utils.ResponseOK(w, Map{
		"token":     tokenString,
		"loginType": int(storage.AuthLocalUsernamePassword),
		"userInfo":  authClaim,
		"pdsJwt":    pdsJwt,
	})
}

func (a *apiAuth) CreateBlueskyPdsAccount(authClaim *storage.AuthClaims, email string) (*xrpc.AuthInfo, error) {
	ctx := context.Background()
	// create invite code
	inviteCode, err := atlib.CreateInviteCode(ctx, a.conf.PdsServer, a.conf.PdsAdminToken)
	if err != nil {
		log.Errorf("Create pds invite code failed. %v", err)
		return nil, err
	}
	passRandom := utils.RandSeq(16)
	accountRes, err := atlib.CreateAccount(ctx, a.conf.PdsServer, authClaim.Username, passRandom, email, inviteCode)
	if err != nil {
		log.Errorf("Create pds account failed. %v", err)
		return nil, err
	}
	now := time.Now()
	pdsUser := storage.PdsUser{
		Handle:     accountRes.Handle,
		Password:   passRandom,
		Email:      email,
		Did:        accountRes.Did,
		InviteCode: inviteCode,
		CreatedAt:  now,
		UpdatedAt:  now,
	}
	if err = a.db.CreatePdsUser(&pdsUser); err != nil {
		log.Errorf("Create Pds user on local db failed. %v", err)
		return nil, err
	}
	return &xrpc.AuthInfo{
		Handle:     accountRes.Handle,
		Did:        accountRes.Did,
		AccessJwt:  accountRes.AccessJwt,
		RefreshJwt: accountRes.RefreshJwt,
	}, nil
}

func (a *apiAuth) UpdatePasskeyFinish(w http.ResponseWriter, r *http.Request) {
	res, err := a.service.FinishUpdatePasskeyHandler(r.Context(), &authpb.FinishUpdatePasskeyRequest{
		Common: &authpb.CommonRequest{
			AuthToken: r.Header.Get("Authorization"),
		},
		Request: &authpb.HttpRequest{
			BodyJson: utils.RequestBodyToString(r.Body),
		},
		SessionKey: r.FormValue("sessionKey"),
		IsReset:    r.FormValue("isReset") == "true",
	})

	if err != nil {
		utils.Response(w, http.StatusInternalServerError, err, nil)
		return
	}
	if res.Error {
		utils.Response(w, http.StatusInternalServerError, fmt.Errorf(res.Msg), nil)
		return
	}
	var data map[string]any
	err = utils.JsonStringToObject(res.Data, &data)
	if err != nil {
		utils.Response(w, http.StatusInternalServerError, err, nil)
	}
	var authClaim storage.AuthClaims
	tokenString := data["token"].(string)
	err = utils.CatchObject(data["user"], &authClaim)
	if err != nil {
		utils.Response(w, http.StatusInternalServerError, err, nil)
		return
	}
	//reset auth info
	utils.ResponseOK(w, Map{
		"token":    tokenString,
		"userInfo": authClaim,
	})
}

func (a *apiAuth) FinishPasskeyTransferRegister(w http.ResponseWriter, r *http.Request) {
	sessionKey := r.FormValue("sessionKey")
	res, err := a.service.FinishRegistrationHandler(r.Context(), &authpb.SessionKeyAndHttpRequest{
		SessionKey: sessionKey,
		Request: &authpb.HttpRequest{
			BodyJson: utils.RequestBodyToString(r.Body),
		},
	})
	if err != nil {
		utils.Response(w, http.StatusInternalServerError, err, nil)
		return
	}
	if res.Error {
		utils.Response(w, http.StatusInternalServerError, fmt.Errorf(res.Msg), nil)
		return
	}

	var data map[string]any
	err = utils.JsonStringToObject(res.Data, &data)
	if err != nil {
		utils.Response(w, http.StatusInternalServerError, err, nil)
		return
	}
	var authClaim storage.AuthClaims
	userClaims, userExist := data["user"]
	token, tokenExist := data["token"]
	if !userExist || !tokenExist {
		utils.Response(w, http.StatusInternalServerError, fmt.Errorf("get login token failed"), nil)
		return
	}
	tokenString := token.(string)
	err = utils.CatchObject(userClaims, &authClaim)
	if err != nil {
		utils.Response(w, http.StatusInternalServerError, err, nil)
		return
	}

	//handler login imediately
	utils.ResponseOK(w, Map{
		"token":     tokenString,
		"loginType": int(storage.AuthMicroservicePasskey),
		"userInfo":  authClaim,
	})
}

func (a *apiAuth) FinishPasskeyRegister(w http.ResponseWriter, r *http.Request) {
	sessionKey := r.FormValue("sessionKey")
	email := r.FormValue("email")
	if utils.IsEmpty(email) {
		utils.Response(w, http.StatusInternalServerError, fmt.Errorf("get email failed"), nil)
		return
	}
	res, err := a.service.FinishRegistrationHandler(r.Context(), &authpb.SessionKeyAndHttpRequest{
		SessionKey: sessionKey,
		Request: &authpb.HttpRequest{
			BodyJson: utils.RequestBodyToString(r.Body),
		},
	})
	if err != nil {
		utils.Response(w, http.StatusInternalServerError, err, nil)
		return
	}
	if res.Error {
		utils.Response(w, http.StatusInternalServerError, fmt.Errorf(res.Msg), nil)
		return
	}
	var data map[string]any
	err = utils.JsonStringToObject(res.Data, &data)
	if err != nil {
		utils.Response(w, http.StatusInternalServerError, err, nil)
		return
	}
	var authClaim storage.AuthClaims
	userClaims, userExist := data["user"]
	token, tokenExist := data["token"]
	if !userExist || !tokenExist {
		utils.Response(w, http.StatusInternalServerError, fmt.Errorf("get login token failed"), nil)
		return
	}
	tokenString := token.(string)
	err = utils.CatchObject(userClaims, &authClaim)
	if err != nil {
		utils.Response(w, http.StatusInternalServerError, err, nil)
		return
	}
	// create bluesky pds account
	pdsJwt, err := a.CreateBlueskyPdsAccount(&authClaim, email)
	if err != nil {
		utils.Response(w, http.StatusInternalServerError, err, nil)
		return
	}
	//handler login imediately
	utils.ResponseOK(w, Map{
		"token":     tokenString,
		"loginType": int(storage.AuthMicroservicePasskey),
		"userInfo":  authClaim,
		"pdsJwt":    pdsJwt,
	})
}

func (a *apiAuth) GenRandomUsername(w http.ResponseWriter, r *http.Request) {
	res, err := a.service.GenRandomUsernameHandler(r.Context())
	if err != nil {
		utils.Response(w, http.StatusInternalServerError, err, nil)
		return
	}
	utils.ResponseOK(w, res)
}

func (a *apiAuth) CheckAuthUsername(w http.ResponseWriter, r *http.Request) {
	username := r.FormValue("userName")
	res, err := a.service.CheckUserHandler(r.Context(), &authpb.WithUsernameRequest{
		Username: username,
	})
	if err != nil {
		utils.Response(w, http.StatusInternalServerError, err, nil)
		return
	}
	var resData map[string]bool
	err = utils.JsonStringToObject(res.Data, &resData)
	if err != nil {
		utils.Response(w, http.StatusInternalServerError, err, nil)
		return
	}
	exist := resData["exist"]
	utils.ResponseOK(w, Map{
		"found": exist,
	})
}

func (a *apiAuth) AssertionResult(w http.ResponseWriter, r *http.Request) {
	sessionKey := r.FormValue("sessionKey")
	res, err := a.service.AssertionResultHandler(r.Context(), &authpb.SessionKeyAndHttpRequest{
		SessionKey: sessionKey,
		Request: &authpb.HttpRequest{
			BodyJson: utils.RequestBodyToString(r.Body),
		},
	})
	if err != nil {
		utils.Response(w, http.StatusInternalServerError, err, nil)
		return
	}

	var data map[string]any
	err = utils.JsonStringToObject(res.Data, &data)
	if err != nil {
		utils.Response(w, http.StatusInternalServerError, fmt.Errorf("parse res data failed"), nil)
		return
	}
	tokenString := ""
	var authClaim storage.AuthClaims
	tokenString, _ = data["token"].(string)
	err = utils.CatchObject(data["user"], &authClaim)
	if err != nil {
		utils.Response(w, http.StatusInternalServerError, err, nil)
		return
	}
	handle := utils.GetHandleFromUsername(a.conf.PdsServer, authClaim.Username)
	// get pds user from db
	pdsUser, err := a.service.GetPdsUserByHandle(handle)
	if err != nil {
		log.Errorf("pds user not exist. %v", err)
		utils.Response(w, http.StatusInternalServerError, err, nil)
		return
	}
	// connect to pds server
	ctx := context.Background()
	agent := atlib.NewAgent(ctx, a.conf.PdsServer, handle, pdsUser.Password)
	jwtOut, err := agent.Connect(ctx)
	if err != nil {
		log.Errorf("connect to pds server failed. %v", err)
		utils.Response(w, http.StatusInternalServerError, err, nil)
		return
	}
	utils.ResponseOK(w, Map{
		"token":     tokenString,
		"loginType": int(storage.AuthMicroservicePasskey),
		"userInfo":  authClaim,
		"pdsJwt":    jwtOut,
	})
}

func (a *apiAuth) AssertionOptions(w http.ResponseWriter, r *http.Request) {
	responseData, err := a.service.AssertionOptionsHandler(r.Context())
	if err != nil {
		utils.Response(w, http.StatusInternalServerError, err, nil)
		return
	}
	utils.ResponseOK(w, responseData)
}

func (a *apiAuth) getAuthMethod(w http.ResponseWriter, r *http.Request) {
	authType := a.service.Conf.AuthType
	if authType != int(storage.AuthLocalUsernamePassword) && authType != int(storage.AuthMicroservicePasskey) {
		authType = int(storage.AuthLocalUsernamePassword)
	}
	utils.ResponseOK(w, authType)
}
