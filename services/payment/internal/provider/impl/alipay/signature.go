package alipay

import (
	"fmt"

	"wklive/services/payment/models"
)

func sign(_ []byte, account *models.TTenantPayAccount) (string, error) {
	if err := validateAccount(account); err != nil {
		return "", err
	}
	return "", fmt.Errorf("%w: alipay sign", ErrNotImplemented)
}

func verify(_ []byte, _ string, account *models.TTenantPayAccount) error {
	if err := validateAccount(account); err != nil {
		return err
	}
	return fmt.Errorf("%w: alipay verify", ErrNotImplemented)
}
