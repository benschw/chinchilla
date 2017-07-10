package config

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/endpoints"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
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
	consul, err := NewConsulClient()
	if err != nil {
		return nil, err
	}
	services, err := consul.Service(os.Getenv("VAULT_SERVICENAME"), "active")
	if err != nil {
		return nil, err
	}
	log.Printf("%+v", services[0].Service)
	protocol := "http"
	svc := services[0].Service
	host := fmt.Sprintf("%s://%s:%d", protocol, svc.Address, svc.Port)

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

	parts := strings.Split(appRolePath, "/")
	bucket := parts[2]
	key := parts[3]

	sess := getAwsSession()

	svc := s3.New(sess)
	input := &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	}

	result, err := svc.GetObject(input)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			log.Println(aerr.Error())
		}
		return "", err
	}
	buf := new(bytes.Buffer)
	buf.ReadFrom(result.Body)
	roleId := buf.String()

	return fmt.Sprintf("%s", roleId), nil
}

func getAwsSession() *session.Session {
	cfg := aws.NewConfig()

	// use custom resolver and path style s3 address if minio s3 server is
	// discovered with docker links on port 9000
	awsEp, usingMinioLocalS3 := os.LookupEnv("S3_PORT_9000_TCP_ADDR")
	if usingMinioLocalS3 {
		awsEnv := os.Getenv("AWS_DEFAULT_REGION")

		defaultResolver := endpoints.DefaultResolver()
		s3CustResolverFn := func(service, region string, optFns ...func(*endpoints.Options)) (endpoints.ResolvedEndpoint, error) {
			if service == "s3" {
				return endpoints.ResolvedEndpoint{
					URL:           fmt.Sprintf("http://%s:9000", awsEp),
					SigningRegion: awsEnv,
				}, nil
			}

			return defaultResolver.EndpointFor(service, region, optFns...)
		}
		cfg.
			WithEndpointResolver(endpoints.ResolverFunc(s3CustResolverFn)).
			WithS3ForcePathStyle(true)
	}

	return session.Must(session.NewSessionWithOptions(session.Options{
		Config: *cfg,
	}))
}
