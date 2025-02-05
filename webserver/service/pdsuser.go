package service

import (
	"fmt"
	"socialat/be/storage"
	"socialat/be/utils"

	"gorm.io/gorm"
)

func (s *Service) GetPdsUserByHandle(handle string) (*storage.PdsUser, error) {
	var user storage.PdsUser
	if err := s.db.Where("handle = ?", handle).Find(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, utils.NewError(fmt.Errorf("pds user not found"), utils.ErrorNotFound)
		}
		log.Error("GetPdsUserInfo:get pds user info fail with error: ", err)
		return nil, err
	}
	return &user, nil
}
