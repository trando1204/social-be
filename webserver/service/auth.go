package service

import (
	"context"
	"fmt"
	"socialat/be/authpb"
	"socialat/be/storage"
	"socialat/be/utils"

	socketio "github.com/googollee/go-socket.io"
	"google.golang.org/grpc"
	"gorm.io/gorm"
)

type Config struct {
	AuthType int    `yaml:"authType"`
	AuthHost string `yaml:"authHost"`
}

type Service struct {
	db         *gorm.DB
	Conf       Config
	socket     *socketio.Server
	AuthClient *authpb.AuthServiceClient
}

func NewService(conf Config, db *gorm.DB, socket *socketio.Server) *Service {
	var authClient *authpb.AuthServiceClient
	if conf.AuthType == int(storage.AuthMicroservicePasskey) {
		authClient = InitAuthClient(conf.AuthHost)
	}
	return &Service{
		db:         db,
		Conf:       conf,
		socket:     socket,
		AuthClient: authClient,
	}
}

func InitAuthClient(authUrl string) *authpb.AuthServiceClient {
	log.Infof("API Gateway :  InitAuthClient")
	//	using WithInsecure() because no SSL running
	cc, err := grpc.Dial(authUrl, grpc.WithInsecure())

	if err != nil {
		log.Infof("Could not connect to auth service:", err)
		return nil
	}
	client := authpb.NewAuthServiceClient(cc)
	return &client
}

func (s *Service) CheckAndInitAuthClient() error {
	if s.AuthClient != nil {
		return nil
	}
	s.AuthClient = InitAuthClient(s.Conf.AuthHost)
	if s.AuthClient == nil {
		return fmt.Errorf("init auth client failed")
	}
	return nil
}

func (s *Service) CheckMiddlewareLogin(ctx context.Context, req *authpb.CommonRequest) (bool, error) {
	err := s.CheckAndInitAuthClient()
	if err != nil {
		return false, err
	}
	_, err = (*s.AuthClient).IsLoggingOn(ctx, req)
	if err != nil {
		return false, utils.HandlerRPCError(err)
	}
	return true, nil
}

func (s *Service) GetAuthClaimsLogin(ctx context.Context, req *authpb.CommonRequest) (*authpb.ResponseData, error) {
	err := s.CheckAndInitAuthClient()
	if err != nil {
		return nil, err
	}
	res, err := (*s.AuthClient).IsLoggingOn(ctx, req)
	if err != nil {
		return res, utils.HandlerRPCError(err)
	}
	return res, nil
}

func (s *Service) BeginRegistrationHandler(ctx context.Context, req *authpb.WithUsernameRequest) (*authpb.ResponseData, error) {
	err := s.CheckAndInitAuthClient()
	if err != nil {
		return nil, err
	}
	res, err := (*s.AuthClient).BeginRegistration(ctx, req)
	if err != nil {
		return res, utils.HandlerRPCError(err)
	}
	return res, nil
}

func (s *Service) CancelRegisterHandler(ctx context.Context, req *authpb.CancelRegisterRequest) (*authpb.ResponseData, error) {
	err := s.CheckAndInitAuthClient()
	if err != nil {
		return nil, err
	}
	res, err := (*s.AuthClient).CancelRegister(ctx, req)
	if err != nil {
		return res, utils.HandlerRPCError(err)
	}
	return res, nil
}

func (s *Service) BeginUpdatePasskeyHandler(ctx context.Context, req *authpb.CommonRequest) (*authpb.ResponseData, error) {
	err := s.CheckAndInitAuthClient()
	if err != nil {
		return nil, err
	}
	res, err := (*s.AuthClient).BeginUpdatePasskey(ctx, req)
	if err != nil {
		return res, utils.HandlerRPCError(err)
	}
	return res, nil
}

func (s *Service) FinishUpdatePasskeyHandler(ctx context.Context, req *authpb.FinishUpdatePasskeyRequest) (*authpb.ResponseData, error) {
	err := s.CheckAndInitAuthClient()
	if err != nil {
		return nil, err
	}
	res, err := (*s.AuthClient).FinishUpdatePasskey(ctx, req)
	if err != nil {
		return res, utils.HandlerRPCError(err)
	}
	return res, nil
}

func (s *Service) FinishRegistrationHandler(ctx context.Context, req *authpb.SessionKeyAndHttpRequest) (*authpb.ResponseData, error) {
	err := s.CheckAndInitAuthClient()
	if err != nil {
		return nil, err
	}
	res, err := (*s.AuthClient).FinishRegistration(ctx, req)
	if err != nil {
		return res, utils.HandlerRPCError(err)
	}
	return res, nil
}

func (s *Service) AssertionOptionsHandler(ctx context.Context) (*authpb.ResponseData, error) {
	err := s.CheckAndInitAuthClient()
	if err != nil {
		return nil, err
	}
	res, err := (*s.AuthClient).AssertionOptions(ctx, &authpb.CommonRequest{})
	if err != nil {
		return res, utils.HandlerRPCError(err)
	}
	return res, nil
}

func (s *Service) AssertionResultHandler(ctx context.Context, req *authpb.SessionKeyAndHttpRequest) (*authpb.ResponseData, error) {
	err := s.CheckAndInitAuthClient()
	if err != nil {
		return nil, err
	}
	res, err := (*s.AuthClient).AssertionResult(ctx, req)
	if err != nil {
		return res, utils.HandlerRPCError(err)
	}
	return res, nil
}

func (s *Service) BeginConfirmPasskeyHandler(ctx context.Context, req *authpb.CommonRequest) (*authpb.ResponseData, error) {
	err := s.CheckAndInitAuthClient()
	if err != nil {
		return nil, err
	}
	res, err := (*s.AuthClient).BeginConfirmPasskey(ctx, req)
	if err != nil {
		return res, utils.HandlerRPCError(err)
	}
	return res, nil
}

func (s *Service) FinishConfirmPasskeyHandler(ctx context.Context, req *authpb.SessionKeyAndHttpRequest) (*authpb.ResponseData, error) {
	err := s.CheckAndInitAuthClient()
	if err != nil {
		return nil, err
	}
	res, err := (*s.AuthClient).FinishConfirmPasskey(ctx, req)
	if err != nil {
		return res, utils.HandlerRPCError(err)
	}
	return res, nil
}

func (s *Service) ChangeUsernameFinishHandler(ctx context.Context, req *authpb.ChangeUsernameFinishRequest) (*authpb.ResponseData, error) {
	err := s.CheckAndInitAuthClient()
	if err != nil {
		return nil, err
	}
	res, err := (*s.AuthClient).ChangeUsernameFinish(ctx, req)
	if err != nil {
		return res, utils.HandlerRPCError(err)
	}
	return res, nil
}

func (s *Service) SyncUsernameDBHandler(ctx context.Context, req *authpb.SyncUsernameDBRequest) (*authpb.ResponseData, error) {
	err := s.CheckAndInitAuthClient()
	if err != nil {
		return nil, err
	}
	res, err := (*s.AuthClient).SyncUsernameDB(ctx, req)
	if err != nil {
		return res, utils.HandlerRPCError(err)
	}
	return res, nil
}

func (s *Service) GetAdminUserListHandler(ctx context.Context, req *authpb.CommonRequest) (*authpb.ResponseData, error) {
	err := s.CheckAndInitAuthClient()
	if err != nil {
		return nil, err
	}
	res, err := (*s.AuthClient).GetAdminUserList(ctx, req)
	if err != nil {
		return res, utils.HandlerRPCError(err)
	}
	return res, nil
}

func (s *Service) GetUserInfoByUsernameHandler(ctx context.Context, req *authpb.WithUsernameRequest) (*authpb.ResponseData, error) {
	err := s.CheckAndInitAuthClient()
	if err != nil {
		return nil, err
	}
	res, err := (*s.AuthClient).GetUserInfoByUsername(ctx, req)
	if err != nil {
		return res, utils.HandlerRPCError(err)
	}
	return res, nil
}

func (s *Service) GetExcludeLoginUserNameListHandler(ctx context.Context, req *authpb.CommonRequest) (*authpb.ResponseData, error) {
	err := s.CheckAndInitAuthClient()
	if err != nil {
		return nil, err
	}
	res, err := (*s.AuthClient).GetExcludeLoginUserNameList(ctx, req)
	if err != nil {
		return res, utils.HandlerRPCError(err)
	}
	return res, nil
}

func (s *Service) IsLoggingOnHandler(ctx context.Context, req *authpb.CommonRequest) (*authpb.ResponseData, error) {
	err := s.CheckAndInitAuthClient()
	if err != nil {
		return nil, err
	}
	res, err := (*s.AuthClient).IsLoggingOn(ctx, req)
	if err != nil {
		return res, utils.HandlerRPCError(err)
	}
	return res, nil
}

func (s *Service) GenRandomUsernameHandler(ctx context.Context) (*authpb.ResponseData, error) {
	err := s.CheckAndInitAuthClient()
	if err != nil {
		return nil, err
	}
	res, err := (*s.AuthClient).GenRandomUsername(ctx, &authpb.CommonRequest{})
	if err != nil {
		return res, utils.HandlerRPCError(err)
	}
	return res, nil
}

func (s *Service) CheckUserHandler(ctx context.Context, req *authpb.WithUsernameRequest) (*authpb.ResponseData, error) {
	err := s.CheckAndInitAuthClient()
	if err != nil {
		return nil, err
	}
	res, err := (*s.AuthClient).CheckUser(ctx, req)
	if err != nil {
		return res, utils.HandlerRPCError(err)
	}
	return res, nil
}

func (s *Service) RegisterByPassword(ctx context.Context, req *authpb.WithPasswordRequest) (*authpb.ResponseData, error) {
	err := s.CheckAndInitAuthClient()
	if err != nil {
		return nil, err
	}
	res, err := (*s.AuthClient).RegisterByPassword(ctx, req)
	if err != nil {
		return res, utils.HandlerRPCError(err)
	}
	return res, nil
}

func (s *Service) LoginByPassword(ctx context.Context, req *authpb.WithPasswordRequest) (*authpb.ResponseData, error) {
	err := s.CheckAndInitAuthClient()
	if err != nil {
		return nil, err
	}
	res, err := (*s.AuthClient).LoginByPassword(ctx, req)
	if err != nil {
		return res, utils.HandlerRPCError(err)
	}
	return res, nil
}
