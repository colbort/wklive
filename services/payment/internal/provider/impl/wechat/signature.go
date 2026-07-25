package wechat

import (
	"fmt"

	"wklive/services/payment/models"
)

func sign(_ string, _ string, _ []byte, account *models.TTenantPayAccount) (string, error) {
	if err := validateAccount(account); err != nil {
		return "", err
	}
	return "", fmt.Errorf("%w: wechat sign", ErrNotImplemented)
}

func verify(_ string, _ string, _ string, _ []byte, account *models.TTenantPayAccount) error {
	if err := validateAccount(account); err != nil {
		return err
	}
	return fmt.Errorf("%w: wechat verify", ErrNotImplemented)
}

func decryptNotify(_ NotifyResource, account *models.TTenantPayAccount) ([]byte, error) {
	if err := validateAccount(account); err != nil {
		return nil, err
	}
	return nil, fmt.Errorf("%w: wechat notification decrypt", ErrNotImplemented)
}
