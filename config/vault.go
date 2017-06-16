package config

import (
	"errors"
	"fmt"
	"log"
	"os"

	"github.com/benschw/srv-lb/lb"
	vaultapi "github.com/hashicorp/vault/api"
)

func getRabbitmqPasswordFromVault(l lb.GenericLoadBalancer, secretsPath string) (string, error) {
	v, err := getVaultClient(l)
	if err != nil {
		return "", err
	}
	result, err := v.Read(secretsPath)
	if err != nil {
		return "", err
	}

	if result.Data != nil {
		if val, ok := result.Data["rabbitmq_password"]; ok {
			return val.(string), nil
		}
	}

	return "", errors.New("rabbitmq password not found in vault")
}

func getVaultClient(l lb.GenericLoadBalancer) (*vaultapi.Logical, error) {
	srvName := fmt.Sprintf("%s.service.consul", os.Getenv("VAULT_SERVICENAME"))

	a, err := l.Next(srvName)
	if err != nil {
		return nil, err
	}
	host := fmt.Sprintf("http://%s:%d", a.Address, a.Port)

	log.Printf("Using vault address: '%s'", host)

	cfg := vaultapi.DefaultConfig()
	cfg.Address = host
	client, err := vaultapi.NewClient(cfg)
	if err != nil {
		return nil, err
	}
	if appRolePath, ok := os.LookupEnv("VAULT_APPROLE_PATH"); ok {
		log.Println("getting approle")
		token := ""
		roleId, err := getVaultRoleId(appRolePath)
		if err != nil {
			return nil, err
		}
		secretId := os.Getenv("VAULT_APPROLE_SECRET_ID")

		resp, err := client.Logical().Write("auth/approle/login", map[string]interface{}{
			"role_id":   roleId,
			"secret_id": secretId,
		})
		if err != nil {
			return nil, err
		}
		token = resp.Auth.ClientToken
		client.SetToken(token)
	}

	vault := client.Logical()
	return vault, nil
}

func getVaultRoleId(appRolePath string) (string, error) {
	return "747e8ae2-f8ad-b32e-cf6a-329070efda5e", nil
}
