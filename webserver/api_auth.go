package webserver

import (
	"fmt"
	"net/http"
	"socialat/be/authpb"
	"socialat/be/storage"
	"socialat/be/utils"
	"time"
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

	utils.ResponseOK(w, Map{
		"token":     tokenString,
		"loginType": int(storage.AuthMicroservicePasskey),
		"userInfo":  authClaim,
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
